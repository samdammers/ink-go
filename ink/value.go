package ink

import (
	"fmt"
	"strconv"
)

// ValueType identifies the type of a Value object.
type ValueType int

const (
	// ValueTypeNoType is the default type for a value.
	ValueTypeNoType ValueType = iota
	// ValueTypeInt is the type for integer values.
	ValueTypeInt
	// ValueTypeBool is the type for boolean values.
	ValueTypeBool
	// ValueTypeFloat is the type for float values.
	ValueTypeFloat
	// ValueTypeString is the type for string values.
	ValueTypeString
	// ValueTypeVariablePointer is the type for variable pointer values.
	ValueTypeVariablePointer
	// ValueTypeDivertTarget is the type for divert target values.
	ValueTypeDivertTarget
	// ValueTypeList is the type for list values.
	ValueTypeList
)

// Value is the interface for all value types.
type Value interface {
	RuntimeObject
	GetValueType() ValueType
	IsTruthy() bool
	Cast(newType ValueType) (Value, error)
	GetValueObject() any
}

// value is the base struct for all value types.
// It is generic and embeds BaseRuntimeObject.
type value[T any] struct {
	BaseRuntimeObject
	Value T
}

// GetValueObject returns the underlying value as interface{}.
func (v *value[T]) GetValueObject() any {
	return v.Value
}

// --- StringValue ---

// StringValue is a value that holds a string.
type StringValue struct {
	value[string]
	isNewline          bool
	isInlineWhitespace bool
}

// NewStringValue creates a new StringValue.
func NewStringValue(val string) *StringValue {
	sv := &StringValue{}
	sv.Value = val
	sv.isNewline = (sv.Value == "\n")
	sv.isInlineWhitespace = true
	for _, r := range sv.Value {
		if r != ' ' && r != '\t' {
			sv.isInlineWhitespace = false
			break
		}
	}
	return sv
}

// GetValueType returns the type of the value.
func (sv *StringValue) GetValueType() ValueType {
	return ValueTypeString
}

// IsTruthy returns true if the string is not empty.
func (sv *StringValue) IsTruthy() bool {
	return len(sv.Value) > 0
}

// GetIsNewline returns true if the string is exactly a newline.
func (sv *StringValue) GetIsNewline() bool {
	return sv.isNewline
}

// Cast converts the string to a new type.
func (sv *StringValue) Cast(newType ValueType) (Value, error) {
	if newType == sv.GetValueType() {
		return sv, nil
	}

	switch newType {
	case ValueTypeInt:
		if i, err := strconv.Atoi(sv.Value); err == nil {
			return NewIntValue(i), nil
		}
		return nil, nil // Java returns null on failed cast
	case ValueTypeFloat:
		if f, err := strconv.ParseFloat(sv.Value, 64); err == nil {
			return NewFloatValue(f), nil
		}
		return nil, nil // Java returns null on failed cast
	}

	return nil, fmt.Errorf("cannot cast StringValue to %v", newType)
}

// --- IntValue ---

// IntValue is a value that holds an integer.
type IntValue struct {
	value[int]
}

// NewIntValue creates a new IntValue.
func NewIntValue(val int) *IntValue {
	iv := &IntValue{}
	iv.Value = val
	return iv
}

// GetValueType returns the type of the value.
func (iv *IntValue) GetValueType() ValueType {
	return ValueTypeInt
}

// IsTruthy returns true if the integer is not zero.
func (iv *IntValue) IsTruthy() bool {
	return iv.Value != 0
}

// Cast converts the integer to a new type.
func (iv *IntValue) Cast(newType ValueType) (Value, error) {
	if newType == iv.GetValueType() {
		return iv, nil
	}

	switch newType {
	case ValueTypeFloat:
		return NewFloatValue(float64(iv.Value)), nil
	case ValueTypeString:
		return NewStringValue(strconv.Itoa(iv.Value)), nil
	}

	return nil, fmt.Errorf("cannot cast IntValue to %v", newType)
}

// --- FloatValue ---

// FloatValue is a value that holds a float.
type FloatValue struct {
	value[float64]
}

// NewFloatValue creates a new FloatValue.
func NewFloatValue(val float64) *FloatValue {
	fv := &FloatValue{}
	fv.Value = val
	return fv
}

// GetValueType returns the type of the value.
func (fv *FloatValue) GetValueType() ValueType {
	return ValueTypeFloat
}

// IsTruthy returns true if the float is not zero.
func (fv *FloatValue) IsTruthy() bool {
	return fv.Value != 0.0
}

// Cast converts the float to a new type.
func (fv *FloatValue) Cast(newType ValueType) (Value, error) {
	if newType == fv.GetValueType() {
		return fv, nil
	}

	switch newType {
	case ValueTypeInt:
		return NewIntValue(int(fv.Value)), nil
	case ValueTypeString:
		return NewStringValue(strconv.FormatFloat(fv.Value, 'f', -1, 64)), nil
	}

	return nil, fmt.Errorf("cannot cast FloatValue to %v", newType)
}

// --- CreateValue ---

// CreateValue is a factory function that creates a Value object from a native Go type.
func CreateValue(val any) Value {
	switch v := val.(type) {
	case int:
		return NewIntValue(v)
	case int32:
		return NewIntValue(int(v))
	case int64:
		return NewIntValue(int(v))
	case float32:
		return NewFloatValue(float64(v))
	case float64:
		return NewFloatValue(v)
	case string:
		return NewStringValue(v)
	case bool:
		if v {
			return NewIntValue(1)
		}
		return NewIntValue(0)
	case nil:
		return nil
	// TODO: Add other types: *Path, List
	default:
		// Fallback for unknown types to avoid panic
		return NewStringValue(fmt.Sprintf("%v", v))
	}
}
