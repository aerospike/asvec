package cmd

import (
	"asvec/cmd/flags"
	"asvec/cmd/writers"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/aerospike/avs-client-go/protos"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var nodeListFlags = &struct {
	clientFlags flags.ClientFlags
}{
	clientFlags: *flags.NewClientFlags(),
}

func newNodeListFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.AddFlagSet(nodeListFlags.clientFlags.NewClientFlagSet())

	return flagSet
}

var nodeListRequiredFlags = []string{}

// listClusterCmd represents the listCluster command
func newNodeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "A command for listing nodes",
		Long: fmt.Sprintf(`A command for listing useful information about AVS nodes.

For example:

%s
asvec node ls
		`, HelpTxtSetupEnv),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return checkSeedsAndHost()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger := logger.With("cmd", "listClusterCmd")
			logger.Debug("parsed flags",
				nodeListFlags.clientFlags.NewSLogAttr()...,
			)

			adminClient, err := createClientFromFlags(&nodeListFlags.clientFlags)
			if err != nil {
				return err
			}
			defer adminClient.Close()

			ctx, cancel := context.WithTimeout(context.Background(), nodeListFlags.clientFlags.Timeout)
			defer cancel()

			nodeIds := adminClient.NodeIds(ctx)

			if len(nodeIds) == 0 {
				// Loadbalancer must be enabled. Passing nil to admin client
				// causes it to fetch info from seed node
				nodeIds = append(nodeIds, nil)

				// TODO add warnings about not seeing nodes
			}

			logger.Debug("received node ids", slog.Any("nodeIds", nodeIds))

			nodeStates := make([]*writers.NodeClusterInfo, len(nodeIds))

			for i, nodeId := range nodeIds {
				l := logger.With("node", nodeId.String())

				if nodeId == nil {
					nodeStates[i] = &writers.NodeClusterInfo{NodeId: &protos.NodeId{Id: 0}}
				} else {
					nodeStates[i] = &writers.NodeClusterInfo{NodeId: nodeId}
				}

				endpoints, err := adminClient.ClusterEndpoints(ctx, nodeId, nil) // TODO add option listener name
				if err != nil {
					l.ErrorContext(ctx,
						"failed to get cluster endpoints",
						slog.Any("error", err),
					)

					view.Printf("Failed to get cluster endpoints from node %s: %s", nodeId.String(), err)
					continue
				}

				l.Debug("received endpoints", slog.Any("endpoints", endpoints))

				nodeStates[i].Endpoints = endpoints

				state, err := adminClient.ClusteringState(ctx, nodeId)
				if err != nil {
					l.ErrorContext(ctx,
						"failed to get clustering state",
						slog.Any("error", err),
					)

					view.Printf("Failed to get clustering state from node %s: %s", nodeId.String(), err)
					continue
				}

				l.Debug("received clustering state", slog.Any("state", state))

				nodeStates[i].State = state

				about, err := adminClient.About(ctx, nodeId)
				if err != nil {
					l.ErrorContext(ctx,
						"failed to get about info",
						slog.Any("error", err),
					)

					view.Printf("Failed to get about info from node %s: %s", nodeId.String(), err)
					continue
				}

				l.Debug("received about info", slog.Any("about", about))

				nodeStates[i].About = about
			}

			idsVisibleToIndividualNodes := map[uint64]map[uint64]struct{}{} // using map as set
			idsVisibleToAllNodes := map[uint64]struct{}{}
			idsVisibleToClient := map[uint64]struct{}{}
			idsToEndpoints := map[uint64]map[string]struct{}{} // using map as set
			// endpointsVisisbleToclient := map[string]struct{}{}

			for _, nodeState := range nodeStates {
				fromId := nodeState.NodeId.GetId()
				idsVisibleToClient[fromId] = struct{}{}

				for toId, endpoint := range nodeState.Endpoints.GetEndpoints() {
					if _, ok := idsVisibleToIndividualNodes[fromId]; !ok {
						idsVisibleToIndividualNodes[fromId] = map[uint64]struct{}{}
					}

					if _, ok := idsToEndpoints[toId]; !ok {
						idsToEndpoints[toId] = map[string]struct{}{}
					}

					idsVisibleToAllNodes[toId] = struct{}{}

					for _, e := range endpoint.GetEndpoints() {
						endpointStr := fmt.Sprintf("%s:%d", e.GetAddress(), e.GetPort())
						idsToEndpoints[toId][endpointStr] = struct{}{}
						idsVisibleToIndividualNodes[fromId][toId] = struct{}{}
					}
				}
			}

			// TODO: Handle case where nodes can't see eachother

			fmt.Println("Visible Nodes")
			for fromId, toIds := range idsVisibleToAllNodes {
				fmt.Printf("Node %d sees nodes: %v\n", fromId, toIds)
			}

			fmt.Println("Visible Endpoints")
			for fromId, endpoints := range idsVisibleToIndividualNodes {
				fmt.Printf("Node %d sees endpoints: %v\n", fromId, endpoints)
			}

			fmt.Println("Visible to Client")
			for id := range idsVisibleToClient {
				fmt.Printf("Node %d is visible to client\n", id)
			}

			logger.Debug("received node states", slog.Any("nodeStates", nodeStates))

			view.PrintNodesClusterState(nodeStates)

			isLB := isLoadBalancer(nodeListFlags.clientFlags.Seeds, nodeListFlags.clientFlags.Host)

			if len(idsVisibleToClient) < len(idsVisibleToAllNodes) {
				if !isLB {
					nodesNotVisibleToClient := []string{}

					for id, _ := range idsVisibleToAllNodes {
						if _, ok := idsVisibleToClient[id]; !ok {
							nodesNotVisibleToClient = append(nodesNotVisibleToClient, strconv.FormatUint(id, 10))
						}
					}

					fmt.Printf(`Warning: Not all nodes are visible to asvec. 
Asvec can't reach: %s
Possible scenarios:	
1. You should use --host instead of --seeds to indicate you are connection through a load balancer.
2. Asvec was able to connect to your seeds but you either forgot to provide --advertised-listener or the server is returning unreachable endpoints.
`, strings.Join(nodesNotVisibleToClient, ", "))
				}
			}

			visibilityErr := false
			nodesCantSee := map[uint64][]string{}
			for id, _ := range idsVisibleToAllNodes {
				for fromId, visibleFromNode := range idsVisibleToIndividualNodes {
					if _, ok := visibleFromNode[id]; !ok {
						if _, ok := nodesCantSee[fromId]; !ok {
							nodesCantSee[fromId] = []string{}
						}

						visibilityErr = true
						nodesCantSee[fromId] = append(nodesCantSee[fromId], strconv.FormatUint(id, 10))
					}
				}
			}

			if visibilityErr {
				view.Print("Cluster formation error:")
				for fromId, cantSee := range nodesCantSee {
					view.Printf("\nNode %d can't see: %s", fromId, strings.Join(cantSee, ", "))
				}
			}

			return nil
		},
	}
}

func init() {
	nodeListCmd := newNodeListCmd()

	clusterCmd.AddCommand(nodeListCmd)
	nodeListCmd.Flags().AddFlagSet(newNodeListFlagSet())

	for _, flag := range nodeListRequiredFlags {
		err := nodeListCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
