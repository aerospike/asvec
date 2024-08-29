//go:build unit

package flags

import (
	"fmt"
	"reflect"
	"testing"
)

func TestVectorFlag_Set(t *testing.T) {
	testCases := []struct {
		name                 string
		testVector           string
		expectedBoolSlice    []bool
		expectedFloat32Slice []float32
		expectedError        error
	}{
		{
			name:              "Valid float vector",
			testVector:        "[1,0,t,f,true,false]",
			expectedBoolSlice: []bool{true, false, true, false, true, false},
			expectedError:     nil,
		},
		{
			name:                 "Valid float vector",
			testVector:           "[1.0,2,3.0]",
			expectedFloat32Slice: []float32{1.0, 2.0, 3.0},
			expectedError:        nil,
		},
		{
			name:              "Empty vector",
			testVector:        "[]",
			expectedBoolSlice: nil,
			expectedError:     fmt.Errorf("failed to parse float vector: strconv.ParseFloat: parsing \"\": invalid syntax"),
		},
		{
			name:                 "Invalid vector format",
			testVector:           "[1.0,2.0,w",
			expectedFloat32Slice: nil,
			expectedError:        fmt.Errorf("failed to parse float vector: strconv.ParseFloat: parsing \"w\": invalid syntax"),
		},
		// Add more test cases here...
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var flag VectorFlag

			err := flag.Set(tc.testVector)
			if err != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("Expected error %v, got %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(flag.FloatSlice, tc.expectedFloat32Slice) {
				t.Errorf("Expected slice %v, got %v", tc.expectedFloat32Slice, flag.FloatSlice)
			}

			if !reflect.DeepEqual(flag.BoolSlice, tc.expectedBoolSlice) {
				t.Errorf("Expected slice %v, got %v", tc.expectedBoolSlice, flag.BoolSlice)
			}
		})
	}
}

func TestVectorFlag_String(t *testing.T) {
	expectedString := "[1.1,2.2,3.3]"
	var flag1 VectorFlag
	flag1.Set(expectedString)

	if flag1.String() != expectedString {
		t.Errorf("Expected string representation %s, got %s", expectedString, flag1.String())
	}

	expectedString = "[1,0,1,0,1,0]"
	var flag2 VectorFlag
	flag2.Set(expectedString)

	if flag2.String() != expectedString {
		t.Errorf("Expected string representation %s, got %s", expectedString, flag2.String())
	}
}

func TestVectorFlag_Type(t *testing.T) {
	var flag VectorFlag
	expectedType := "[]float32 or []int"

	if flag.Type() != expectedType {
		t.Errorf("Expected type %s, got %s", expectedType, flag.Type())
	}
}
