package cmd

import (
	"asvec/cmd/flags"
	"asvec/cmd/writers"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/aerospike/avs-client-go"
	"github.com/aerospike/avs-client-go/protos"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var nodeListFlags = &struct {
	clientFlags flags.ClientFlags
	format      int
}{
	clientFlags: *flags.NewClientFlags(),
}

func newNodeListFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.AddFlagSet(nodeListFlags.clientFlags.NewClientFlagSet())
	flags.AddFormatTestFlag(flagSet, &nodeListFlags.format)

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
		Run: func(_ *cobra.Command, _ []string) {
			logger := logger.With("cmd", "listClusterCmd")
			logger.Debug("parsed flags",
				nodeListFlags.clientFlags.NewSLogAttr()...,
			)

			adminClient, err := createClientFromFlags(&nodeListFlags.clientFlags)
			if err != nil {
				view.Error(err.Error())
				return
			}
			defer adminClient.Close()

			ctx, cancel := context.WithTimeout(context.Background(), nodeListFlags.clientFlags.Timeout)
			defer cancel()

			nodeInfos := getAllNodesInfo(ctx, adminClient)

			logger.Debug("received node states", slog.Any("nodeStates", nodeInfos))

			isLB := isLoadBalancer(nodeListFlags.clientFlags.Seeds, nodeListFlags.clientFlags.Host)

			view.PrintNodeInfoList(nodeInfos, isLB, nodeListFlags.format)

			idsVisibleToAllNodes := getIDsVisibleToAllNodes(nodeInfos)
			idsVisibleToClient := map[uint64]struct{}{}

			for _, nodeState := range nodeInfos {
				idsVisibleToClient[nodeState.NodeId.GetId()] = struct{}{}
			}

			idsNotVisibleToClient := getNodesNotVisibleToClient(idsVisibleToAllNodes, idsVisibleToClient)

			if len(idsNotVisibleToClient) != 0 {
				if !isLB {
					// TODO handle case where only seedConn are available.
					view.Warningf(`Not all nodes are visible to asvec. 
Asvec can't reach: %s
Possible scenarios:	
1. You should use --host instead of --seeds to indicate you are connection through a load balancer.
2. Asvec was able to connect to your seeds but the server(s) are returning unreachable endpoints. Did you forget --listener-name.
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

func getAllNodesInfo(ctx context.Context, adminClient *avs.AdminClient) []*writers.NodeInfo {
	nodeIds := adminClient.NodeIDs(ctx)

	logger.Debug("received node ids", slog.Any("nodeIds", nodeIds))

	if len(nodeIds) == 0 {
		// Loadbalancer must be enabled. Passing nil to admin client
		// causes it to fetch info from seed node
		nodeIds = append(nodeIds, nil)
	}

	nodeInfos := make([]*writers.NodeInfo, len(nodeIds))
	wg := sync.WaitGroup{}

	for i, nodeId := range nodeIds {
		wg.Add(1)
		go func(i int, nodeId *protos.NodeId) {
			defer wg.Done()

			l := logger.With("node", nodeId.String())

			if nodeId == nil {
				nodeInfos[i] = &writers.NodeInfo{NodeId: &protos.NodeId{Id: 0}}
			} else {
				nodeInfos[i] = &writers.NodeInfo{NodeId: nodeId}
			}

			wg.Add(1)
			go func() {
				defer wg.Done()

				connectedEndpoint, err := adminClient.ConnectedNodeEndpoint(ctx, nodeId)
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

				endpoints, err := adminClient.ClusterEndpoints(ctx, nodeId, nil) // TODO add option listener name
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

				state, err := adminClient.ClusteringState(ctx, nodeId)
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

				about, err := adminClient.About(ctx, nodeId)
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
		}(i, nodeId)
	}

	wg.Wait()

	return nodeInfos
}

func getIDsVisibleToAllNodes(nodeInfos []*writers.NodeInfo) map[uint64]struct{} {
	idsVisibleToAllNodes := map[uint64]struct{}{}

	for _, nodeState := range nodeInfos {
		for toId := range nodeState.Endpoints.GetEndpoints() {
			idsVisibleToAllNodes[toId] = struct{}{}
		}
	}

	return idsVisibleToAllNodes
}

func getIDsVisibleToEachNode(nodeInfos []*writers.NodeInfo) map[uint64]map[uint64]struct{} {
	idsVisibleToIndividualNodes := map[uint64]map[uint64]struct{}{} // using map as set

	for _, nodeState := range nodeInfos {
		fromId := nodeState.NodeId.GetId()

		for toId := range nodeState.Endpoints.GetEndpoints() {
			if _, ok := idsVisibleToIndividualNodes[fromId]; !ok {
				idsVisibleToIndividualNodes[fromId] = map[uint64]struct{}{}
			}

			idsVisibleToIndividualNodes[fromId][toId] = struct{}{}
		}
	}

	return idsVisibleToIndividualNodes
}

func getNodesNotVisibleToClient(idsVisibleToAllNodes map[uint64]struct{}, idsVisibleToClient map[uint64]struct{}) []string {
	nodesNotVisibleToClient := []string{}

	for id := range idsVisibleToAllNodes {
		if _, ok := idsVisibleToClient[id]; !ok {
			nodesNotVisibleToClient = append(nodesNotVisibleToClient, strconv.FormatUint(id, 10))
		}
	}

	return nodesNotVisibleToClient
}

func getNodesNotVisibleToEachNode(idsVisibleToEachNode map[uint64]map[uint64]struct{}, idsVisibleToAllNodes map[uint64]struct{}) map[uint64][]string {
	nodesNotVisibleToEachNode := map[uint64][]string{}

	for id := range idsVisibleToAllNodes {
		for fromId, visibleFromNode := range idsVisibleToEachNode {
			if _, ok := visibleFromNode[id]; !ok {
				if _, ok := nodesNotVisibleToEachNode[fromId]; !ok {
					nodesNotVisibleToEachNode[fromId] = []string{}
				}

				nodesNotVisibleToEachNode[fromId] = append(nodesNotVisibleToEachNode[fromId], strconv.FormatUint(id, 10))
			}
		}
	}

	return nodesNotVisibleToEachNode
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
