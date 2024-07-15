//go:build unit

package flags

import (
	"testing"
	"time"
)

func TestUnixTimestampFlag_Set(t *testing.T) {
	var flag UnixTimestampFlag
	testTimestamp := "1609459200" // Corresponds to 2021-01-01 00:00:00 UTC
	expectedTime := time.Unix(1609459200, 0)

	err := flag.Set(testTimestamp)
	if err != nil {
		t.Errorf("Failed to set UnixTimestampFlag: %v", err)
	}

	if !flag.Time().Equal(expectedTime) {
		t.Errorf("Expected time %v, got %v", expectedTime, flag.Time())
	}
}

func TestUnixTimestampFlag_String(t *testing.T) {
	expectedString := "1609459200"
	var flag UnixTimestampFlag
	flag.Set(expectedString)

	if flag.String() != expectedString {
		t.Errorf("Expected string representation %s, got %s", expectedString, flag.String())
	}
}

func TestUnixTimestampFlag_Type(t *testing.T) {
	var flag UnixTimestampFlag
	expectedType := "unix-timestamp (sec)"

	if flag.Type() != expectedType {
		t.Errorf("Expected type %s, got %s", expectedType, flag.Type())
	}
}
