package flags

import (
	"fmt"
	"log/slog"
	"time"

	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/pflag"
)

type ClientFlags struct {
	Host         *HostPortFlag
	Seeds        *SeedsSliceFlag
	ListenerName StringOptionalFlag
	User         StringOptionalFlag
	Password     commonFlags.PasswordFlag
	Timeout      time.Duration
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
	flagSet.VarP(&cf.User, AuthUser, "U", commonFlags.DefaultWrapHelpString("The AVS user to authenticate with."))                                                                                         //nolint:lll // For readability
	flagSet.VarP(&cf.Password, AuthPassword, "P", commonFlags.DefaultWrapHelpString("The AVS password for the specified user."))                                                                           //nolint:lll // For readability
	flagSet.DurationVar(&cf.Timeout, Timeout, time.Second*5, commonFlags.DefaultWrapHelpString("The timeout to use for each request to AVS"))                                                              //nolint:lll // For readability
	flagSet.AddFlagSet(cf.NewTLSFlagSet(commonFlags.DefaultWrapHelpString))

	return flagSet
}

func (cf *ClientFlags) NewSLogAttr() []any {
	logPass := ""
	if cf.Password.String() != "" {
		logPass = "*"
	}

	return []any{slog.String(Host, cf.Host.String()),
		slog.String(Seeds, cf.Seeds.String()),
		slog.String(ListenerName, cf.ListenerName.String()),
		slog.String(AuthUser, cf.User.String()),
		slog.String(AuthPassword, logPass),
		slog.Bool(TLSCaFile, cf.TLSRootCAFile != nil),
		slog.Bool(TLSCaPath, cf.TLSRootCAPath != nil),
		slog.Bool(TLSCertFile, cf.TLSCertFile != nil),
		slog.Bool(TLSKeyFile, cf.TLSKeyFile != nil),
		slog.Bool(TLSKeyFilePass, cf.TLSKeyFilePass != nil),
		slog.Duration(Timeout, cf.Timeout),
	}
}
