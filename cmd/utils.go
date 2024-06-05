package cmd

import (
	"asvec/cmd/flags"

	avs "github.com/aerospike/aerospike-proximus-client-go"
)

func parseBothHostSeedsFlag(seeds flags.SeedsSliceFlag, host flags.HostPortFlag) (avs.HostPortSlice, bool) {
	isLoadBalancer := false
	hosts := avs.HostPortSlice{}

	if len(seeds.Seeds) > 0 {
		logger.Debug("seeds is set")

		hosts = append(hosts, seeds.Seeds...)
	} else {
		logger.Debug("hosts is set")

		isLoadBalancer = true

		hosts = append(hosts, &host.HostPort)
	}

	return hosts, isLoadBalancer
}
