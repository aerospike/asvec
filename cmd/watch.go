package cmd

import (
	"asvec/cmd/flags"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// WatchFlags contains the flags for watch functionality
type WatchFlags struct {
	Watch         bool
	WatchInterval int
}

// NewWatchFlags creates a new WatchFlags instance with default values
func NewWatchFlags() *WatchFlags {
	return &WatchFlags{
		Watch:         false,
		WatchInterval: flags.DefaultWatchInterval,
	}
}

// AddWatchFlagSet adds watch flags to the provided flag set
func AddWatchFlagSet(flagSet *pflag.FlagSet, watchFlags *WatchFlags) {
	flagSet.BoolVar(&watchFlags.Watch, flags.Watch,
		false, "Watch mode: continuously rerun the command at a set interval")
	flagSet.IntVar(&watchFlags.WatchInterval, flags.WatchInterval,
		flags.DefaultWatchInterval, "Interval in seconds at which the watched command is rerun")
}

// LineCountingWriter is a writer that counts the number of lines written
type LineCountingWriter struct {
	Writer    io.Writer
	LineCount int
}

// Write implements the io.Writer interface
func (w *LineCountingWriter) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	w.LineCount += bytes.Count(p, []byte{'\n'})

	return n, err
}

// RunWithWatch wraps a command's RunE function with watch functionality
func RunWithWatch(cmd *cobra.Command, args []string, watchFlags *WatchFlags,
	runFunc func(cmd *cobra.Command, args []string) error) error {
	if !watchFlags.Watch {
		// If watch mode is not enabled, just run the command once
		return runFunc(cmd, args)
	}

	logger.Debug("watch mode active",
		slog.Int("interval", watchFlags.WatchInterval),
		slog.String("command", cmd.CommandPath()))

	// Set up signal handling for clean exit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	// Get the full command with all flags
	cmdArgs := os.Args

	// Remove the watch flags from the command to avoid recursion
	cmdArgs = removeWatchFlags(cmdArgs)

	ticker := time.NewTicker(time.Duration(watchFlags.WatchInterval) * time.Second)
	defer ticker.Stop()

	// Create a line counting writer to track output lines
	lineCounter := &LineCountingWriter{Writer: view.out}

	// Save the original stdout and stderr
	originalOut := view.out
	originalErr := view.err

	// Count how many header lines we're printing
	headerLineCount := 0

	// Helper function to print the header with updated timestamp
	printHeader := func() {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		headerText := fmt.Sprintf("Watch mode: refresh every %d seconds (press Ctrl+C to exit) - Last update: %s",
			watchFlags.WatchInterval, timestamp)
		view.Print(headerText)
		view.Printf("> %s", strings.Join(cmdArgs, " "))
		view.Print("") // Add a blank line after the command for better readability
	}

	// Print the header for the first time
	printHeader()

	headerLineCount = 3 // Three lines: header text, command, and blank line

	// Log watch mode information
	logger.Info("running command in watch mode",
		slog.Int("refresh_interval_seconds", watchFlags.WatchInterval),
		slog.String("command", strings.Join(cmdArgs, " ")),
	)

	// Helper function to run the command and handle output
	runCommandAndCaptureOutput := func() (int, error) {
		// Set the line counter as the output writer
		lineCounter.LineCount = 0
		view.out = lineCounter

		// Run the command
		err := runFunc(cmd, args)

		// Store the number of lines
		lineCount := lineCounter.LineCount

		// Restore original stdout/stderr
		view.out = originalOut
		view.err = originalErr

		return lineCount, err
	}

	// Run the command for the first time
	previousLineCount, err := runCommandAndCaptureOutput()
	if err != nil {
		return err
	}

	// Then run it on each tick
	for {
		select {
		case <-ctx.Done():
			logger.Debug("Watch mode terminated by user")
			return nil
		case <-ticker.C:
			// Move cursor up to the beginning of the previous output (including the header)
			totalLinesToClear := previousLineCount + headerLineCount
			for i := 0; i < totalLinesToClear; i++ {
				fmt.Fprint(view.out, "\033[1A") // Move cursor up one line
				fmt.Fprint(view.out, "\033[2K") // Clear the entire line
			}

			// Log refresh information
			logger.Debug("Refreshing command output",
				slog.Int("refresh_interval_seconds", watchFlags.WatchInterval),
				slog.String("command", strings.Join(cmdArgs, " ")),
			)

			// Print the header with updated timestamp
			printHeader()

			// Run the command again
			var err error

			previousLineCount, err = runCommandAndCaptureOutput()
			if err != nil {
				logger.Error("Error executing command in watch mode", slog.Any("error", err))
				return err
			}
		}
	}
}

// removeWatchFlags removes watch-related flags from the command arguments
func removeWatchFlags(args []string) []string {
	result := make([]string, 0, len(args))
	skip := false

	for i, arg := range args {
		// Skip the current arg if the previous iteration marked it for skipping
		if skip {
			skip = false
			continue
		}

		// Check for watch flags
		if arg == fmt.Sprintf("--%s", flags.Watch) {
			continue
		}

		// Check for watch interval flag with value as separate argument
		if arg == fmt.Sprintf("--%s", flags.WatchInterval) && i+1 < len(args) {
			skip = true // Skip the next arg (the value)
			continue
		}

		// Check for watch interval flag with value in same argument
		if strings.HasPrefix(arg, fmt.Sprintf("--%s=", flags.WatchInterval)) {
			continue
		}

		result = append(result, arg)
	}

	return result
}

// wrapCommandWithWatch adds watch functionality to a command
// This function should be called in the init() function of commands that want to support watch mode
func wrapCommandWithWatch(cmd *cobra.Command) {
	// Add watch flags to the command
	watchFlags := NewWatchFlags()
	AddWatchFlagSet(cmd.Flags(), watchFlags)

	// Save the original RunE function
	originalRunE := cmd.RunE
	originalRun := cmd.Run

	// Replace the Run/RunE function with one that uses watch
	switch {
	case originalRunE != nil:
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			// Extract watch flags from command
			watchFlags.Watch, _ = cmd.Flags().GetBool(flags.Watch)
			watchFlags.WatchInterval, _ = cmd.Flags().GetInt(flags.WatchInterval)

			return RunWithWatch(cmd, args, watchFlags, originalRunE)
		}
		cmd.Run = nil
	case originalRun != nil:
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			// Extract watch flags from command
			watchFlags.Watch, _ = cmd.Flags().GetBool(flags.Watch)
			watchFlags.WatchInterval, _ = cmd.Flags().GetInt(flags.WatchInterval)

			return RunWithWatch(cmd, args, watchFlags, func(cmd *cobra.Command, args []string) error {
				originalRun(cmd, args)
				return nil
			})
		}
		cmd.Run = nil
	default:
		panic("No RunE or Run function found for command")
	}

	logger.Debug("Added watch functionality to command", slog.String("command", cmd.CommandPath()))
}
