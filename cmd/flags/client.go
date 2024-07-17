package flags

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/pflag"
)

//nolint:govet // Padding not a concern for a CLI
type ClientFlags struct {
	Host            *HostPortFlag
	Seeds           *SeedsSliceFlag
	ListenerName    StringOptionalFlag
	AuthCredentials CredentialsFlag
	Timeout         time.Duration
	TLSFlags
}

func NewClientFlags() *ClientFlags {
	return &ClientFlags{
		Host:            NewDefaultHostPortFlag(),
		Seeds:           &SeedsSliceFlag{},
		AuthCredentials: CredentialsFlag{},
		TLSFlags:        *NewTLSFlags(),
	}
}

func (cf *ClientFlags) NewClientFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.VarP(cf.Host, Host, "h", fmt.Sprintf("The AVS host to connect to. If cluster discovery is needed use --%s. Additionally can be set using the environment variable ASVEC_HOST.", Seeds))                                                                                                     //nolint:lll // For readability
	flagSet.Var(cf.Seeds, Seeds, fmt.Sprintf("The AVS seeds to use for cluster discovery. If no cluster discovery is needed (i.e. load-balancer) then use --%s. Additionally can be set using the environment variable ASVEC_SEEDS.", Host))                                                            //nolint:lll // For readability
	flagSet.VarP(&cf.ListenerName, ListenerName, "l", "The listener to ask the AVS server for as configured in the AVS server. Likely required for cloud deployments.")                                                                                                                                 //nolint:lll // For readability
	flagSet.VarP(&cf.AuthCredentials.User, AuthUser, "U", "The AVS user used to authenticate. Additionally can be set using the environment variable ASVEC_USER")                                                                                                                                       //nolint:lll // For readability
	flagSet.VarP(&cf.AuthCredentials.Password, AuthPassword, "P", "The AVS password for the specified user. If a password is not provided you will be prompted. Additionally can be set using the environment variable ASVEC_PASSWORD.")                                                                //nolint:lll // For readability
	flagSet.VarP(&cf.AuthCredentials, AuthCredentials, "C", "The AVS user and password used to authenticate. Additionally can be set using the environment variable ASVEC_CREDENTIALS. If a password is not provided you will be prompted. This flag is provided in addition to --user and --password") //nolint:lll // For readability
	flagSet.DurationVar(&cf.Timeout, Timeout, time.Second*5, "The timeout to use for each request to AVS")                                                                                                                                                                                              //nolint:lll // For readability
	flagSet.AddFlagSet(cf.NewTLSFlagSet(func(s string) string { return s }))

	return flagSet
}

func (cf *ClientFlags) NewSLogAttr() []any {
	logPass := ""
	if cf.AuthCredentials.Password.String() != "" {
		logPass = "*"
	}

	return []any{slog.String(Host, cf.Host.String()),
		slog.String(Seeds, cf.Seeds.String()),
		slog.String(ListenerName, cf.ListenerName.String()),
		slog.String(AuthUser, cf.AuthCredentials.String()),
		slog.String(AuthPassword, logPass),
		slog.Bool(TLSCaFile, cf.TLSRootCAFile != nil),
		slog.Bool(TLSCaPath, cf.TLSRootCAPath != nil),
		slog.Bool(TLSCertFile, cf.TLSCertFile != nil),
		slog.Bool(TLSKeyFile, cf.TLSKeyFile != nil),
		slog.Bool(TLSKeyFilePass, cf.TLSKeyFilePass != nil),
		slog.Duration(Timeout, cf.Timeout),
	}
}
