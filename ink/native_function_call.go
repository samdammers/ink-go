package ink

import (
	"fmt"
)

// NativeFunctionCall represents a call to built-in function.
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

// NativeFunctionCallNumberOfParameters returns the number of parameters.
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

// Call executes the native function.
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
	return nil, fmt.Errorf("operation not implemented: %s", n.Name)
}

func (n *NativeFunctionCall) coerceValues(v1, v2 RuntimeObject) (Value, Value, error) {
	val1, ok1 := v1.(Value)
	val2, ok2 := v2.(Value)
	if !ok1 || !ok2 {
		return nil, nil, fmt.Errorf("operands are not values")
	}
	return val1, val2, nil
}

func (n *NativeFunctionCall) coerceToString(v RuntimeObject) string {
	if s, ok := v.(*StringValue); ok {
		return s.Value
	}
	if i, ok := v.(*IntValue); ok {
		return fmt.Sprintf("%d", i.Value)
	}
	if f, ok := v.(*FloatValue); ok {
		return fmt.Sprintf("%v", f.Value)
	}
	return ""
}

func (n *NativeFunctionCall) performBinaryNumericOp(v1, v2 RuntimeObject, intOp func(int, int) int, floatOp func(float64, float64) float64) (RuntimeObject, error) {
	val1, val2, err := n.coerceValues(v1, v2)
	if err != nil {
		return nil, err
	}

	i1, isInt1 := val1.(*IntValue)
	i2, isInt2 := val2.(*IntValue)

	if isInt1 && isInt2 {
		return NewIntValue(intOp(i1.Value, i2.Value)), nil
	}

	f1, isFloat1 := val1.(*FloatValue)
	f2, isFloat2 := val2.(*FloatValue)

	if isFloat1 && isFloat2 {
		return NewFloatValue(floatOp(f1.Value, f2.Value)), nil
	}
	if isInt1 && isFloat2 {
		return NewFloatValue(floatOp(float64(i1.Value), f2.Value)), nil
	}
	if isFloat1 && isInt2 {
		return NewFloatValue(floatOp(f1.Value, float64(i2.Value))), nil
	}

	return nil, fmt.Errorf("cannot perform operation on %T and %T", v1, v2)
}

func (n *NativeFunctionCall) add(v1, v2 RuntimeObject) (RuntimeObject, error) {
	_, isStr1 := v1.(*StringValue)
	_, isStr2 := v2.(*StringValue)

	if isStr1 || isStr2 {
		str1 := n.coerceToString(v1)
		str2 := n.coerceToString(v2)
		return NewStringValue(str1 + str2), nil
	}

	l1, isList1 := v1.(*ListValue)
	l2, isList2 := v2.(*ListValue)
	if isList1 && isList2 {
		return NewListValue(l1.Value.Union(l2.Value)), nil
	}

	return n.addNumbers(v1, v2)
}

func (n *NativeFunctionCall) addNumbers(v1, v2 RuntimeObject) (RuntimeObject, error) {
	return n.performBinaryNumericOp(v1, v2,
		func(a, b int) int { return a + b },
		func(a, b float64) float64 { return a + b },
	)
}

func (n *NativeFunctionCall) subtract(v1, v2 RuntimeObject) (RuntimeObject, error) {
	l1, isList1 := v1.(*ListValue)
	l2, isList2 := v2.(*ListValue)
	if isList1 && isList2 {
		return NewListValue(l1.Value.Subtract(l2.Value)), nil
	}
	return n.performBinaryNumericOp(v1, v2,
		func(a, b int) int { return a - b },
		func(a, b float64) float64 { return a - b },
	)
}

func (n *NativeFunctionCall) multiply(v1, v2 RuntimeObject) (RuntimeObject, error) {
	return n.performBinaryNumericOp(v1, v2,
		func(a, b int) int { return a * b },
		func(a, b float64) float64 { return a * b },
	)
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
		return NewIntValue(i1.Value / i2.Value), nil
	}

	f1, isFloat1 := val1.(*FloatValue)
	f2, isFloat2 := val2.(*FloatValue)

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
	return nil, fmt.Errorf("modulo not supported for non-integers")
}

func (n *NativeFunctionCall) equalNumbers(v1, v2 RuntimeObject) (int, bool) {
	i1, isInt1 := v1.(*IntValue)
	i2, isInt2 := v2.(*IntValue)
	if isInt1 && isInt2 {
		if i1.Value == i2.Value {
			return 1, true
		}
		return 0, true
	}

	f1, isFloat1 := v1.(*FloatValue)
	f2, isFloat2 := v2.(*FloatValue)
	if isFloat1 && isFloat2 {
		if f1.Value == f2.Value {
			return 1, true
		}
		return 0, true
	}

	if isInt1 && isFloat2 {
		if float64(i1.Value) == f2.Value {
			return 1, true
		}
		return 0, true
	}
	if isFloat1 && isInt2 {
		if f1.Value == float64(i2.Value) {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

func (n *NativeFunctionCall) equal(v1, v2 RuntimeObject) (RuntimeObject, error) {
	if v1 == nil && v2 == nil {
		return NewIntValue(1), nil
	}
	if v1 == nil || v2 == nil {
		return NewIntValue(0), nil
	}

	val1, ok1 := v1.(Value)
	val2, ok2 := v2.(Value)

	if !ok1 || !ok2 {
		return NewIntValue(0), nil
	}

	if eq, ok := n.equalNumbers(val1, val2); ok {
		return NewIntValue(eq), nil
	}

	s1, isStr1 := val1.(*StringValue)
	s2, isStr2 := val2.(*StringValue)
	if isStr1 && isStr2 {
		if s1.Value == s2.Value {
			return NewIntValue(1), nil
		}
		return NewIntValue(0), nil
	}

	return NewIntValue(0), nil
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
		return nil, fmt.Errorf("operand is not a value")
	}
	if val.IsTruthy() {
		return NewIntValue(0), nil
	}
	return NewIntValue(1), nil
}
