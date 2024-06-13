//go:build unit

package cmd

import (
	"asvec/cmd/flags"
	"testing"

	avs "github.com/aerospike/aerospike-proximus-client-go"
	"github.com/stretchr/testify/assert"
)

func TestParseBothHostSeedsFlag(t *testing.T) {
	testCases := []struct {
		seeds                  *flags.SeedsSliceFlag
		host                   *flags.HostPortFlag
		expectedSlice          avs.HostPortSlice
		expectedIsLoadBalancer bool
	}{
		{
			&flags.SeedsSliceFlag{
				Seeds: avs.HostPortSlice{
					avs.NewHostPort("1.1.1.1", 5000, false),
				},
			},
			flags.NewDefaultHostPortFlag(),
			avs.HostPortSlice{
				avs.NewHostPort("1.1.1.1", 5000, false),
			},
			false,
		},
		{
			&flags.SeedsSliceFlag{
				Seeds: avs.HostPortSlice{},
			},
			flags.NewDefaultHostPortFlag(),
			avs.HostPortSlice{
				&flags.NewDefaultHostPortFlag().HostPort,
			},
			true,
		},
	}

	for _, tc := range testCases {
		actualSlice, actualBool := parseBothHostSeedsFlag(tc.seeds, tc.host)
		assert.Equal(t, tc.expectedSlice, actualSlice)
		assert.Equal(t, tc.expectedIsLoadBalancer, actualBool)
	}
}
