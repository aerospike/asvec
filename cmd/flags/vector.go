package flags

import (
	"fmt"
	"strconv"
	"strings"
)

// A cobra PFlag to parse and display help info for the host[:port]
// input option.  It implements the pflag Value and SliceValue interfaces to
// enable automatic parsing by cobra.
type VectorFlag struct {
	FloatSlice []float32
	BoolSlice  []bool
}

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
		// Will fail with 1.0 or 0.0
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
	if v.BoolSlice != nil {
		return fmt.Sprintf("%v", v.BoolSlice)
	}

	return fmt.Sprintf("%v", v.FloatSlice)
}

func (v *VectorFlag) IsSet() bool {
	return v.BoolSlice != nil || v.FloatSlice != nil
}
