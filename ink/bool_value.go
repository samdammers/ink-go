// Package ink provides the runtime for the ink language.
package ink

import "fmt"

// BoolValue is a value that holds a boolean.
type BoolValue struct {
	value[bool]
}

// NewBoolValue creates a new BoolValue.
func NewBoolValue(val bool) *BoolValue {
	bv := &BoolValue{}
	bv.Value = val
	return bv
}

// GetValueType returns the type of the value.
func (bv *BoolValue) GetValueType() ValueType {
	return ValueTypeBool
}

// IsTruthy returns true if the boolean is true.
func (bv *BoolValue) IsTruthy() bool {
	return bv.Value
}

// Cast converts the boolean to a new type.
func (bv *BoolValue) Cast(newType ValueType) (Value, error) {
	if newType == bv.GetValueType() {
		return bv, nil
	}

	switch newType {
	case ValueTypeInt:
		if bv.Value {
			return NewIntValue(1), nil
		}
		return NewIntValue(0), nil
	case ValueTypeFloat:
		if bv.Value {
			return NewFloatValue(1.0), nil
		}
		return NewFloatValue(0.0), nil
	case ValueTypeString:
		if bv.Value {
			return NewStringValue("true"), nil
		}
		return NewStringValue("false"), nil
	}

	return nil, fmt.Errorf("cannot cast BoolValue to %v", newType)
}

func (bv *BoolValue) String() string {
	if bv.Value {
		return "true"
	}
	return "false"
}
