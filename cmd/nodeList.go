package cmd

import (
	"asvec/cmd/flags"
	"asvec/cmd/writers"
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/aerospike/avs-client-go"
	"github.com/aerospike/avs-client-go/protos"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var nodeListFlags = &struct {
	clientFlags *flags.ClientFlags
	format      int // For testing. Hidden
}{
	clientFlags: rootFlags.clientFlags,
}

func newNodeListFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.AddFlagSet(nodeListFlags.clientFlags.NewClientFlagSet())

	err := flags.AddFormatTestFlag(flagSet, &nodeListFlags.format)
	if err != nil {
		panic(err)
	}

	return flagSet
}

// nodeListCmd creates a new cobra command for listing nodes.
func newNodeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "A command for listing nodes.",
		Long: fmt.Sprintf(`A command for listing useful information about AVS nodes.

For example:

%s
asvec node ls
		`, HelpTxtSetupEnv),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return checkSeedsAndHost()
		},
		Run: func(_ *cobra.Command, _ []string) {
			logger := logger.With("cmd", "listNodeCmd")
			logger.Debug("parsed flags",
				nodeListFlags.clientFlags.NewSLogAttr()...,
			)

			client, err := createClientFromFlags(nodeListFlags.clientFlags)
			if err != nil {
				view.Error(err.Error())
				return
			}
			defer client.Close()

			ctx, cancel := context.WithTimeout(context.Background(), nodeListFlags.clientFlags.Timeout)
			defer cancel()

			nodeInfos := getAllNodesInfo(ctx, client)

			logger.Debug("received node states", slog.Any("nodeStates", nodeInfos))

			isLB := isLoadBalancer(nodeListFlags.clientFlags.Seeds)

			view.PrintNodeInfoList(nodeInfos, isLB, nodeListFlags.format)

			idsVisibleToAllNodes := getIDsVisibleToAllNodes(nodeInfos)
			idsVisibleToClient := map[uint64]struct{}{}

			for _, nodeState := range nodeInfos {
				idsVisibleToClient[nodeState.NodeID.GetId()] = struct{}{}
			}

			idsNotVisibleToClient := getNodesNotVisibleToClient(idsVisibleToAllNodes, idsVisibleToClient)

			if len(idsNotVisibleToClient) != 0 {
				if !isLB {
					// TODO handle case where only seedConn are available.
					view.Warningf(`Not all nodes are visible to asvec. 
Asvec can't reach: %s
Possible scenarios:
1. You should use --host instead of --seeds to indicate you are connection through a load balancer.
2. Asvec was able to connect to your seeds but the server(s) are returning unreachable endpoints.
   Did you forget --listener-name?
`, strings.Join(idsNotVisibleToClient, ", "))
				}
			}

			idsVisibleToEachNode := getIDsVisibleToEachNode(nodeInfos)
			nodesNotVisibleToEachNode := getNodesNotVisibleToEachNode(idsVisibleToEachNode, idsVisibleToAllNodes)

			if len(nodesNotVisibleToEachNode) != 0 {
				msg := "Not all nodes are visible to each other. The following nodes are not visible to each other:\n"

				for id, nodesNotVisible := range nodesNotVisibleToEachNode {
					msg += fmt.Sprintf("Node %d can't see: %s\n", id, strings.Join(nodesNotVisible, ", "))
				}

				view.Warning(msg)
			}
		},
	}
}

func getAllNodesInfo(ctx context.Context, client *avs.Client) []*writers.NodeInfo {
	nodeIDs := client.NodeIDs(ctx)

	logger.Debug("received node ids", slog.Any("nodeIds", nodeIDs))

	if len(nodeIDs) == 0 {
		// Loadbalancer must be enabled. Passing nil to admin client
		// causes it to fetch info from seed node
		nodeIDs = append(nodeIDs, nil)
	}

	nodeInfos := make([]*writers.NodeInfo, len(nodeIDs))
	wg := sync.WaitGroup{}

	for i, nodeID := range nodeIDs {
		wg.Add(1)

		go func(i int, nodeId *protos.NodeId) {
			defer wg.Done()

			l := logger.With("node", nodeId.String())

			if nodeId == nil {
				nodeInfos[i] = &writers.NodeInfo{NodeID: &protos.NodeId{Id: 0}}
			} else {
				nodeInfos[i] = &writers.NodeInfo{NodeID: nodeId}
			}

			wg.Add(1)

			go func() {
				defer wg.Done()

				connectedEndpoint, err := client.ConnectedNodeEndpoint(ctx, nodeId)
				if err != nil {
					l.ErrorContext(ctx,
						"failed to get connected endpoint",
						slog.Any("error", err),
					)

					view.Errorf("Failed to get connected endpoint from node %s: %s", nodeId.String(), err)

					return
				}

				l.Debug("received connected endpoint", slog.Any("connectedEndpoint", connectedEndpoint))

				nodeInfos[i].ConnectedEndpoint = connectedEndpoint
			}()

			wg.Add(1)

			go func() {
				defer wg.Done()

				endpoints, err := client.ClusterEndpoints(
					ctx,
					nodeId,
					nodeListFlags.clientFlags.ListenerName.Val, // TODO: May want to request more names.
				)
				if err != nil {
					l.ErrorContext(ctx,
						"failed to get cluster endpoints",
						slog.Any("error", err),
					)

					view.Errorf("Failed to get cluster endpoints from node %s: %s", nodeId.String(), err)

					return
				}

				l.Debug("received endpoints", slog.Any("endpoints", endpoints))

				nodeInfos[i].Endpoints = endpoints
			}()

			wg.Add(1)

			go func() {
				defer wg.Done()

				state, err := client.ClusteringState(ctx, nodeId)
				if err != nil {
					l.ErrorContext(ctx,
						"failed to get clustering state",
						slog.Any("error", err),
					)

					view.Errorf("Failed to get clustering state from node %s: %s", nodeId.String(), err)

					return
				}

				l.Debug("received clustering state", slog.Any("state", state))

				nodeInfos[i].State = state
			}()

			wg.Add(1)

			go func() {
				defer wg.Done()

				fmt.Println("nodeId ", nodeId.GetId())
				about, err := client.About(ctx, nodeId)
				if err != nil {
					l.ErrorContext(ctx,
						"failed to get about info",
						slog.Any("error", err),
					)

					view.Errorf("Failed to get about info from node %s: %s", nodeId.String(), err)

					return
				}

				l.Debug("received about info", slog.Any("about", about))

				nodeInfos[i].About = about
			}()
		}(i, nodeID)
	}

	wg.Wait()

	return nodeInfos
}

func getIDsVisibleToAllNodes(nodeInfos []*writers.NodeInfo) map[uint64]struct{} {
	idsVisibleToAllNodes := map[uint64]struct{}{}

	for _, nodeState := range nodeInfos {
		for toID := range nodeState.Endpoints.GetEndpoints() {
			idsVisibleToAllNodes[toID] = struct{}{}
		}
	}

	return idsVisibleToAllNodes
}

func getIDsVisibleToEachNode(nodeInfos []*writers.NodeInfo) map[uint64]map[uint64]struct{} {
	idsVisibleToIndividualNodes := map[uint64]map[uint64]struct{}{} // using map as set

	for _, nodeState := range nodeInfos {
		fromID := nodeState.NodeID.GetId()

		for toID := range nodeState.Endpoints.GetEndpoints() {
			if _, ok := idsVisibleToIndividualNodes[fromID]; !ok {
				idsVisibleToIndividualNodes[fromID] = map[uint64]struct{}{}
			}

			idsVisibleToIndividualNodes[fromID][toID] = struct{}{}
		}
	}

	return idsVisibleToIndividualNodes
}

func getNodesNotVisibleToClient(idsVisibleToAllNodes, idsVisibleToClient map[uint64]struct{}) []string {
	nodesNotVisibleToClient := []string{}

	for id := range idsVisibleToAllNodes {
		if _, ok := idsVisibleToClient[id]; !ok {
			nodesNotVisibleToClient = append(nodesNotVisibleToClient, strconv.FormatUint(id, 10))
		}
	}

	slices.Sort(nodesNotVisibleToClient)

	sort.Slice(nodesNotVisibleToClient, func(i, j int) bool {
		return nodesNotVisibleToClient[i] < nodesNotVisibleToClient[j]
	})

	return nodesNotVisibleToClient
}

func getNodesNotVisibleToEachNode(
	idsVisibleToEachNode map[uint64]map[uint64]struct{},
	idsVisibleToAllNodes map[uint64]struct{},
) map[uint64][]string {
	nodesNotVisibleToEachNode := map[uint64][]string{}

	for id := range idsVisibleToAllNodes {
		for fromID, visibleFromNode := range idsVisibleToEachNode {
			if _, ok := visibleFromNode[id]; !ok {
				if _, ok := nodesNotVisibleToEachNode[fromID]; !ok {
					nodesNotVisibleToEachNode[fromID] = []string{}
				}

				nodesNotVisibleToEachNode[fromID] = append(nodesNotVisibleToEachNode[fromID], strconv.FormatUint(id, 10))
			}
		}
	}

	return nodesNotVisibleToEachNode
}

func init() {
	nodeListCmd := newNodeListCmd()

	nodeCmd.AddCommand(nodeListCmd)
	nodeListCmd.Flags().AddFlagSet(newNodeListFlagSet())
}
