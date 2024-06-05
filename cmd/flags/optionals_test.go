package flags

import (
	"testing"

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
