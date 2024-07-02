//go:build unit

package flags

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type DistanceMetricFlagTestSuite struct {
	suite.Suite
}

func (suite *DistanceMetricFlagTestSuite) TestSet() {
	tests := []struct {
		input      string
		expect_err bool
		expected   DistanceMetricFlag
	}{
		{
			input:    "SQUARED_EUCLIDEAN",
			expected: DistanceMetricFlag("SQUARED_EUCLIDEAN"),
		},
		{
			input:    "cosine",
			expected: DistanceMetricFlag("COSINE"),
		},
		{
			input:    "manhattan",
			expected: DistanceMetricFlag("MANHATTAN"),
		},
		{
			input:    "dot_product",
			expected: DistanceMetricFlag("DOT_PRODUCT"),
		},
		{
			input:    "HAMming",
			expected: DistanceMetricFlag("HAMMING"),
		},
		{
			input:      "unknown",
			expect_err: true,
			// You can set the expected value to nil or any other appropriate value
			expected: DistanceMetricFlag(""),
		},
	}

	for _, test := range tests {
		suite.Run(test.input, func() {
			flag := DistanceMetricFlag("")
			err := flag.Set(test.input)
			if test.expect_err {
				suite.Error(err)
			} else {
				suite.NoError(err)
				suite.Equal(test.expected, flag)
			}
		})
	}
}

func (suite *DistanceMetricFlagTestSuite) TestType() {
	flag := DistanceMetricFlag("")
	suite.Equal("enum", flag.Type())
}

func (suite *DistanceMetricFlagTestSuite) TestString() {
	flag := DistanceMetricFlag("euclidean")
	suite.Equal("euclidean", flag.String())

	flag = DistanceMetricFlag("cosine")
	suite.Equal("cosine", flag.String())

	flag = DistanceMetricFlag("manhattan")
	suite.Equal("manhattan", flag.String())
}

func (suite *DistanceMetricFlagTestSuite) TestDistanceMetricEnum() {
	suite.Equal([]string{"COSINE", "DOT_PRODUCT", "HAMMING", "MANHATTAN", "SQUARED_EUCLIDEAN"}, DistanceMetricEnum())
}

func TestDistanceMetricFlagSuite(t *testing.T) {
	suite.Run(t, new(DistanceMetricFlagTestSuite))
}
