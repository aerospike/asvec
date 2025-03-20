package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"
	"math"
	"reflect"

	"github.com/aerospike/avs-client-go"
	"github.com/aerospike/avs-client-go/protos"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//nolint:govet // Padding not a concern for a CLI
var queryFlags = &struct {
	clientFlags     *flags.ClientFlags
	namespace       string
	set             flags.StringOptionalFlag
	indexName       string
	keyString       flags.StringOptionalFlag
	keyInt          flags.IntOptionalFlag
	vector          flags.VectorFlag
	maxResults      uint32
	maxDataKeys     uint
	maxDataColWidth uint
	includeFields   []string
	hnswEf          flags.Uint32OptionalFlag
	format          int // For testing. Hidden
}{
	clientFlags: rootFlags.clientFlags,
}

const (
	defaultMaxResults             = 5
	defaultMaxDataKeys            = 5
	failedToRunVectorSearchErrMsg = "failed to run vector search"
)

func newQueryFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.StringVarP(&queryFlags.namespace, flags.Namespace, flags.NamespaceShort, "", "The namespace for the index to query.")                                                                                          //nolint:lll // For readability
	flagSet.VarP(&queryFlags.set, flags.Set, flags.SetShort, fmt.Sprintf("When a --%s query is done you may also need to provide a set so the appropriate record is retrieved.", flags.KeyString))                         //nolint:lll // For readability
	flagSet.StringVarP(&queryFlags.indexName, flags.IndexName, flags.IndexNameShort, "", "The name of the index to query.")                                                                                                //nolint:lll // For readability
	flagSet.VarP(&queryFlags.keyString, flags.KeyString, flags.KeyStrShort, "Optionally use the vector from the given string key to perform a query.")                                                                     //nolint:lll // For readability
	flagSet.VarP(&queryFlags.keyInt, flags.KeyInt, flags.KeyIntShort, "Optionally use the vector from the given integer key to perform a query.")                                                                          //nolint:lll // For readability
	flagSet.VarP(&queryFlags.vector, flags.Vector, flags.VectorShort, "The vector to use as a query. Values true/false and 1/0 will result in a binary vector. Values containing a decimal will result in a float vector") //nolint:lll // For readability
	flagSet.Uint32VarP(&queryFlags.maxResults, flags.MaxResults, "r", defaultMaxResults, "The maximum number of records to return.")                                                                                       //nolint:lll // For readability
	flagSet.UintVarP(&queryFlags.maxDataKeys, flags.MaxDataKeys, "m", defaultMaxDataKeys, "The maximum number of record data keys to display before truncating.")                                                          //nolint:lll // For readability
	flagSet.UintVarP(&queryFlags.maxDataColWidth, flags.MaxDataColWidth, flags.MaxDataColWidthShort, 50, "The maximum column width for record data before wrapping. To display long values on a single line set to 0.")    //nolint:lll // For readability
	flagSet.StringSliceVarP(&queryFlags.includeFields, flags.Fields, "f", nil, "Fields names to include when displaying record data.")                                                                                     //nolint:lll // For readability
	flagSet.Var(&queryFlags.hnswEf, flags.HnswEf, "The default number of candidate nearest neighbors shortlisted during search. Larger values provide better recall at the cost of longer search times.")                  //nolint:lll // For readability

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

// queryCmd represents the query command
func newQueryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "query",
		Short: "A command for exploring your data",
		Long: fmt.Sprintf(`A command for querying an index. A few example commands are provided
below that provide various options for browsing your vector index and different 
ways to control the amount of information displayed.

For example:

%s

# Query using the zero vector displaying only fields name,age
asvec query -i my-index -n my-namespace -f name,age

# Query 10 vectors using an existing vector
asvec query -i my-index -n my-namespace -s my-set -k my-key --max-results 10

# Query using your own float vector and change the displayed DATA column width to 100 characters.
asvec query -i my-index -n my-namespace -v "[0.5,0.1,0.3,0.4,1.0]" --max-width 100

# Query using your own bool vector and change the number of DATA rows displayed to 10.
asvec query -i my-index -n my-namespace -v "[1,0,1,0,0,0,1,0,1,1]" --max-keys 10

		`, HelpTxtSetupEnv),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if viper.IsSet(flags.Set) && !(viper.IsSet(flags.KeyString) || viper.IsSet(flags.KeyInt)) {
				view.Warningf(
					"The --%s flag is only used when the --%s or --%s flag is set.",
					flags.Set,
					flags.KeyString,
					flags.KeyInt,
				)
			}

			return checkSeedsAndHost()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(rootFlags.clientFlags.NewSLogAttr(),
					slog.String(flags.Namespace, queryFlags.namespace),
					slog.Any(flags.Set, queryFlags.set.Val),
					slog.String(flags.IndexName, queryFlags.indexName),
					slog.Any(flags.KeyString, queryFlags.keyString.Val),
					slog.Any(flags.KeyInt, queryFlags.keyInt.Val),
					slog.Any(flags.Vector, queryFlags.vector),
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

			hnswSearchParams := &protos.HnswSearchParams{
				Ef: queryFlags.hnswEf.Val,
			}

			var (
				neighbors []*avs.Neighbor
			)

			if queryFlags.vector.IsSet() {
				neighbors, err = queryVectorByVector(ctx, client, hnswSearchParams)

				if err != nil {
					logger.ErrorContext(ctx, "unable to get vector using provided vector", slog.Any("error", err))
					view.Errorf("Failed to get vector using vector: %s", err)

					return err
				}
			} else {
				indexDef, err := client.IndexGet(ctx, queryFlags.namespace, queryFlags.indexName, false)
				if err != nil {
					logger.ErrorContext(ctx, "unable to get index definition", slog.Any("error", err))
					view.Errorf("Failed to get index definition: %s", err)

					return err
				}

				if queryFlags.keyString.Val != nil || queryFlags.keyInt.Val != nil {
					neighbors, err = queryVectorByKey(ctx, client, indexDef, hnswSearchParams)
					if err != nil {
						logger.ErrorContext(ctx, "unable to get vector using provided key", slog.Any("error", err))
						view.Errorf("Failed to get vector using key: %s", err)

						return err
					}
				} else {
					neighbors, err = trialAndErrorQuery(ctx, client, int(indexDef.Dimensions), hnswSearchParams)
					if err != nil {
						logger.ErrorContext(ctx, "unable to get vector using zero vector", slog.Any("error", err))
						view.Errorf("Failed to get vector using zero vector: %s", err)

						return err
					}
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

			if queryFlags.maxDataKeys > math.MaxInt {
				err := fmt.Errorf("maxDataKeys value is larger than the maximum integer: %d", queryFlags.maxDataKeys)
				logger.ErrorContext(ctx, "unable to convert maxDataKeys to int", slog.Any("error", err))
				view.Errorf("Failed to get index definition: %s", err)

				return err
			}

			if queryFlags.maxDataColWidth > math.MaxInt {
				err := fmt.Errorf("maxDataColWidth value is larger than the maximum integer: %d", queryFlags.maxDataColWidth)
				logger.ErrorContext(ctx, "unable to convert maxDataColWidth to int", slog.Any("error", err))
				view.Errorf("Failed to get index definition: %s", err)

				return err
			}

			//nolint:gosec // Overflow is checked above
			view.PrintQueryResults(neighbors, queryFlags.format, int(queryFlags.maxDataKeys), int(queryFlags.maxDataColWidth))

			if !viper.IsSet(flags.MaxResults) {
				view.Printf("Hint: To increase the number of records returned, use the --%s flag.", flags.MaxResults)

				if !viper.IsSet(flags.Fields) {
					view.Printf(
						"Hint: To choose which record keys are displayed, use the --%s flag. By default only %d are displayed.",
						flags.Fields,
						defaultMaxDataKeys,
					)
				}
			}

			return nil
		},
	}
}

func queryVectorByVector(
	ctx context.Context,
	client *avs.Client,
	hnswSearchParams *protos.HnswSearchParams,
) ([]*avs.Neighbor, error) {
	if queryFlags.vector.FloatSlice != nil {
		return client.VectorSearchFloat32(
			ctx,
			queryFlags.namespace,
			queryFlags.indexName,
			queryFlags.vector.FloatSlice,
			queryFlags.maxResults,
			hnswSearchParams,
			queryFlags.includeFields,
			nil,
		)
	}

	return client.VectorSearchBool(
		ctx,
		queryFlags.namespace,
		queryFlags.indexName,
		queryFlags.vector.BoolSlice,
		queryFlags.maxResults,
		hnswSearchParams,
		queryFlags.includeFields,
		nil,
	)
}

func queryVectorByKey(
	ctx context.Context,
	client *avs.Client,
	indexDef *protos.IndexDefinition,
	hnswSearchParams *protos.HnswSearchParams,
) ([]*avs.Neighbor, error) {
	logger := logger.With(
		slog.String("key-str", queryFlags.keyString.String()),
		slog.String("key-int", queryFlags.keyInt.String()),
		slog.String("index", queryFlags.indexName),
		slog.String("namespace", queryFlags.namespace),
		slog.String("field", indexDef.Field),
	)

	set := queryFlags.set.Val
	if set == nil {
		// If the user did not specify a set try to get the set from the index.
		// An index does not need a set defined when created to index a record
		// in a set. However, a get request does not work the same way.
		set = indexDef.SetFilter
	}

	var key any

	switch {
	case queryFlags.keyString.Val != nil:
		key = *queryFlags.keyString.Val
	case queryFlags.keyInt.Val != nil:
		key = *queryFlags.keyInt.Val
	default:
		logger.ErrorContext(ctx, "no key provided")
		return nil, fmt.Errorf("no key provided, this should not happen")
	}

	record, err := client.Get(ctx, queryFlags.namespace, set, key, []string{indexDef.Field}, nil)
	if err != nil {
		msg := "unable to get record"
		logger.ErrorContext(ctx, msg, slog.Any("error", err))

		if set == nil {
			view.Warningf(
				"The requested record was not found. If the record is in a set, you may also need to provide the --%s flag.",
				flags.Set,
			)
		}

		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	logger.DebugContext(ctx, "queried record", slog.Any("record", record))

	queryVector, ok := record.Data[indexDef.Field]
	if !ok {
		msg := "field not found in specified record"
		logger.ErrorContext(ctx, msg, slog.String("field", indexDef.Field), slog.Any("data", record.Data))

		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	var neighbors []*avs.Neighbor

	switch v := queryVector.(type) {
	case []float32:
		neighbors, err = client.VectorSearchFloat32(
			ctx,
			queryFlags.namespace,
			queryFlags.indexName,
			v,
			queryFlags.maxResults+1, // we will remove queried vector from results
			hnswSearchParams,
			queryFlags.includeFields,
			nil,
		)

	case []bool:
		neighbors, err = client.VectorSearchBool(
			ctx,
			queryFlags.namespace,
			queryFlags.indexName,
			v,
			queryFlags.maxResults+1, // we will remove queried vector from results
			hnswSearchParams,
			queryFlags.includeFields,
			nil,
		)
	}

	if err != nil {
		logger.ErrorContext(ctx, failedToRunVectorSearchErrMsg, slog.Any("error", err))
		view.Errorf("Unable to run vector query: %s", err)

		return nil, err
	}

	// Remove the queried vector from the results
	newNeighbors := make([]*avs.Neighbor, 0, len(neighbors)-1)

	for _, n := range neighbors {
		if !reflect.DeepEqual(n.Key, key) {
			newNeighbors = append(newNeighbors, n)
		}
	}

	neighbors = newNeighbors

	return neighbors, nil
}

func trialAndErrorQuery(
	ctx context.Context,
	client *avs.Client,
	dimension int,
	hnswSearchParams *protos.HnswSearchParams,
) ([]*avs.Neighbor, error) {
	logger := logger.With(
		slog.String("index", queryFlags.indexName),
		slog.String("namespace", queryFlags.namespace),
		slog.Int("dimension", dimension),
	)

	queryFloat32 := make([]float32, dimension)
	neighbors, err := client.VectorSearchFloat32(
		ctx,
		queryFlags.namespace,
		queryFlags.indexName,
		queryFloat32,
		queryFlags.maxResults,
		hnswSearchParams,
		queryFlags.includeFields,
		nil,
	)

	if err != nil {
		logger.WarnContext(ctx, failedToRunVectorSearchErrMsg, slog.Any("error", err))
	}

	if err != nil || len(neighbors) == 0 {
		queryBool := make([]bool, dimension)
		neighbors, err = client.VectorSearchBool(
			ctx,
			queryFlags.namespace,
			queryFlags.indexName,
			queryBool,
			queryFlags.maxResults,
			hnswSearchParams,
			queryFlags.includeFields,
			nil,
		)

		if err != nil {
			logger.ErrorContext(ctx, failedToRunVectorSearchErrMsg, slog.Any("error", err))
			view.Errorf("Unable to run vector query: %s", err)

			return nil, err
		}
	}

	return neighbors, nil
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

	queryCmd.MarkFlagsMutuallyExclusive(flags.Vector, flags.KeyString, flags.KeyInt)

	// Add watch functionality to the query command
	wrapCommandWithWatch(queryCmd)
}
