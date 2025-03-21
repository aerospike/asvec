package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/aerospike/tools-common-go/config"
	common "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var lvl = new(slog.LevelVar)
var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))
var view = NewView(os.Stdout, os.Stderr, logger)
var Version = "development" // Overwritten at build time by ld_flags
var defaultConfigFile = "asvec.yml"

var rootFlags = &struct {
	clientFlags *flags.ClientFlags
	logLevel    flags.LogLevelFlag
	confFile    string
	clusterName string
	noColor     bool
}{
	clientFlags: flags.NewClientFlags(),
}

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
		config.SetDefaultConfName("asvec")
		config.BindPFlags(cmd.Flags(), rootFlags.clusterName)

		configFile, err := config.InitConfig(rootFlags.confFile, "", cmd.Flags())
		if err != nil {
			return err
		}

		err = config.SetFlags("", cmd.Flags())
		if err != nil {
			return err
		}

		if configFile != "" {
			logger.Info("Loading configuration parameters from file", slog.String("file", configFile))
		}

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

		flagToEnv := func(flag string) string {
			env := strings.ReplaceAll(flag, "-", "_")
			env = strings.ToUpper(env)
			env = "ASVEC_" + env
			return env
		}

		// Bind specified flags to ASVEC_*. Viper fails to set it correctly
		// because of the funky stuff we do in config.SetFlags()
		for _, flagName := range bindEnvs {
			env := flagToEnv(flagName)

			if value := os.Getenv(env); value != "" {
				err := cmd.Flags().Lookup(flagName).Value.Set(value)
				if err != nil {
					return err
				}
			}
		}

		return nil
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
	rootCmd.PersistentFlags().Var(&rootFlags.logLevel, flags.LogLevel, fmt.Sprintf("Log level for additional details and debugging. Valid values: %s", strings.Join(flags.LogLevelEnum(), ", "))) //nolint:lll // For readability
	rootCmd.PersistentFlags().BoolVar(&rootFlags.noColor, flags.NoColor, false, "Disable color in output")                                                                                        //nolint:lll // For readability
	rootCmd.PersistentFlags().StringVar(&rootFlags.confFile, flags.ConfigFile, "", fmt.Sprintf("Config file (default is %s/%s)", config.DefaultConfDir, defaultConfigFile))                       //nolint:lll // For readability
	rootCmd.PersistentFlags().StringVar(&rootFlags.clusterName, flags.ClusterName, "default", "Cluster name to use as defined in your configuration file")                                        //nolint:lll // For readability
	rootCmd.PersistentFlags().AddFlagSet(rootFlags.clientFlags.NewClientFlagSet())

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
}
