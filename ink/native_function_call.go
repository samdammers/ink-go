package ink

import (
	"fmt"
)

// NativeFunctionCall represents a call to a built-in function.
type NativeFunctionCall struct {
	*BaseRuntimeObject
	Name               string
	NumberOfParameters int
}

// Native function names.
const (
	NativeFunctionCallAdd                 = "+"
	NativeFunctionCallSubtract            = "-"
	NativeFunctionCallDivide              = "/"
	NativeFunctionCallMultiply            = "*"
	NativeFunctionCallMod                 = "%"
	NativeFunctionCallNegate              = "_"
	NativeFunctionCallEqual               = "=="
	NativeFunctionCallGreater             = ">"
	NativeFunctionCallLess                = "<"
	NativeFunctionCallGreaterThanOrEquals = ">="
	NativeFunctionCallLessThanOrEquals    = "<="
	NativeFunctionCallNotEquals           = "!="
	NativeFunctionCallNot                 = "!"
	NativeFunctionCallAnd                 = "&&"
	NativeFunctionCallOr                  = "||"
	NativeFunctionCallMin                 = "MIN"
	NativeFunctionCallMax                 = "MAX"
	NativeFunctionCallPow                 = "POW"
	NativeFunctionCallFloor               = "FLOOR"
	NativeFunctionCallCeiling             = "CEILING"
	NativeFunctionCallInt                 = "INT"
	NativeFunctionCallFloat               = "FLOAT"
	NativeFunctionCallListIntersect       = "^"
	NativeFunctionCallListHas             = "?"
	NativeFunctionCallListHasnt           = "!?"
)

// NewNativeFunctionCall creates a new NativeFunctionCall.
func NewNativeFunctionCall(name string) *NativeFunctionCall {
	nf := &NativeFunctionCall{
		BaseRuntimeObject: NewBaseRuntimeObject(),
		Name:              name,
	}
	nf.NumberOfParameters = NativeFunctionCallNumberOfParameters(name)
	return nf
}

// NativeFunctionCallNumberOfParameters returns the number of parameters for a given native function name.
func NativeFunctionCallNumberOfParameters(name string) int {
	switch name {
	case NativeFunctionCallAdd, NativeFunctionCallSubtract, NativeFunctionCallMultiply, NativeFunctionCallDivide, NativeFunctionCallMod,
		NativeFunctionCallEqual, NativeFunctionCallGreater, NativeFunctionCallLess, NativeFunctionCallGreaterThanOrEquals, NativeFunctionCallLessThanOrEquals, NativeFunctionCallNotEquals,
		NativeFunctionCallAnd, NativeFunctionCallOr, NativeFunctionCallMin, NativeFunctionCallMax, NativeFunctionCallPow,
		NativeFunctionCallListIntersect, NativeFunctionCallListHas, NativeFunctionCallListHasnt:
		return 2
	case NativeFunctionCallNegate, NativeFunctionCallNot, NativeFunctionCallFloor, NativeFunctionCallCeiling, NativeFunctionCallInt, NativeFunctionCallFloat:
		return 1
	}
	return 0
}

// Call executes the native function with the given parameters.
func (n *NativeFunctionCall) Call(parameters []RuntimeObject) (RuntimeObject, error) {
	if len(parameters) != n.NumberOfParameters {
		return nil, fmt.Errorf("unexpected number of parameters")
	}

	switch n.Name {
	case NativeFunctionCallAdd:
		return n.add(parameters[0], parameters[1])
	case NativeFunctionCallSubtract:
		return n.subtract(parameters[0], parameters[1])
	case NativeFunctionCallMultiply:
		return n.multiply(parameters[0], parameters[1])
	case NativeFunctionCallDivide:
		return n.divide(parameters[0], parameters[1])
	case NativeFunctionCallMod:
		return n.mod(parameters[0], parameters[1])
	case NativeFunctionCallEqual:
		return n.equal(parameters[0], parameters[1])
	case NativeFunctionCallGreater:
		return n.greater(parameters[0], parameters[1])
	case NativeFunctionCallLess:
		return n.less(parameters[0], parameters[1])
	case NativeFunctionCallNot:
		return n.not(parameters[0])
	case NativeFunctionCallListIntersect:
		if l1, ok1 := parameters[0].(*ListValue); ok1 {
			if l2, ok2 := parameters[1].(*ListValue); ok2 {
				return NewListValue(l1.Value.Intersect(l2.Value)), nil
			}
		}
		return nil, fmt.Errorf("cannot intersect %T and %T", parameters[0], parameters[1])
	case NativeFunctionCallListHas:
		if l1, ok1 := parameters[0].(*ListValue); ok1 {
			if l2, ok2 := parameters[1].(*ListValue); ok2 {
				return NewBoolValue(l1.Value.Has(l2.Value)), nil
			}
		}
		return nil, fmt.Errorf("cannot check Has on %T and %T", parameters[0], parameters[1])
	case NativeFunctionCallListHasnt:
		if l1, ok1 := parameters[0].(*ListValue); ok1 {
			if l2, ok2 := parameters[1].(*ListValue); ok2 {
				return NewBoolValue(!l1.Value.Has(l2.Value)), nil
			}
		}
		return nil, fmt.Errorf("cannot check Hasnt on %T and %T", parameters[0], parameters[1])
	}
	// TODO: Implement other operations
	return nil, fmt.Errorf("operation not implemented: %s", n.Name)
}

func (n *NativeFunctionCall) coerceValues(v1, v2 RuntimeObject) (Value, Value, error) {
	val1, ok1 := v1.(Value)
	val2, ok2 := v2.(Value)
	if !ok1 || !ok2 {
		return nil, nil, fmt.Errorf("operands are not values")
	}
	// TODO: Better type coercion (e.g. Int -> Float if one is Float)
	return val1, val2, nil
}

func (n *NativeFunctionCall) add(v1, v2 RuntimeObject) (RuntimeObject, error) {
	// Special Case: String concatenation
	// If either is string, convert both to string
	s1, isStr1 := v1.(*StringValue)
	s2, isStr2 := v2.(*StringValue)

	if isStr1 || isStr2 {
		str1 := ""
		str2 := ""
		if isStr1 {
			str1 = s1.Value
		} else {
			str1 = fmt.Sprintf("%v", v1)
		} // Simplified stringify
		if isStr2 {
			str2 = s2.Value
		} else {
			str2 = fmt.Sprintf("%v", v2)
		}

		// In full Ink, non-string values need proper string representation (IntValue.ToString, etc)
		// For now, assuming basic values work with Sprintf, or we rely on Values implementing String()

		// Wait, Values should have String() ?
		// If v1 is IntValue, we can cast.

		if !isStr1 {
			if i, ok := v1.(*IntValue); ok {
				str1 = fmt.Sprintf("%d", i.Value)
			} else if f, ok := v1.(*FloatValue); ok {
				str1 = fmt.Sprintf("%v", f.Value) // %v avoids trailing zeros sometimes better than %f
			} else {
				// Fallback
				str1 = "" // or panic/error?
			}
		}
		if !isStr2 {
			if i, ok := v2.(*IntValue); ok {
				str2 = fmt.Sprintf("%d", i.Value)
			} else if f, ok := v2.(*FloatValue); ok {
				str2 = fmt.Sprintf("%v", f.Value)
			} else {
				str2 = ""
			}
		}

		return NewStringValue(str1 + str2), nil
	}

	// List Addition (Union)
	l1, isList1 := v1.(*ListValue)
	l2, isList2 := v2.(*ListValue)
	if isList1 && isList2 {
		return NewListValue(l1.Value.Union(l2.Value)), nil
	}

	// Numeric Addition
	val1, val2, err := n.coerceValues(v1, v2)
	if err != nil {
		return nil, err
	}

	i1, isInt1 := val1.(*IntValue)
	i2, isInt2 := val2.(*IntValue)

	if isInt1 && isInt2 {
		return NewIntValue(i1.Value + i2.Value), nil
	}

	f1, isFloat1 := val1.(*FloatValue)
	f2, isFloat2 := val2.(*FloatValue)

	if isFloat1 && isFloat2 {
		return NewFloatValue(f1.Value + f2.Value), nil
	}
	if isInt1 && isFloat2 {
		return NewFloatValue(float64(i1.Value) + f2.Value), nil
	}
	if isFloat1 && isInt2 {
		return NewFloatValue(f1.Value + float64(i2.Value)), nil
	}

	return nil, fmt.Errorf("cannot add %T and %T", v1, v2)
}

func (n *NativeFunctionCall) subtract(v1, v2 RuntimeObject) (RuntimeObject, error) {
	// List Subtraction
	l1, isList1 := v1.(*ListValue)
	l2, isList2 := v2.(*ListValue)
	if isList1 && isList2 {
		return NewListValue(l1.Value.Subtract(l2.Value)), nil
	}

	val1, val2, err := n.coerceValues(v1, v2)
	if err != nil {
		return nil, err
	}

	i1, isInt1 := val1.(*IntValue)
	i2, isInt2 := val2.(*IntValue)

	if isInt1 && isInt2 {
		return NewIntValue(i1.Value - i2.Value), nil
	}

	f1, isFloat1 := val1.(*FloatValue)
	f2, isFloat2 := val2.(*FloatValue)

	if isFloat1 && isFloat2 {
		return NewFloatValue(f1.Value - f2.Value), nil
	}
	if isInt1 && isFloat2 {
		return NewFloatValue(float64(i1.Value) - f2.Value), nil
	}
	if isFloat1 && isInt2 {
		return NewFloatValue(f1.Value - float64(i2.Value)), nil
	}

	return nil, fmt.Errorf("cannot subtract %T and %T", v1, v2)
}

func (n *NativeFunctionCall) multiply(v1, v2 RuntimeObject) (RuntimeObject, error) {
	val1, val2, err := n.coerceValues(v1, v2)
	if err != nil {
		return nil, err
	}

	i1, isInt1 := val1.(*IntValue)
	i2, isInt2 := val2.(*IntValue)

	if isInt1 && isInt2 {
		return NewIntValue(i1.Value * i2.Value), nil
	}

	f1, isFloat1 := val1.(*FloatValue)
	f2, isFloat2 := val2.(*FloatValue)

	if isFloat1 && isFloat2 {
		return NewFloatValue(f1.Value * f2.Value), nil
	}
	if isInt1 && isFloat2 {
		return NewFloatValue(float64(i1.Value) * f2.Value), nil
	}
	if isFloat1 && isInt2 {
		return NewFloatValue(f1.Value * float64(i2.Value)), nil
	}

	return nil, fmt.Errorf("cannot multiply %T and %T", v1, v2)
}

func (n *NativeFunctionCall) divide(v1, v2 RuntimeObject) (RuntimeObject, error) {
	val1, val2, err := n.coerceValues(v1, v2)
	if err != nil {
		return nil, err
	}

	i1, isInt1 := val1.(*IntValue)
	i2, isInt2 := val2.(*IntValue)

	if isInt1 && isInt2 {
		if i2.Value == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return NewIntValue(i1.Value / i2.Value), nil // Floor division
	}

	f1, isFloat1 := val1.(*FloatValue)
	f2, isFloat2 := val2.(*FloatValue)

	// Promote to float
	v1Val := float64(0)
	v2Val := float64(0)
	if isInt1 {
		v1Val = float64(i1.Value)
	} else if isFloat1 {
		v1Val = f1.Value
	}
	if isInt2 {
		v2Val = float64(i2.Value)
	} else if isFloat2 {
		v2Val = f2.Value
	}

	if v2Val == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	return NewFloatValue(v1Val / v2Val), nil
}

func (n *NativeFunctionCall) mod(v1, v2 RuntimeObject) (RuntimeObject, error) {
	val1, val2, err := n.coerceValues(v1, v2)
	if err != nil {
		return nil, err
	}

	i1, isInt1 := val1.(*IntValue)
	i2, isInt2 := val2.(*IntValue)

	if isInt1 && isInt2 {
		if i2.Value == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		return NewIntValue(i1.Value % i2.Value), nil
	}
	return nil, fmt.Errorf("modulo only supported for ints")
}

func (n *NativeFunctionCall) equal(v1, v2 RuntimeObject) (RuntimeObject, error) {
	// Null logic
	if v1 == nil && v2 == nil {
		return NewIntValue(1), nil
	} // true
	if v1 == nil || v2 == nil {
		return NewIntValue(0), nil
	} // false

	val1, ok1 := v1.(Value)
	val2, ok2 := v2.(Value)

	if !ok1 || !ok2 {
		return NewIntValue(0), nil
	} // Comparison of different object types?

	// Types must match? Ink allows 5 == 5.0 -> True
	// TODO: Full equality logic

	// Simple int/float check
	i1, isInt1 := val1.(*IntValue)
	i2, isInt2 := val2.(*IntValue)
	if isInt1 && isInt2 {
		if i1.Value == i2.Value {
			return NewIntValue(1), nil
		}
		return NewIntValue(0), nil
	}

	f1, isFloat1 := val1.(*FloatValue)
	f2, isFloat2 := val2.(*FloatValue)

	if isFloat1 && isFloat2 {
		// Use epsilon? For now direct equality
		if f1.Value == f2.Value {
			return NewIntValue(1), nil
		}
		return NewIntValue(0), nil
	}

	// Mixed Int/Float
	if isInt1 && isFloat2 {
		if float64(i1.Value) == f2.Value {
			return NewIntValue(1), nil
		}
		return NewIntValue(0), nil
	}

	if isFloat1 && isInt2 {
		if f1.Value == float64(i2.Value) {
			return NewIntValue(1), nil
		}
		return NewIntValue(0), nil
	}

	// String Equality
	s1, isStr1 := val1.(*StringValue)
	s2, isStr2 := val2.(*StringValue)
	if isStr1 && isStr2 {
		if s1.Value == s2.Value {
			return NewIntValue(1), nil
		}
		return NewIntValue(0), nil
	}

	// TODO: DivertTarget, etc

	return NewIntValue(0), nil // Default false
}

func (n *NativeFunctionCall) greater(v1, v2 RuntimeObject) (RuntimeObject, error) {
	val1, val2, err := n.coerceValues(v1, v2)
	if err != nil {
		return nil, err
	}

	i1, isInt1 := val1.(*IntValue)
	i2, isInt2 := val2.(*IntValue)

	if isInt1 && isInt2 {
		if i1.Value > i2.Value {
			return NewIntValue(1), nil
		}
		return NewIntValue(0), nil
	}

	// Float etc
	return NewIntValue(0), nil
}

func (n *NativeFunctionCall) less(v1, v2 RuntimeObject) (RuntimeObject, error) {
	val1, val2, err := n.coerceValues(v1, v2)
	if err != nil {
		return nil, err
	}

	i1, isInt1 := val1.(*IntValue)
	i2, isInt2 := val2.(*IntValue)

	if isInt1 && isInt2 {
		if i1.Value < i2.Value {
			return NewIntValue(1), nil
		}
		return NewIntValue(0), nil
	}
	return NewIntValue(0), nil
}

func (n *NativeFunctionCall) not(v1 RuntimeObject) (RuntimeObject, error) {
	val, ok := v1.(Value)
	if !ok {
		return nil, fmt.Errorf("not a value")
	}

	if i, ok := val.(*IntValue); ok {
		if i.Value == 0 {
			return NewIntValue(1), nil
		}
		return NewIntValue(0), nil
	}

	// Default truthy check
	// Ink logic: 0 is false, everything else true?
	// Actually IsTruthy is on Value interface.

	// TODO: Use IsTruthy
	return NewIntValue(0), nil
}
