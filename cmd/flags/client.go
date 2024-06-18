package flags

import (
	"fmt"

	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/pflag"
)

type ClientFlags struct {
	Host         *HostPortFlag
	Seeds        *SeedsSliceFlag
	ListenerName StringOptionalFlag
	TLSFlags
}

func NewClientFlags() *ClientFlags {
	return &ClientFlags{
		Host:     NewDefaultHostPortFlag(),
		Seeds:    &SeedsSliceFlag{},
		TLSFlags: *NewTLSFlags(),
	}
}

func (cf *ClientFlags) NewClientFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.VarP(cf.Host, Host, "h", commonFlags.DefaultWrapHelpString(fmt.Sprintf("The AVS host to connect to. If cluster discovery is needed use --%s", Seeds)))                                         //nolint:lll // For readability
	flagSet.Var(cf.Seeds, Seeds, commonFlags.DefaultWrapHelpString(fmt.Sprintf("The AVS seeds to use for cluster discovery. If no cluster discovery is needed (i.e. load-balancer) then use --%s", Host))) //nolint:lll // For readability
	flagSet.VarP(&cf.ListenerName, ListenerName, "l", commonFlags.DefaultWrapHelpString("The listener to ask the AVS server for as configured in the AVS server. Likely required for cloud deployments."))

	flagSet.AddFlagSet(cf.NewTLSFlagSet(commonFlags.DefaultWrapHelpString))

	return flagSet
}
