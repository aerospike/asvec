package flags

import (
	"fmt"
	"log/slog"

	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/pflag"
)

type ClientFlags struct {
	Host         *HostPortFlag
	Seeds        *SeedsSliceFlag
	ListenerName StringOptionalFlag
	User         StringOptionalFlag
	Password     commonFlags.PasswordFlag
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
	flagSet.VarP(&cf.ListenerName, ListenerName, "l", commonFlags.DefaultWrapHelpString("The listener to ask the AVS server for as configured in the AVS server. Likely required for cloud deployments.")) //nolint:lll // For readability
	flagSet.VarP(&cf.User, User, "U", commonFlags.DefaultWrapHelpString("The AVS user to authenticate with."))                                                                                             //nolint:lll // For readability
	flagSet.VarP(&cf.Password, Password, "P", commonFlags.DefaultWrapHelpString("The AVS password for the specified user."))                                                                               //nolint:lll // For readability
	flagSet.AddFlagSet(cf.NewTLSFlagSet(commonFlags.DefaultWrapHelpString))

	return flagSet
}

func (cf *ClientFlags) NewSLogAttr() []any {
	return []any{slog.String(Host, cf.Host.String()),
		slog.String(Seeds, cf.Seeds.String()),
		slog.String(ListenerName, cf.ListenerName.String()),
		slog.String(User, cf.User.String()),
		slog.String(Password, cf.Password.String()),
		slog.Bool(TLSCaFile, cf.TLSRootCAFile != nil),
		slog.Bool(TLSCaPath, cf.TLSRootCAPath != nil),
		slog.Bool(TLSCertFile, cf.TLSCertFile != nil),
		slog.Bool(TLSKeyFile, cf.TLSKeyFile != nil),
		slog.Bool(TLSKeyFilePass, cf.TLSKeyFilePass != nil),
	}
}
