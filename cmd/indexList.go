package cmd

import (
	"asvec/cmd/flags"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/aerospike/avs-client-go/protos"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

//nolint:govet // Padding not a concern for a CLI
var indexListFlags = &struct {
	clientFlags flags.ClientFlags
	verbose     bool
	format      int
	yaml        bool
}{
	clientFlags: *flags.NewClientFlags(),
}

func newIndexListFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.BoolVarP(&indexListFlags.verbose, flags.Verbose, "v", false, "Print detailed index information.")                                                    //nolint:lll // For readability
	flagSet.BoolVar(&indexListFlags.yaml, flags.Yaml, false, "Output indexes in yaml format to later be used with \"asvec index create --file <index-def.yaml>") //nolint:lll // For readability
	flagSet.AddFlagSet(indexListFlags.clientFlags.NewClientFlagSet())

	err := flags.AddFormatTestFlag(flagSet, &indexListFlags.format)
	if err != nil {
		panic(err)
	}

	return flagSet
}

var indexListRequiredFlags = []string{}

// listIndexCmd represents the listIndex command
func newIndexListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "A command for listing indexes",
		Long: fmt.Sprintf(`A command for listing useful information about AVS indexes. To display additional
index information use the --%s flag.

For example:

%s
asvec index ls
		`, flags.Verbose, HelpTxtSetupEnv),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return checkSeedsAndHost()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(indexListFlags.clientFlags.NewSLogAttr(),
					slog.Bool(flags.Verbose, indexListFlags.verbose),
				)...,
			)

			adminClient, err := createClientFromFlags(&indexListFlags.clientFlags)
			if err != nil {
				return err
			}
			defer adminClient.Close()

			ctx, cancel := context.WithTimeout(context.Background(), indexListFlags.clientFlags.Timeout)
			defer cancel()

			indexList, err := adminClient.IndexList(ctx)
			if err != nil {
				logger.Error("failed to list indexes", slog.Any("error", err))
				return err
			}

			indexStatusList := make([]*protos.IndexStatusResponse, len(indexList.GetIndices()))

			cancel()

			ctx, cancel = context.WithTimeout(context.Background(), indexListFlags.clientFlags.Timeout)
			defer cancel()

			wg := sync.WaitGroup{}
			for i, index := range indexList.GetIndices() {
				wg.Add(1)
				go func(i int, index *protos.IndexDefinition) {
					defer wg.Done()
					indexStatus, err := adminClient.IndexGetStatus(ctx, index.Id.Namespace, index.Id.Name)
					if err != nil {
						logger.ErrorContext(ctx,
							"failed to get index status",
							slog.Any("error", err),
							slog.String("index", index.Id.String()),
						)
						return
					}

					indexStatusList[i] = indexStatus
					logger.Debug("server index status", slog.Int("index", i), slog.Any("response", indexStatus))
				}(i, index)
			}

			wg.Wait()

			logger.Debug("server index list", slog.String("response", indexList.String()))

			if indexListFlags.yaml {
				out, err := protojson.Marshal(indexList)
				if err != nil {
					logger.Error("failed to marshal index list", slog.Any("error", err))
					return err
				}

				logger.Debug("marshalled index list", slog.String("response", string(out)))

				var intermediate map[string]interface{}
				err = json.Unmarshal(out, &intermediate)
				if err != nil {
					logger.Error("failed to unmarshal JSON", slog.Any("error", err))
					return err
				}

				// Marshal the intermediate map to YAML
				yamlData, err := yaml.Marshal(intermediate)
				if err != nil {
					logger.Error("failed to marshal to YAML", slog.Any("error", err))
					return err
				}

				logger.Debug("marshalled index list to YAML", slog.String("response", string(yamlData)))

				view.Print(string(yamlData))
			} else {
				view.PrintIndexes(indexList, indexStatusList, indexListFlags.verbose, indexListFlags.format)

				if indexListFlags.verbose {
					view.Print("Values ending with * can be dynamically configured using the 'asvec index update' command.")
				}
			}

			return nil
		},
	}
}

func init() {
	indexListCmd := newIndexListCmd()

	indexCmd.AddCommand(indexListCmd)
	indexListCmd.Flags().AddFlagSet(newIndexListFlagSet())

	for _, flag := range indexListRequiredFlags {
		err := indexListCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
