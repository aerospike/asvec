package flags

import (
	"strings"
)

type CredentialsFlag struct {
	User     StringOptionalFlag
	Password StringOptionalFlag
}

func (f *CredentialsFlag) Set(val string) error {
	splitVal := strings.SplitN(val, ":", 2)

	f.User.Set(splitVal[0])

	if len(splitVal) > 1 {
		f.Password.Set(splitVal[1])
	}

	return nil
}

func (f *CredentialsFlag) Type() string {
	return "<user>[:<pass>]"
}

func (f *CredentialsFlag) String() string {
	return strings.Join([]string{f.User.String(), f.Password.String()}, ":")
}
