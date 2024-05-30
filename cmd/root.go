/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log/slog"
	"os"

	common "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var lvl = new(slog.LevelVar)
var logLevelFlagName = "log-level"
var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))
var view = NewView(os.Stdout, logger)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "asvec",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true
		viper.BindPFlags(cmd.PersistentFlags())
		viper.BindPFlags(cmd.Flags())

		if viper.IsSet(logLevelFlagName) {
			level := viper.GetString(logLevelFlagName)
			handler := logger.Handler()
			lvl.UnmarshalText([]byte(level))

			handler.Enabled(context.Background(), lvl.Level())
		} else {
			lvl.Set(slog.LevelError + 1) // disable all logging
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
	logLevel := LogLevelFlag("disabled")
	rootCmd.PersistentFlags().Var(&logLevel, logLevelFlagName, "Log level for additional details and debugging")
	common.SetupRoot(rootCmd, "aerospike-vector-search", "0.0.0")
	viper.SetEnvPrefix("ASVEC")
	viper.AutomaticEnv()
}
