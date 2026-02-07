package ink

import (
	"math"
	"testing"
)

func TestValueCreation(t *testing.T) {
	tests := []struct {
		name     string
		val      Value
		expected any
	}{
		{"String", NewStringValue("hello"), "hello"},
		{"Int", NewIntValue(123), 123},
		{"Float", NewFloatValue(123.456), 123.456},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.val.GetValueObject() != tt.expected {
				t.Errorf("got %v, want %v", tt.val.GetValueObject(), tt.expected)
			}
		})
	}
}

func TestValueCasting(t *testing.T) {
	tests := []struct {
		name        string
		startVal    Value
		castTo      ValueType
		expectedVal any
		expectErr   bool
	}{
		// String casting
		{"StringToInt", NewStringValue("123"), ValueTypeInt, 123, false},
		{"StringToFloat", NewStringValue("123.45"), ValueTypeFloat, 123.45, false},
		{"StringToString", NewStringValue("hello"), ValueTypeString, "hello", false},
		{"StringToInvalidInt", NewStringValue("hello"), ValueTypeInt, nil, false}, // Java returns null

		// Int casting
		{"IntToString", NewIntValue(456), ValueTypeString, "456", false},
		{"IntToFloat", NewIntValue(456), ValueTypeFloat, float64(456), false},
		{"IntToInt", NewIntValue(456), ValueTypeInt, 456, false},
		{"IntToInvalid", NewIntValue(456), ValueTypeString + 100, nil, true},

		// Float casting
		{"FloatToString", NewFloatValue(78.9), ValueTypeString, "78.9", false},
		{"FloatToInt", NewFloatValue(78.9), ValueTypeInt, 78, false},
		{"FloatToFloat", NewFloatValue(78.9), ValueTypeFloat, 78.9, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			castVal, err := tt.startVal.Cast(tt.castTo)

			if (err != nil) != tt.expectErr {
				t.Fatalf("Cast() error = %v, expectErr %v", err, tt.expectErr)
			}

			if err != nil {
				return
			}

			if castVal == nil && tt.expectedVal == nil {
				return // Success
			}

			if castVal == nil || tt.expectedVal == nil {
				t.Fatalf("got val %v, want %v", castVal, tt.expectedVal)
			}

			// Special case for float comparison
			if fv, ok := tt.expectedVal.(float64); ok {
				gotFv := castVal.GetValueObject().(float64)
				if math.Abs(fv-gotFv) > 1e-6 {
					t.Errorf("got %v, want %v", gotFv, fv)
				}
				return
			}

			if castVal.GetValueObject() != tt.expectedVal {
				t.Errorf("got %v, want %v", castVal.GetValueObject(), tt.expectedVal)
			}
		})
	}
}

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		name     string
		val      Value
		expected bool
	}{
		{"StringTrue", NewStringValue("hello"), true},
		{"StringFalse", NewStringValue(""), false},
		{"IntTrue", NewIntValue(123), true},
		{"IntFalse", NewIntValue(0), false},
		{"FloatTrue", NewFloatValue(0.1), true},
		{"FloatFalse", NewFloatValue(0.0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.val.IsTruthy() != tt.expected {
				t.Errorf("got %v, want %v", tt.val.IsTruthy(), tt.expected)
			}
		})
	}
}

func TestCreateValue(t *testing.T) {
	tests := []struct {
		name         string
		input        any
		expectedType ValueType
		expectedVal  any
	}{
		{"FromInt", 123, ValueTypeInt, 123},
		{"FromInt32", int32(456), ValueTypeInt, 456},
		{"FromFloat32", float32(1.23), ValueTypeFloat, float64(1.23)},
		{"FromFloat64", 9.87, ValueTypeFloat, 9.87},
		{"FromString", "hello world", ValueTypeString, "hello world"},
		{"FromNil", nil, ValueTypeNoType, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := CreateValue(tt.input)

			if tt.expectedVal == nil {
				if val != nil {
					t.Fatalf("Expected nil value, but got %v", val)
				}
				return
			}

			if val.GetValueType() != tt.expectedType {
				t.Errorf("got type %v, want %v", val.GetValueType(), tt.expectedType)
			}

			// Special case for float comparison
			if fv, ok := tt.expectedVal.(float64); ok {
				gotFv := val.GetValueObject().(float64)
				if math.Abs(fv-gotFv) > 1e-6 {
					t.Errorf("got %v, want %v", gotFv, fv)
				}
				return
			}

			if val.GetValueObject() != tt.expectedVal {
				t.Errorf("got val %v, want %v", val.GetValueObject(), tt.expectedVal)
			}
		})
	}
}

func TestStackMath(t *testing.T) {
	stack := make([]Value, 0)

	// Push 3, then 2
	stack = append(stack, NewIntValue(3))
	stack = append(stack, NewIntValue(2))

	// Pop them off
	b := stack[len(stack)-1]
	stack = stack[:len(stack)-1]
	a := stack[len(stack)-1]

	// Check types and values
	bVal, ok := b.GetValueObject().(int)
	if !ok || bVal != 2 {
		t.Fatalf("First pop should have been 2, but got %v", b.GetValueObject())
	}

	aVal, ok := a.GetValueObject().(int)
	if !ok || aVal != 3 {
		t.Fatalf("Second pop should have been 3, but got %v", a.GetValueObject())
	}

	// Confirm subtraction order
	result := aVal - bVal
	if result != 1 {
		t.Errorf("Expected 3 - 2 = 1, but got %d", result)
	}
}
