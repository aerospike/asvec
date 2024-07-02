package flags

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	avs "github.com/aerospike/aerospike-proximus-client-go"
)

const (
	DefaultIPv4 = "127.0.0.1"
	DefaultPort = 5000
)

func parseHostPort(v string) (*avs.HostPort, error) {
	host := &avs.HostPort{}
	ipv6HostPattern := `^\[(?P<host>.*)\]`
	hostPattern := `^(?P<host>[^:]*)` // matched ipv4 and hostname
	portPattern := `(?P<port>\d+)$`
	reIPv6Host := regexp.MustCompile(fmt.Sprintf("%s$", ipv6HostPattern))
	reIPv6HostPort := regexp.MustCompile(fmt.Sprintf("%s:%s", ipv6HostPattern, portPattern))
	reIPv4Host := regexp.MustCompile(fmt.Sprintf("%s$", hostPattern))
	reIPv4HostPort := regexp.MustCompile(fmt.Sprintf("%s:%s", hostPattern, portPattern))

	regexsAndNames := []struct {
		regex      *regexp.Regexp
		groupNames []string
	}{
		// The order is important since the ipv4 pattern also matches ipv6
		{reIPv6HostPort, reIPv6HostPort.SubexpNames()},
		{reIPv6Host, reIPv6Host.SubexpNames()},
		{reIPv4HostPort, reIPv4HostPort.SubexpNames()},
		{reIPv4Host, reIPv4Host.SubexpNames()},
	}

	for _, r := range regexsAndNames {
		regex := r.regex
		groupNames := r.groupNames

		if matchs := regex.FindStringSubmatch(v); matchs != nil {
			for idx, match := range matchs {
				var err error

				name := groupNames[idx]

				switch {
				case name == "host":
					host.Host = match
				case name == "port":
					var intPort int64

					intPort, err = strconv.ParseInt(match, 0, 0)
					if err == nil {
						host.Port = int(intPort)
					}
				}

				if err != nil {
					return host, fmt.Errorf("failed to parse %s : %s", name, err)
				}
			}

			if host.Port == 0 {
				host.Port = DefaultPort
			}

			return host, nil
		}
	}

	return host, fmt.Errorf("does not match any expected formats")
}

// A cobra PFlag to parse and display help info for the host[:port]
// input option.  It implements the pflag Value and SliceValue interfaces to
// enable automatic parsing by cobra.
type HostPortFlag struct {
	HostPort avs.HostPort
}

func NewDefaultHostPortFlag() *HostPortFlag {
	return &HostPortFlag{
		HostPort: avs.HostPort{
			Host: DefaultIPv4,
			Port: DefaultPort,
		},
	}
}

func (hp *HostPortFlag) Set(val string) error {
	hostPort, err := parseHostPort(val)
	if err != nil {
		return err
	}

	hp.HostPort = *hostPort

	return nil
}

func (hp *HostPortFlag) Type() string {
	return "host[:port]"
}

func (hp *HostPortFlag) String() string {
	return hp.HostPort.String()
}

type SeedsSliceFlag struct {
	Seeds avs.HostPortSlice
}

func NewSeedsSliceFlag() SeedsSliceFlag {
	return SeedsSliceFlag{
		Seeds: avs.HostPortSlice{},
	}
}

// Append adds the specified value to the end of the flag value list.
func (slice *SeedsSliceFlag) Append(val string) error {
	host, err := parseHostPort(val)

	if err != nil {
		return err
	}

	slice.Seeds = append(slice.Seeds, host)

	return nil
}

// Replace will fully overwrite any data currently in the flag value list.
func (slice *SeedsSliceFlag) Replace(vals []string) error {
	slice.Seeds = avs.HostPortSlice{}

	for _, val := range vals {
		if err := slice.Append(val); err != nil {
			return err
		}
	}

	return nil
}

// GetSlice returns the flag value list as an array of strings.
func (slice *SeedsSliceFlag) GetSlice() []string {
	strs := []string{}

	for _, elem := range slice.Seeds {
		strs = append(strs, elem.String())
	}

	return strs
}

func (slice *SeedsSliceFlag) Set(commaSepVal string) error {
	vals := strings.Split(commaSepVal, ",")

	for _, val := range vals {
		if err := slice.Append(val); err != nil {
			return err
		}
	}

	return nil
}

func (slice *SeedsSliceFlag) Type() string {
	return "seed[:port][,...]"
}

func (slice *SeedsSliceFlag) String() string {
	return slice.Seeds.String()
}
