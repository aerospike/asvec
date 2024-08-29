package flags

import (
	"asvec/cmd/writers"
	"fmt"
	"strconv"
	"strings"
)

// A cobra PFlag to parse and store a vector of floats or booleans.
type VectorFlag struct {
	FloatSlice []float32
	BoolSlice  []bool
}

// NewVectorFlag returns a new VectorFlag. It parses either a bool or float
// array into the appropriate type. Boolean vectors look like [true,false,0,1].
// Anything else is parsed as a float [1.0,1,2,3,4.123]
func (v *VectorFlag) Set(val string) error {
	val = strings.Trim(val, "[]")
	val = strings.ReplaceAll(val, " ", "")
	val = strings.ReplaceAll(val, "\n", "")

	ss := strings.Split(val, ",")

	if len(ss) == 0 {
		return fmt.Errorf("empty vector not allowed")
	}

	boolSlice := make([]bool, len(ss))
	lastIdx := 0

	for i, s := range ss {
		// Will fail with 1.0 or 0.0 but pass on true,false,0,1,t,f
		tempBool, err := strconv.ParseBool(s)
		if err != nil {
			// Try to parse as float
			break
		}

		boolSlice[i] = tempBool
		lastIdx++
	}

	if lastIdx == len(ss) {
		v.BoolSlice = boolSlice
		return nil
	}

	float32Slice := make([]float32, len(ss))

	for i, s := range ss {
		tempFloat, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return fmt.Errorf("failed to parse float vector: %v", err)
		}

		float32Slice[i] = float32(tempFloat)
	}

	v.FloatSlice = float32Slice

	return nil
}

func (v *VectorFlag) Type() string {
	return "[]float32 or []int"
}

func (v *VectorFlag) String() string {
	vals := make([]string, max(len(v.BoolSlice), len(v.FloatSlice)))

	if v.BoolSlice != nil {
		for i, b := range v.BoolSlice {
			if b {
				vals[i] = "1"
			} else {
				vals[i] = "0"
			}
		}
	} else {
		for i, f := range v.FloatSlice {
			vals[i] = writers.RemoveTrailingZeros(fmt.Sprintf("%f", f))
		}
	}

	return fmt.Sprintf("[%s]", strings.Join(vals, ","))
}

func (v *VectorFlag) IsSet() bool {
	return v.BoolSlice != nil || v.FloatSlice != nil
}
