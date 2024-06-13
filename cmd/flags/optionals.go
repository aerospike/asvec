package flags

import "strconv"

const optionalEmptyString = "<nil>"

type StringOptionalFlag struct {
	Val *string
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
