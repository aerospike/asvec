package cmd

import (
	"fmt"
	"net"
	"strings"

	"github.com/aerospike/aerospike-proximus-client-go/protos"
	"github.com/spf13/pflag"
)

const (
	flagNameSeeds     = "seeds"
	flagNamePort      = "port"
	flagNameNamespace = "namespace"
	flagNameSets      = "sets"
	flagNameIndexName = "index-name"
	flagNameVector    = "vector-field"
	flagNameDimension = "dimension"
	flagNameDistance  = "distance-metric"
	flagNameIndexMeta = "index-meta"
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
	fsb.StringP(flagNameVector, "v", "vector-field", "The name of the vector field.")

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

type DistanceMetricFlag string

// This is just a set of valid VectorDistanceMetrics. The value does not have meaning
var distanceMetricSet = protos.VectorDistanceMetric_value

func (mode *DistanceMetricFlag) Set(val string) error {
	val = strings.ToUpper(val)
	if val, ok := distanceMetricSet[val]; ok {
		*mode = DistanceMetricFlag(val)
		return nil
	}

	return fmt.Errorf("unrecognized distance metric")
}

func (mode *DistanceMetricFlag) Type() string {
	names := []string{}

	for key := range distanceMetricSet {
		names = append(names, key)
	}

	return strings.Join(names, ",")
}

func (mode *DistanceMetricFlag) String() string {
	return string(*mode)
}
