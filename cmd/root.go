/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	common "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var lvl = new(slog.LevelVar)
var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))
var view = NewView(os.Stdout, logger)

var rootFlags = &struct {
	logLevel LogLevelFlag
}{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "asvec",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
	rootCmd.PersistentFlags().Var(&rootFlags.logLevel, logLevelFlagName, "Log level for additional details and debugging")
	common.SetupRoot(rootCmd, "aerospike-vector-search", "0.0.0")
	viper.SetEnvPrefix("ASVEC")

	if err := viper.BindEnv(flagNameHost); err != nil {
		logger.Error("failed to bind environment variable", slog.Any("error", err))
	}

	if err := viper.BindEnv(flagNameSeeds); err != nil {
		logger.Error("failed to bind environment variable", slog.Any("error", err))
	}
}
