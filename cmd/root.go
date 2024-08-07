package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	common "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var lvl = new(slog.LevelVar)
var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))
var view = NewView(os.Stdout, os.Stderr, logger)
var Version = "development" // Overwritten at build time by ld_flags

var rootFlags = &struct {
	logLevel flags.LogLevelFlag
	noColor  bool
}{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "asvec",
	Short: "Aerospike Vector Search CLI",
	Long: fmt.Sprintf(`Welcome to the AVS Deployment Manager CLI Tool!
To start using this tool, please consult the detailed documentation available at https://aerospike.com/docs/vector.
Should you encounter any issues or have questions, feel free to report them by creating a GitHub issue.
Enterprise customers requiring support should contact Aerospike Support directly at https://aerospike.com/support.

For example:
%s
asvec --help
	`, HelpTxtSetupEnv),
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		if rootFlags.logLevel.NotSet() {
			lvl.Set(slog.LevelError + 1) // disable all logging
		} else {
			level := rootFlags.logLevel
			handler := logger.Handler()

			err := lvl.UnmarshalText([]byte(level))
			if err != nil {
				return err
			}

			handler.Enabled(context.Background(), lvl.Level())
		}

		if rootFlags.noColor {
			view.DisableColor()
		}

		cmd.SilenceUsage = true

		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		var persistedErr error
		flags := cmd.Flags()

		flags.VisitAll(func(f *pflag.Flag) {
			val := viper.GetString(f.Name)

			// Apply the viper config value to the flag when viper has a value
			if viper.IsSet(f.Name) && !f.Changed {
				if err := f.Value.Set(val); err != nil {
					persistedErr = fmt.Errorf("failed to parse flag %s: %s", f.Name, err)
				}
			}
		})

		return persistedErr
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		code := errCode.Load()

		if code != 0 {
			os.Exit(int(code))
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Var(
		&rootFlags.logLevel,
		flags.LogLevel,
		fmt.Sprintf("Log level for additional details and debugging. Valid values: %s", strings.Join(flags.LogLevelEnum(), ", ")), //nolint:lll // For readability
	)
	rootCmd.PersistentFlags().BoolVar(
		&rootFlags.noColor,
		flags.NoColor,
		false,
		"Disable color in output",
	)
	common.SetupRoot(rootCmd, "aerospike-vector-search", Version)

	// TODO: Add custom template for usage to take into account terminal width
	// Ex: https://github.com/sigstore/cosign/pull/3011/files
	// Below is the poor man version
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err == nil {
		usageTemplate := rootCmd.UsageTemplate()
		usageTemplate = strings.ReplaceAll(usageTemplate, ".FlagUsages", fmt.Sprintf(".FlagUsagesWrapped %d", width))
		rootCmd.SetUsageTemplate(usageTemplate)
	} else {
		logger.Debug("failed to get terminal width", slog.Any("error", err))
	}

	viper.SetEnvPrefix("ASVEC")

	bindEnvs := []string{
		flags.Host,
		flags.Seeds,
		flags.AuthUser,
		flags.AuthPassword,
		flags.AuthCredentials,
		flags.TLSCaFile,
		flags.TLSCaPath,
		flags.TLSCertFile,
		flags.TLSKeyFile,
		flags.TLSKeyFilePass,
	}

	// Bind specified flags to ASVEC_*
	for _, env := range bindEnvs {
		if err := viper.BindEnv(env); err != nil {
			panic(fmt.Sprintf("failed to bind environment variable: %s", err))
		}
	}
}
