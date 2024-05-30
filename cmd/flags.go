package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/aerospike/aerospike-proximus-client-go/protos"
	"github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	flagNameSeeds          = "seeds"
	flagNameHost           = "host"
	flagNameListenerName   = "listener-name"
	flagNameNamespace      = "namespace"
	flagNameSets           = "sets"
	flagNameIndexName      = "index-name"
	flagNameVectorField    = "vector-field"
	flagNameDimension      = "dimension"
	flagNameDistanceMetric = "distance-metric"
	flagNameIndexMeta      = "index-meta"
	flagNameTimeout        = "timeout"
	flagNameVerbose        = "verbose"
)

func viperGetIfSetString(flagName string) *string {
	if viper.IsSet(flagName) {
		s := viper.GetString(flagName)
		return &s
	}

	return nil
}

func viperGetIfSetBool(flagName string) *bool {
	if viper.IsSet(flagName) {
		s := viper.GetBool(flagName)
		return &s
	}

	return nil
}

func viperGetIfSetUint32(flagName string) *uint32 {
	if viper.IsSet(flagName) {
		s := viper.GetUint32(flagName)
		return &s
	}

	return nil
}

type FlagSetBuilder struct {
	*pflag.FlagSet
}

func NewFlagSetBuilder(flagSet *pflag.FlagSet) *FlagSetBuilder {
	return &FlagSetBuilder{
		flagSet,
	}
}

// TODO: Should this be a list of IPs? Should we support IP:PORT?
func (fsb *FlagSetBuilder) AddSeedFlag() {
	fsb.StringArrayP(flagNameSeeds, "s", []string{}, flags.DefaultWrapHelpString(fmt.Sprintf("The AVS seeds to use for cluster discovery. If no cluster discovery is needed (i.e. load-balancer) then use --%s", flagNameHost)))
}

func (fsb *FlagSetBuilder) AddHostFlag() {
	fsb.StringP(flagNameHost, "h", "127.0.0.1:5000", fmt.Sprintf("The AVS host to connect to. If cluster discovery is needed use --%s", flagNameSeeds))
}

func (fsb *FlagSetBuilder) AddNamespaceFlag() {
	fsb.StringP(flagNameNamespace, "n", "", "The namespace for the index.")
}

func (fsb *FlagSetBuilder) AddSetsFlag() {
	fsb.StringArrayP(flagNameSets, "s", nil, "The sets for the index.")
}

func (fsb *FlagSetBuilder) AddIndexNameFlag() {
	fsb.StringP(flagNameIndexName, "i", "", "The name of the index.")

}

func (fsb *FlagSetBuilder) AddVectorFieldFlag() {
	fsb.StringP(flagNameVectorField, "f", "", "The name of the vector field.")

}

func (fsb *FlagSetBuilder) AddDimensionFlag() {
	fsb.IntP(flagNameDimension, "d", 0, "The dimension of the vector field.")

}

func (fsb *FlagSetBuilder) AddDistanceMetricFlag() {
	distMetric := DistanceMetricFlag("")
	fsb.VarP(&distMetric, flagNameDistanceMetric, "m", "The distance metric for the index.")
}

func (fsb *FlagSetBuilder) AddIndexMetaFlag() {
	fsb.StringToStringP(flagNameIndexMeta, "e", nil, "The metadata for the index.")
}

func (fsb *FlagSetBuilder) AddTimeoutFlag() {
	fsb.DurationP(flagNameTimeout, "t", time.Second*5, "The timeout used for each request.")
}

func (fsb *FlagSetBuilder) AddVerbose() {
	fsb.BoolP(flagNameVerbose, "v", false, "Display extra detail about an index.")
}

type DistanceMetricFlag string

// This is just a set of valid VectorDistanceMetrics. The value does not have meaning
var distanceMetricSet = protos.VectorDistanceMetric_value

func (f *DistanceMetricFlag) Set(val string) error {
	val = strings.ToUpper(val)
	if _, ok := distanceMetricSet[val]; ok {
		*f = DistanceMetricFlag(val)
		return nil
	}

	return fmt.Errorf("unrecognized distance metric")
}

func (f *DistanceMetricFlag) Type() string {
	names := []string{}

	for key := range distanceMetricSet {
		names = append(names, key)
	}

	return strings.Join(names, ",")
}

func (f *DistanceMetricFlag) String() string {
	return string(*f)
}

type LogLevelFlag string

var logLevelSet = map[string]struct{}{
	"DEBUG": {},
	"INFO":  {},
	"WARN":  {},
	"ERROR": {},
}

func (f *LogLevelFlag) Set(val string) error {
	val = strings.ToUpper(val)
	if _, ok := logLevelSet[val]; ok {
		*f = LogLevelFlag(val)
		return nil
	}

	return fmt.Errorf("unrecognized log level")
}

func (f *LogLevelFlag) Type() string {
	names := []string{}

	for key := range logLevelSet {
		names = append(names, key)
	}

	return strings.Join(names, ",")
}

func (f *LogLevelFlag) String() string {
	return string(*f)
}
