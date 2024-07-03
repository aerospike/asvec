//go:build unit

package flags

import (
	"testing"

	avs "github.com/aerospike/avs-client-go"

	"github.com/stretchr/testify/suite"
)

type HostPortFlagTestSuite struct {
	suite.Suite
}

func (suite *HostPortFlagTestSuite) TestHostPortSetGet() {
	testCases := []struct {
		input        string
		output       *HostPortFlag
		expect_err   bool
		expected_str string
	}{
		{
			"127.0.0.1",
			&HostPortFlag{
				HostPort: avs.HostPort{
					Host: "127.0.0.1",
					Port: 5000,
				},
			},
			false,
			"127.0.0.1:5000",
		},
		{
			"127.0.0.2:3002",
			&HostPortFlag{
				HostPort: avs.HostPort{
					Host: "127.0.0.2",
					Port: 3002,
				},
			},
			false,
			"127.0.0.2:3002",
		},
		{
			"127.0.0.2:3002:",
			nil,
			true,
			"",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.input, func(t *testing.T) {
			actual := NewDefaultHostPortFlag()

			if tc.expect_err {
				suite.Error(actual.Set(tc.input))
			} else {
				suite.NoError(actual.Set(tc.input))
				suite.Equal(tc.output, actual)
				suite.Equal(tc.expected_str, actual.String())
			}
		})
	}
}

func (suite *HostPortFlagTestSuite) TestSeedsSliceSetGet() {
	testCases := []struct {
		input  string
		output SeedsSliceFlag
		slice  []string
	}{
		{
			"127.0.0.1",
			SeedsSliceFlag{
				Seeds: avs.HostPortSlice{
					{
						Host: "127.0.0.1",
						Port: 5000,
					},
				},
			},
			[]string{"127.0.0.1:5000"},
		},
		{
			"127.0.0.1,127.0.0.2",
			SeedsSliceFlag{
				Seeds: avs.HostPortSlice{
					{
						Host: "127.0.0.1",
						Port: 5000,
					},
					{
						Host: "127.0.0.2",
						Port: 5000,
					},
				},
			},
			[]string{"127.0.0.1:5000", "127.0.0.2:5000"},
		},
		{
			"127.0.0.2:3002",
			SeedsSliceFlag{
				Seeds: avs.HostPortSlice{
					{
						Host: "127.0.0.2",
						Port: 3002,
					},
				},
			},
			[]string{"127.0.0.2:3002"},
		},
		{
			"127.0.0.2:3002,127.0.0.3:3003",
			SeedsSliceFlag{

				Seeds: avs.HostPortSlice{
					{
						Host: "127.0.0.2",
						Port: 3002,
					},
					{
						Host: "127.0.0.3",
						Port: 3003,
					},
				},
			},
			[]string{"127.0.0.2:3002", "127.0.0.3:3003"},
		},
		{
			"127.0.0.3:3003",
			SeedsSliceFlag{

				Seeds: avs.HostPortSlice{
					{
						Host: "127.0.0.3",

						Port: 3003,
					},
				},
			},
			[]string{"127.0.0.3:3003"},
		},
		{
			"127.0.0.3:3003,127.0.0.4:3004",
			SeedsSliceFlag{

				Seeds: avs.HostPortSlice{
					{
						Host: "127.0.0.3",
						Port: 3003,
					},
					{
						Host: "127.0.0.4",
						Port: 3004,
					},
				},
			},
			[]string{"127.0.0.3:3003", "127.0.0.4:3004"},
		},
		{
			"127.0.0.3:3003,127.0.0.4:3004",
			SeedsSliceFlag{

				Seeds: avs.HostPortSlice{
					{
						Host: "127.0.0.3",
						Port: 3003,
					},
					{
						Host: "127.0.0.4",
						Port: 3004,
					},
				},
			},
			[]string{"127.0.0.3:3003", "127.0.0.4:3004"},
		},
		// Not supported yet, so not testing
		// {
		// 	"[2001:0db8:85a3:0000:0000:8a2e:0370:7334]",
		// 	SeedsSliceFlag{

		// 		Seeds: avs.HostPortSlice{
		// 			{
		// 				Host: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		// 			},
		// 		},
		// 	},
		// 	[]string{"2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		// },
		// {
		// 	"[fe80::1ff:fe23:4567:890a]:3002",
		// 	SeedsSliceFlag{

		// 		Seeds: avs.HostPortSlice{
		// 			{
		// 				Host: "fe80::1ff:fe23:4567:890a",
		// 				Port: 3002,
		// 			},
		// 		},
		// 	},
		// 	[]string{"fe80::1ff:fe23:4567:890a:3002"},
		// },
		// {
		// 	"[100::]:3003",
		// 	SeedsSliceFlag{

		// 		Seeds: avs.HostPortSlice{
		// 			{
		// 				Host: "100::",

		// 				Port: 3003,
		// 			},
		// 		},
		// 	},
		// 	[]string{"100:::3003"},
		// },
	}

	for _, tc := range testCases {
		suite.T().Run(tc.input, func(t *testing.T) {
			actual := NewSeedsSliceFlag()

			suite.NoError(actual.Set(tc.input))
			suite.Equal(tc.output, actual)
			suite.Equal(tc.slice, actual.GetSlice())
		})
	}
}

func (suite *HostPortFlagTestSuite) TestHostPortAppend() {
	testCases := []struct {
		input  string
		append string
		output SeedsSliceFlag
	}{
		{
			"127.0.0.1",
			"127.0.0.2",
			SeedsSliceFlag{

				Seeds: avs.HostPortSlice{
					{
						Host: "127.0.0.1",
						Port: 5000,
					},
					{
						Host: "127.0.0.2",
						Port: 5000,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.input, func(t *testing.T) {
			actual := NewSeedsSliceFlag()

			suite.NoError(actual.Set(tc.input))
			suite.NoError(actual.Set(tc.append))
			suite.Equal(tc.output, actual)
		})
	}
}

func (suite *HostPortFlagTestSuite) TestHostPortString() {
	testCases := []struct {
		input  SeedsSliceFlag
		output string
	}{
		{
			SeedsSliceFlag{

				Seeds: avs.HostPortSlice{
					{
						Host: "127.0.0.1",
						Port: 3000,
					},
					{
						Host: "127.0.0.2",
						Port: 3002,
					},
				},
			},
			"[127.0.0.1:3000, 127.0.0.2:3002]",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.output, func(t *testing.T) {
			suite.Equal(tc.output, tc.input.String())
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestRunHostTestSuite(t *testing.T) {
	suite.Run(t, new(HostPortFlagTestSuite))
}
