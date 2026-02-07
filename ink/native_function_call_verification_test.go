package ink

import (
	"testing"
)

func TestNativeFunctionMathVerification(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		args     []RuntimeObject
		want     interface{} // int or float64 or string error substring
		wantErr  bool
	}{
		// Add
		{"Add Ints", "+", []RuntimeObject{NewIntValue(5), NewIntValue(3)}, 8, false},
		{"Add Floats", "+", []RuntimeObject{NewFloatValue(5.5), NewFloatValue(2.5)}, 8.0, false},
		{"Add Mixed", "+", []RuntimeObject{NewIntValue(5), NewFloatValue(2.5)}, 7.5, false},

		// Subtract
		{"Sub Ints", "-", []RuntimeObject{NewIntValue(10), NewIntValue(3)}, 7, false},
		{"Sub Floats", "-", []RuntimeObject{NewFloatValue(10.5), NewFloatValue(2.5)}, 8.0, false},
		{"Sub Mixed", "-", []RuntimeObject{NewIntValue(10), NewFloatValue(2.5)}, 7.5, false},

		// Multiply
		{"Mul Ints", "*", []RuntimeObject{NewIntValue(4), NewIntValue(3)}, 12, false},
		{"Mul Floats", "*", []RuntimeObject{NewFloatValue(2.5), NewFloatValue(4.0)}, 10.0, false},
		{"Mul Mixed", "*", []RuntimeObject{NewIntValue(4), NewFloatValue(1.5)}, 6.0, false},

		// Divide
		{"Div Ints Exact", "/", []RuntimeObject{NewIntValue(10), NewIntValue(2)}, 5, false},
		{"Div Ints trunc", "/", []RuntimeObject{NewIntValue(10), NewIntValue(3)}, 3, false},
		{"Div Floats", "/", []RuntimeObject{NewFloatValue(10.0), NewFloatValue(2.0)}, 5.0, false},
		{"Div Mixed", "/", []RuntimeObject{NewIntValue(10), NewFloatValue(2.0)}, 5.0, false},
		{"Div Zero", "/", []RuntimeObject{NewIntValue(10), NewIntValue(0)}, "division by zero", true},

		// Mod
		{"Mod Ints", "%", []RuntimeObject{NewIntValue(10), NewIntValue(3)}, 1, false},
		{"Mod Zero", "%", []RuntimeObject{NewIntValue(10), NewIntValue(0)}, "modulo by zero", true},

		// Equality
		{"Eq Ints", "==", []RuntimeObject{NewIntValue(5), NewIntValue(5)}, 1, false},
		{"Neq Ints", "==", []RuntimeObject{NewIntValue(5), NewIntValue(6)}, 0, false},
		{"Eq Mixed", "==", []RuntimeObject{NewIntValue(5), NewFloatValue(5.0)}, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nf := NewNativeFunctionCall(tt.funcName)
			got, err := nf.Call(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.want)
				}
				// check error string if needed
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			switch v := got.(type) {
			case *IntValue:
				if wantInt, ok := tt.want.(int); ok {
					if v.Value != wantInt {
						t.Errorf("got %d, want %d", v.Value, wantInt)
					}
				} else {
					t.Errorf("got IntValue, want %T", tt.want)
				}
			case *FloatValue:
				if wantFloat, ok := tt.want.(float64); ok {
					if v.Value != wantFloat {
						t.Errorf("got %f, want %f", v.Value, wantFloat)
					}
				} else {
					t.Errorf("got FloatValue, want %T", tt.want)
				}
			default:
				t.Errorf("unexpected return type: %T", got)
			}
		})
	}
}
