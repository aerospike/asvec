package flags

import (
	"asvec/utils"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const optionalEmptyString = "<nil>"

type StringOptionalFlag struct {
	Val *string
}

func NewStringOptionalFlag() StringOptionalFlag {
	return StringOptionalFlag{}
}

func (f *StringOptionalFlag) Set(val string) error {
	f.Val = &val
	return nil
}

func (f *StringOptionalFlag) Type() string {
	return "string"
}

func (f *StringOptionalFlag) String() string {
	if f.Val != nil {
		return *f.Val
	}

	return optionalEmptyString
}

type Uint32OptionalFlag struct {
	Val *uint32
}

func (f *Uint32OptionalFlag) Set(val string) error {
	v, err := strconv.ParseUint(val, 0, 32)
	u32Val := uint32(v)
	f.Val = &u32Val

	return err
}

func (f *Uint32OptionalFlag) Type() string {
	return "uint32"
}

func (f *Uint32OptionalFlag) String() string {
	if f.Val != nil {
		return strconv.FormatUint(uint64(*f.Val), 10)
	}

	return optionalEmptyString
}

type Uint64OptionalFlag struct {
	Val *uint64
}

func (f *Uint64OptionalFlag) Set(val string) error {
	v, err := strconv.ParseUint(val, 0, 64)
	f.Val = &v

	return err
}

func (f *Uint64OptionalFlag) Type() string {
	return "uint64"
}

func (f *Uint64OptionalFlag) String() string {
	if f.Val != nil {
		return strconv.FormatUint(*f.Val, 10)
	}

	return optionalEmptyString
}

type IntOptionalFlag struct {
	Val *int64
}

func (f *IntOptionalFlag) Set(val string) error {
	v, err := strconv.ParseInt(val, 0, 64)
	f.Val = &v

	return err
}

func (f *IntOptionalFlag) Type() string {
	return "int"
}

func (f *IntOptionalFlag) String() string {
	if f.Val != nil {
		return strconv.FormatInt(*f.Val, 10)
	}

	return optionalEmptyString
}

type Float32OptionalFlag struct {
	Val *float32
}

func (f *Float32OptionalFlag) Set(val string) error {
	v, err := strconv.ParseFloat(val, 32)
	f32Val := float32(v)
	f.Val = &f32Val

	return err
}

func (f *Float32OptionalFlag) Type() string {
	return "float32"
}

func (f *Float32OptionalFlag) String() string {
	if f.Val != nil {
		return strconv.FormatFloat(float64(*f.Val), 'f', 2, 32)
	}

	return optionalEmptyString
}

type BoolOptionalFlag struct {
	Val *bool
}

func (f *BoolOptionalFlag) Set(val string) error {
	v, err := strconv.ParseBool(val)
	f.Val = &v

	return err
}

func (f *BoolOptionalFlag) Type() string {
	return "bool"
}

func (f *BoolOptionalFlag) String() string {
	if f.Val != nil {
		return strconv.FormatBool(*f.Val)
	}

	return optionalEmptyString
}

type DurationOptionalFlag struct {
	Val *time.Duration
}

func (f *DurationOptionalFlag) Set(val string) error {
	d, err := time.ParseDuration(val)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	f.Val = &d

	return err
}

func (f *DurationOptionalFlag) Type() string {
	return "time.Duration"
}

func (f *DurationOptionalFlag) String() string {
	if f.Val != nil {
		return f.Val.String()
	}

	return optionalEmptyString
}

func (f *DurationOptionalFlag) Uint64() *uint64 {
	if f.Val == nil {
		return nil
	}

	milli := uint64(f.Val.Milliseconds())

	return &milli
}

func (f *DurationOptionalFlag) Uint32() *uint32 {
	if f.Val == nil {
		return nil
	}

	milli := uint32(f.Val.Milliseconds())

	return &milli
}

func (f *DurationOptionalFlag) Int64() *int64 {
	if f.Val == nil {
		return nil
	}

	milli := f.Val.Milliseconds()

	return &milli
}

// InfDurationOptionalFlag is a flag that can be either a time.duration or -1 (never expire).
// It is used for flags like --hnsw-index-cache-expiry which can be set to never expire (-1)
type InfDurationOptionalFlag struct {
	duration   DurationOptionalFlag
	isInfinite bool
}

func (f *InfDurationOptionalFlag) Set(val string) error {
	err := f.duration.Set(val)
	if err == nil {
		return nil
	}

	val = strings.ToLower(val)

	if val == strconv.Itoa(Infinity) {
		f.isInfinite = true
	} else {
		return fmt.Errorf("invalid duration %s", val)
	}

	return nil
}

func (f *InfDurationOptionalFlag) Type() string {
	return "time.Duration"
}

func (f *InfDurationOptionalFlag) String() string {
	if f.isInfinite {
		return "-1"
	}

	if f.duration.Val != nil {
		return f.duration.String()
	}

	return optionalEmptyString
}

// Uint64 returns the duration as a uint64. If the duration is infinite, it returns -1.
// The AVS server uses -1 for cache expiry to represent infinity or never expire.
func (f *InfDurationOptionalFlag) Int64() *int64 {
	if f.isInfinite {
		return utils.Ptr(int64(Infinity))
	}

	return f.duration.Int64()
}
