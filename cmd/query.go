package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"

	"github.com/aerospike/avs-client-go/protos"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

//nolint:govet // Padding not a concern for a CLI
var queryFlags = &struct {
	clientFlags   *flags.ClientFlags
	namespace     string
	indexName     string
	maxResults    uint32
	maxDataKeys   uint
	includeFields []string
	hnswEf        flags.Uint32OptionalFlag
	format        int // For testing. Hidden
}{
	clientFlags: rootFlags.clientFlags,
}

const (
	defaultMaxResults  = 5
	defaultMaxDataKeys = 5
)

func newQueryFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.StringVarP(&queryFlags.namespace, flags.Namespace, flags.NamespaceShort, "", "The namespace for the index to query.")                                                                         //nolint:lll // For readability
	flagSet.StringVarP(&queryFlags.indexName, flags.IndexName, flags.IndexNameShort, "", "The name of the index to query.")                                                                               //nolint:lll // For readability
	flagSet.Uint32VarP(&queryFlags.maxResults, flags.MaxResults, "r", defaultMaxDataKeys, "The maximum number of records to return.")                                                                     //nolint:lll // For readability
	flagSet.UintVarP(&queryFlags.maxDataKeys, flags.MaxDataKeys, "m", defaultMaxDataKeys, "The maximum number of records to return.")                                                                     //nolint:lll // For readability
	flagSet.StringSliceVarP(&queryFlags.includeFields, flags.Fields, "f", nil, "Fields names in include when displaying record data.")                                                                    //nolint:lll // For readability
	flagSet.Var(&queryFlags.hnswEf, flags.HnswEf, "The default number of candidate nearest neighbors shortlisted during search. Larger values provide better recall at the cost of longer search times.") //nolint:lll // For readability

	err := flags.AddFormatTestFlag(flagSet, &queryFlags.format)
	if err != nil {
		panic(err)
	}

	return flagSet
}

var queryRequiredFlags = []string{
	flags.Namespace,
	flags.IndexName,
}

// listIndexCmd represents the listIndex command
func newQueryCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "query",
		Aliases: []string{"list"},
		Short:   "A command for exploring your data",
		Long: fmt.Sprintf(`A command for querying an index using a zero vector. 
This command is primarily intended for verifying and displaying the structure of 
your data, rather than providing robust query functionality. Several flags are 
available to adjust the amount of information displayed. To control which fields 
from a record are shown, use the --%s flag.

For example:

%s
asvec query
		`, flags.Fields, HelpTxtSetupEnv),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return checkSeedsAndHost()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(rootFlags.clientFlags.NewSLogAttr(),
					slog.String(flags.Namespace, queryFlags.namespace),
					slog.String(flags.IndexName, queryFlags.indexName),
					slog.Any(flags.MaxResults, queryFlags.maxResults),
					slog.Any(flags.MaxDataKeys, queryFlags.maxDataKeys),
					slog.Any(flags.Fields, queryFlags.includeFields),
				)...,
			)

			client, err := createClientFromFlags(rootFlags.clientFlags)
			if err != nil {
				return err
			}
			defer client.Close()

			ctx, cancel := context.WithTimeout(context.Background(), rootFlags.clientFlags.Timeout)
			defer cancel()

			index, err := client.IndexGet(ctx, queryFlags.namespace, queryFlags.indexName)
			if err != nil {
				msg := "unable to get index definition"
				logger.ErrorContext(ctx, msg, slog.Any("error", err))
				return err
			}

			hnswSearchParams := &protos.HnswSearchParams{
				Ef: queryFlags.hnswEf.Val,
			}

			queryFloat32 := make([]float32, index.Dimensions)
			neighbors, err := client.VectorSearchFloat32(
				ctx,
				index.Id.Namespace,
				index.Id.Name,
				queryFloat32,
				queryFlags.maxResults,
				hnswSearchParams,
				queryFlags.includeFields,
				nil,
			)
			if err != nil {
				msg := "failed to run vector search"
				logger.WarnContext(ctx, msg, slog.Any("error", err))
			}

			if err != nil || len(neighbors) == 0 {
				queryBool := make([]bool, index.Dimensions)
				neighbors, err = client.VectorSearchBool(
					ctx,
					index.Id.Namespace,
					index.Id.Name,
					queryBool,
					queryFlags.maxResults,
					hnswSearchParams,
					queryFlags.includeFields,
					nil,
				)

				if err != nil {
					msg := "failed to run vector search"
					logger.WarnContext(ctx, msg, slog.Any("error", err))
					view.Errorf("Unable to run vector query: %s", err)
					return err
				}
			}

			logger.DebugContext(ctx, "server vector search", slog.Any("response", neighbors))

			if len(neighbors) == 0 {
				view.Warning("Query returned zero results.")
				return nil
			}

			if queryFlags.includeFields != nil {
				// If the user has specified fields to include, we should not limit
				queryFlags.maxDataKeys = 0
			}

			view.PrintQueryResults(neighbors, queryFlags.format, int(queryFlags.maxDataKeys))

			if queryFlags.maxResults == defaultMaxResults {
				view.Printf("To increase the number of records returned, use the --%s flag.", flags.MaxResults)
			}

			if queryFlags.maxDataKeys == defaultMaxDataKeys && queryFlags.includeFields == nil {
				view.Printf("To choose which record keys are displayed, use the --%s flag. By default only %d are displayed.", flags.Fields, defaultMaxDataKeys)
			}

			return nil
		},
	}
}

func init() {
	queryCmd := newQueryCmd()

	rootCmd.AddCommand(queryCmd)
	queryCmd.Flags().AddFlagSet(newQueryFlagSet())

	for _, flag := range queryRequiredFlags {
		err := queryCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
