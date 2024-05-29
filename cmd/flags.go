package cmd

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/aerospike/aerospike-proximus-client-go/protos"
	"github.com/spf13/pflag"
)

const (
	flagNameSeeds       = "seeds"
	flagNamePort        = "port"
	flagNameNamespace   = "namespace"
	flagNameSets        = "sets"
	flagNameIndexName   = "index-name"
	flagNameVectorField = "vector-field"
	flagNameDimension   = "dimension"
	flagNameDistance    = "distance-metric"
	flagNameIndexMeta   = "index-meta"
	flagNameTimeout     = "timeout"
	flagNameVerbose     = "verbose"
)

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
	fsb.IPP(flagNameSeeds, "h", net.ParseIP("127.0.0.1"), "The AVS seed host for cluster discovery.")
}

func (fsb *FlagSetBuilder) AddPortFlag() {
	fsb.IntP(flagNamePort, "p", 5000, "The AVS seed port for cluster discovery.")
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
	fsb.VarP(&distMetric, "distance-metric", "m", "The distance metric for the index.")
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
