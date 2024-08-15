package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

//nolint:govet // Padding not a concern for a CLI
var indexGCFlags = &struct {
	clientFlags flags.ClientFlags
	namespace   string
	indexName   string
	cutoffTime  flags.UnixTimestampFlag
}{
	clientFlags: *flags.NewClientFlags(),
}

func newIndexGCFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.StringVarP(&indexGCFlags.namespace, flags.Namespace, "n", "", "The namespace for the index.") //nolint:lll // For readability
	flagSet.StringVarP(&indexGCFlags.indexName, flags.IndexName, "i", "", "The name of the index.")       //nolint:lll // For readability
	flagSet.VarP(&indexGCFlags.cutoffTime, flags.CutoffTime, "c", "The cutoff time for gc.")              //nolint:lll // For readability
	flagSet.AddFlagSet(indexGCFlags.clientFlags.NewClientFlagSet())

	return flagSet
}

var indexGCRequiredFlags = []string{
	flags.Namespace,
	flags.IndexName,
	flags.CutoffTime,
}

// gcIndexCmd represents the gcIndex command
func newIndexGCCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gc",
		Short: "A command for proactively garbage collecting indexes",
		Long: fmt.Sprintf(`A command for proactively garbage collecting indexes.
Vertices identified as invalid before --%s (Unix timestamp) are garbage collected.
For guidance on managing your indexes, refer to: 
https://aerospike.com/docs/vector/operate/index-management"

For example:

%s
asvec index gc -i myindex -n test -c 1720744696
			`, flags.CutoffTime, HelpTxtSetupEnv),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return checkSeedsAndHost()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			debugFlags := indexGCFlags.clientFlags.NewSLogAttr()
			logger.Debug("parsed flags",
				append(debugFlags,
					slog.String(flags.Namespace, indexGCFlags.namespace),
					slog.String(flags.IndexName, indexGCFlags.indexName),
					slog.Time(flags.CutoffTime, indexGCFlags.cutoffTime.Time()),
				)...,
			)

			client, err := createClientFromFlags(&indexGCFlags.clientFlags)
			if err != nil {
				return err
			}
			defer client.Close()

			ctx, cancel := context.WithTimeout(context.Background(), indexGCFlags.clientFlags.Timeout)
			defer cancel()

			err = client.GcInvalidVertices(
				ctx,
				indexGCFlags.namespace,
				indexGCFlags.indexName,
				indexGCFlags.cutoffTime.Time(),
			)
			if err != nil {
				logger.Error("unable to garbage collect index", slog.Any("error", err))
				return err
			}

			view.Printf(
				"Successfully started garbage collection for index %s.%s",
				indexGCFlags.namespace,
				indexGCFlags.indexName,
			)

			return nil
		},
	}
}

func init() {
	gcIndexCmd := newIndexGCCmd()
	indexCmd.AddCommand(gcIndexCmd)

	// TODO: Add custom template for usage to take into account terminal width
	// Ex: https://github.com/sigstore/cosign/pull/3011/files

	flagSet := newIndexGCFlagSet()
	gcIndexCmd.Flags().AddFlagSet(flagSet)

	for _, flag := range indexGCRequiredFlags {
		err := gcIndexCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
