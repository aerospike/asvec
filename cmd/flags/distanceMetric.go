package flags

import (
	"fmt"
	"slices"
	"strings"

	"github.com/aerospike/avs-client-go/protos"
)

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
	return FlagTypeEnum
}

func (f *DistanceMetricFlag) String() string {
	return string(*f)
}

func DistanceMetricEnum() []string {
	names := []string{}

	for key := range distanceMetricSet {
		names = append(names, key)
	}

	slices.Sort(names)

	return names
}
