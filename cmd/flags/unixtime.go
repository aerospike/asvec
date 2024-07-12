package flags

import (
	"fmt"
	"strconv"
	"time"
)

type UnixTimestampFlag time.Time

func (f *UnixTimestampFlag) Set(val string) error {
	timestamp, err := strconv.ParseUint(val, 0, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	*f = UnixTimestampFlag(time.Unix(int64(timestamp), 0))

	return nil
}

func (f *UnixTimestampFlag) Type() string {
	return "unix-timestamp (sec)"
}

func (f *UnixTimestampFlag) String() string {
	return fmt.Sprintf("%d", time.Time(*f).Unix())
}

func (f *UnixTimestampFlag) Time() time.Time {
	return time.Time(*f)
}
