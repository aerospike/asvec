package cmd

import (
	"testing"

	"asvec/cmd/flags"
	"bytes"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRemoveWatchFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "No watch flags",
			args:     []string{"asvec", "command", "--flag", "value"},
			expected: []string{"asvec", "command", "--flag", "value"},
		},
		{
			name:     "With watch flag",
			args:     []string{"asvec", "command", "--watch", "--flag", "value"},
			expected: []string{"asvec", "command", "--flag", "value"},
		},
		{
			name:     "With watch interval flag separate value",
			args:     []string{"asvec", "command", "--watch-interval", "5", "--flag", "value"},
			expected: []string{"asvec", "command", "--flag", "value"},
		},
		{
			name:     "With watch interval flag combined value",
			args:     []string{"asvec", "command", "--watch-interval=5", "--flag", "value"},
			expected: []string{"asvec", "command", "--flag", "value"},
		},
		{
			name:     "With multiple watch flags",
			args:     []string{"asvec", "command", "--watch", "--watch-interval", "5", "--flag", "value"},
			expected: []string{"asvec", "command", "--flag", "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeWatchFlags(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRunWithWatch(t *testing.T) {
	// Create a test command
	testCmd := &cobra.Command{
		Use: "test",
	}

	// Setup test cases
	tests := []struct {
		name        string
		watchFlags  *WatchFlags
		runCount    int
		shouldError bool
		continuous  bool // Whether this is a continuous watch test that should run multiple times
	}{
		{
			name: "Without watch",
			watchFlags: &WatchFlags{
				Watch:         false,
				WatchInterval: flags.DefaultWatchInterval,
			},
			runCount:    1,
			shouldError: false,
			continuous:  false,
		},
		{
			name: "With watch but error",
			watchFlags: &WatchFlags{
				Watch:         true,
				WatchInterval: flags.DefaultWatchInterval,
			},
			runCount:    1,
			shouldError: true,
			continuous:  false,
		},
		{
			name: "With watch continuous execution",
			watchFlags: &WatchFlags{
				Watch:         true,
				WatchInterval: 1, // Use shorter interval for faster tests
			},
			runCount:    2, // Should run at least twice
			shouldError: false,
			continuous:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			runCount := 0
			runFunc := func(cmd *cobra.Command, args []string) error {
				runCount++

				if tt.shouldError {
					return assert.AnError
				}

				// For continuous tests, exit after reaching the expected run count
				if tt.continuous && runCount >= tt.runCount {
					// Send interrupt signal to exit the watch loop
					go func() {
						// Give a little time for the current run to complete
						time.Sleep(100 * time.Millisecond)
						p, _ := os.FindProcess(os.Getpid())
						p.Signal(os.Interrupt)
					}()
				}

				return nil
			}

			// Create a buffer to capture output
			outBuf := new(bytes.Buffer)
			errBuf := new(bytes.Buffer)

			// Save original view outputs
			originalOut := view.out
			originalErr := view.err

			// Set test buffers
			view.out = outBuf
			view.err = errBuf

			// Run the test
			err := RunWithWatch(testCmd, []string{}, tt.watchFlags, runFunc)

			// Restore original outputs
			view.out = originalOut
			view.err = originalErr

			// Verify results
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify run count
			if tt.continuous {
				assert.GreaterOrEqual(t, runCount, tt.runCount, "Command should be executed at least %d times", tt.runCount)
			} else {
				assert.Equal(t, tt.runCount, runCount)
			}

			// For watch mode tests, verify output contains expected text
			if tt.watchFlags.Watch {
				assert.Contains(t, outBuf.String(), "Watch mode: refresh every")
				// Verify the command line is present
				assert.Contains(t, outBuf.String(), "> ")
			}
		})
	}
}

// TestWrapCommandWithWatch tests the WrapCommandWithWatch function
func TestWrapCommandWithWatch(t *testing.T) {
	// Create a test command with Run function
	runCalled := false
	cmdWithRun := &cobra.Command{
		Use: "test-run",
		Run: func(cmd *cobra.Command, args []string) {
			runCalled = true
		},
	}

	// Create a test command with RunE function
	runECalled := false
	cmdWithRunE := &cobra.Command{
		Use: "test-rune",
		RunE: func(cmd *cobra.Command, args []string) error {
			runECalled = true
			return nil
		},
	}

	// Test wrapping a command with Run function
	t.Run("Wrap command with Run function", func(t *testing.T) {
		// Apply watch functionality
		wrapCommandWithWatch(cmdWithRun)

		// Verify that watch flags were added
		watchFlag := cmdWithRun.Flags().Lookup(flags.Watch)
		assert.NotNil(t, watchFlag, "Watch flag should be added to the command")

		watchIntervalFlag := cmdWithRun.Flags().Lookup(flags.WatchInterval)
		assert.NotNil(t, watchIntervalFlag, "Watch interval flag should be added to the command")

		// Verify that Run was replaced with RunE
		assert.Nil(t, cmdWithRun.Run, "Run should be replaced with RunE")
		assert.NotNil(t, cmdWithRun.RunE, "RunE should be set")

		// Execute the command without watch
		err := cmdWithRun.Execute()
		assert.NoError(t, err)
		assert.True(t, runCalled, "Original Run function should be called")
	})

	// Test wrapping a command with RunE function
	t.Run("Wrap command with RunE function", func(t *testing.T) {
		// Apply watch functionality
		wrapCommandWithWatch(cmdWithRunE)

		// Verify that watch flags were added
		watchFlag := cmdWithRunE.Flags().Lookup(flags.Watch)
		assert.NotNil(t, watchFlag, "Watch flag should be added to the command")

		watchIntervalFlag := cmdWithRunE.Flags().Lookup(flags.WatchInterval)
		assert.NotNil(t, watchIntervalFlag, "Watch interval flag should be added to the command")

		// Execute the command without watch
		err := cmdWithRunE.Execute()
		assert.NoError(t, err)
		assert.True(t, runECalled, "Original RunE function should be called")
	})
}
