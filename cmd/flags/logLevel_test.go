package flags

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type LogLevelSuite struct {
	suite.Suite
}

func TestLogLevelSuite(t *testing.T) {
	suite.Run(t, new(LogLevelSuite))
}

func (suite *LogLevelSuite) TestNotSet() {
	flag := LogLevelFlag("")
	suite.True(flag.NotSet())

	flag = LogLevelFlag("DEBUG")
	suite.False(flag.NotSet())
}

func (suite *LogLevelSuite) TestSet() {
	flag := LogLevelFlag("")
	err := flag.Set("DEBUG")
	suite.NoError(err)
	suite.Equal(LogLevelFlag("DEBUG"), flag)

	err = flag.Set("INVALID")
	suite.Error(err)
	suite.Equal(LogLevelFlag("DEBUG"), flag)
}

func (suite *LogLevelSuite) TestType() {
	flag := LogLevelFlag("")
	suite.Equal("DEBUG,INFO,WARN,ERROR", flag.Type())
}

func (suite *LogLevelSuite) TestString() {
	flag := LogLevelFlag("DEBUG")
	suite.Equal("DEBUG", flag.String())
}
