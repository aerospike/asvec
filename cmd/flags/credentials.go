package flags

import (
	"fmt"
	"strings"
)

type CredentialsFlag struct {
	User     StringOptionalFlag
	Password StringOptionalFlag
}

func (f *CredentialsFlag) Set(val string) error {
	splitVal := strings.SplitN(val, ":", 2)

	err := f.User.Set(splitVal[0])
	if err != nil {
		return fmt.Errorf("could not parse user: %w", err)
	}

	if len(splitVal) > 1 {
		err = f.Password.Set(splitVal[1])
		if err != nil {
			return fmt.Errorf("could not parse password: %w", err)
		}
	}

	return nil
}

func (f *CredentialsFlag) Type() string {
	return "<user>[:<pass>]"
}

func (f *CredentialsFlag) String() string {
	return f.User.String() + ":" + f.Password.String()
}
