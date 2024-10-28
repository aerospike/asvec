//go:build unit

package flags

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type OptionalFlagSuite struct {
	suite.Suite
}

func TestOptionalFlagSuite(t *testing.T) {
	suite.Run(t, new(OptionalFlagSuite))
}

func (suite *OptionalFlagSuite) TestBoolOptionalFlag() {
	f := &BoolOptionalFlag{}

	err := f.Set("true")
	if err != nil {
		suite.T().Errorf("Unexpected error: %v", err)
	}

	if f.Val == nil || *f.Val != true {
		suite.T().Errorf("Expected true, got %v", f.Val)
	}

	err = f.Set("not a bool")
	if err == nil {
		suite.T().Errorf("Expected error, got nil")
	}
}

func (suite *OptionalFlagSuite) TestStringOptionalFlag() {
	f := &StringOptionalFlag{}

	err := f.Set("hello")
	if err != nil {
		suite.T().Errorf("Unexpected error: %v", err)
	}

	if f.Val == nil || *f.Val != "hello" {
		suite.T().Errorf("Expected 'hello', got %v", f.Val)
	}

	err = f.Set("")
	if err != nil {
		suite.T().Errorf("Unexpected error: %v", err)
	}

	if f.Val == nil {
		suite.T().Errorf("Expected not nil")
	}
}

func (suite *OptionalFlagSuite) TestUint32OptionalFlag() {
	f := &Uint32OptionalFlag{}

	err := f.Set("42")
	if err != nil {
		suite.T().Errorf("Unexpected error: %v", err)
	}

	if f.Val == nil || *f.Val != 42 {
		suite.T().Errorf("Expected 42, got %v", f.Val)
	}

	err = f.Set("not a number")
	if err == nil {
		suite.T().Errorf("Expected error, got nil")
	}
}

func (suite *OptionalFlagSuite) TestIntOptionalFlag() {
	f := &IntOptionalFlag{}

	err := f.Set("42")
	if err != nil {
		suite.T().Errorf("Unexpected error: %v", err)
	}

	if f.Val == nil || *f.Val != 42 {
		suite.T().Errorf("Expected 42, got %v", f.Val)
	}

	err = f.Set("not a number")
	if err == nil {
		suite.T().Errorf("Expected error, got nil")
	}
}

func (suite *OptionalFlagSuite) TestDurationOptionalFlag() {
	f := &DurationOptionalFlag{}

	err := f.Set("300ms")
	if err != nil {
		suite.T().Errorf("Unexpected error: %v", err)
	}

	if f.Val == nil || *f.Val != time.Duration(300)*time.Millisecond {
		suite.T().Errorf("Expected 300ms, got %v", f.Val)
	}

	err = f.Set("not a time")
	if err == nil {
		suite.T().Errorf("Expected error, got nil")
	}
}

func (suite *OptionalFlagSuite) TestInfDurationOptionalFlag() {
	f := &InfDurationOptionalFlag{}

	err := f.Set("-1")
	if err != nil {
		suite.T().Errorf("Unexpected error: %v", err)
	}

	suite.Equal("-1", f.String())
	suite.Equal(int64(-1), *f.Int64())
	f = &InfDurationOptionalFlag{}

	err = f.Set("20m")
	if err != nil {
		suite.T().Errorf("Unexpected error: %v", err)
	}

	expectedDuration := time.Duration(20) * time.Minute
	suite.Equal(expectedDuration.String(), f.String())
	suite.Equal(expectedDuration.Milliseconds(), *f.Int64())
}
