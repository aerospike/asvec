package flags

import (
	"fmt"
	"sort"
	"strings"
)

type LogLevelFlag string

var logLevelSet = map[string]int{
	"DEBUG": 0,
	"INFO":  1,
	"WARN":  2,
	"ERROR": 3,
}

func (f *LogLevelFlag) NotSet() bool {
	return *f == ""
}

func (f *LogLevelFlag) Set(val string) error {
	if val == "" {
		*f = LogLevelFlag("")
		return nil
	}

	val = strings.ToUpper(val)
	if _, ok := logLevelSet[val]; ok {
		*f = LogLevelFlag(val)
		return nil
	}

	return fmt.Errorf("unrecognized log level")
}

func (f *LogLevelFlag) Type() string {
	return "enum"
}

func (f *LogLevelFlag) String() string {
	return string(*f)
}

func LogLevelEnum() []string {
	names := []string{}

	for key := range logLevelSet {
		names = append(names, key)
	}

	sort.Slice(names, func(i, j int) bool {
		return logLevelSet[names[i]] < logLevelSet[names[j]]
	})

	return names
}
