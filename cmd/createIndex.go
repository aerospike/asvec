/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log/slog"
	"net"

	avs "github.com/aerospike/aerospike-proximus-client-go"
	"github.com/aerospike/aerospike-proximus-client-go/protos"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// createIndexCmd represents the createIndex command
var createIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		port := viper.GetInt("port")
		hostPort := avs.NewHostPort(host, port, false)
		namespace := viper.GetString("namespace")
		sets := viper.GetStringSlice("sets")
		indexName := viper.GetString("index-name")
		vectorField := viper.GetString("vector-field")
		dimension := viper.GetUint32("dimension")
		// distanceMetric := viper.GetInt("distance-metric")
		indexMeta := viper.GetStringMapString("index-meta")

		logger.Debug("Parsed flags", slog.String("host", host), slog.Int("port", port), slog.String("namespace", namespace), slog.Any("sets", sets), slog.String("index-name", indexName), slog.String("vector-field", vectorField), slog.Uint64("dimension", uint64(dimension)), slog.Any("index-meta", indexMeta))

		ctx := context.TODO()

		adminClient, err := avs.NewAdminClient(ctx, []*avs.HostPort{hostPort}, nil, false, logger)
		if err != nil {
			logger.Error("failed to create AVS client", slog.Any("error", err))
			view.Printf("Failed to connect to AVS: %v", err)
			return
		}

		// TODO: parse cosine
		err = adminClient.IndexCreate(ctx, namespace, sets, indexName, vectorField, dimension, protos.VectorDistanceMetric_COSINE, nil, indexMeta)
		if err != nil {
			logger.Error("unable to create index", slog.Any("error", err))
			view.Printf("Unable to create index: %v", err)
			return
		}

		view.Printf("Successfully created index %s.%s", namespace, indexName)
	},
}

func init() {
	createCmd.AddCommand(createIndexCmd)
	createIndexCmd.PersistentFlags().IPP("host", "h", net.ParseIP("127.0.0.1"), "TODO")
	createIndexCmd.PersistentFlags().IntP("port", "p", 5000, "TODO")
	createIndexCmd.Flags().StringP("namespace", "n", "", "TODO")
	createIndexCmd.Flags().StringArrayP("sets", "s", nil, "TODO")
	createIndexCmd.Flags().StringP("index-name", "i", "", "TODO")
	createIndexCmd.Flags().StringP("vector-field", "v", "vector", "TODO")
	createIndexCmd.Flags().IntP("dimension", "d", 0, "TODO")
	createIndexCmd.Flags().Uint32P("distance-metric", "m", 0, "TODO")
	createIndexCmd.Flags().StringToStringP("index-meta", "e", nil, "TODO")
	// TODO hnsw metadata

	createIndexCmd.MarkFlagRequired("namespace")
	createIndexCmd.MarkFlagRequired("set")
	createIndexCmd.MarkFlagRequired("index-name")
	// createIndexCmd.MarkFlagRequired("vector-field")
	createIndexCmd.MarkFlagRequired("dimension")
	// createIndexCmd.MarkFlagRequired("distance-metric")
	viper.BindPFlags(createIndexCmd.PersistentFlags())
	viper.BindPFlags(createIndexCmd.Flags())

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createIndexCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createIndexCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
