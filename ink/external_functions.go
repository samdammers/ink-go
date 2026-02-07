package ink

import (
	"fmt"
)

// BindExternalFunction binds a go function to the story.
func (s *Story) BindExternalFunction(name string, f ExternalFunction) error {
	if _, ok := s.externalFunctions[name]; ok {
		return fmt.Errorf("function '%s' is already bound", name)
	}
	s.externalFunctions[name] = f
	return nil
}

// UnbindExternalFunction unbinds an external function.
func (s *Story) UnbindExternalFunction(name string) {
	delete(s.externalFunctions, name)
}

// callExternalFunction calls a bound external function.
func (s *Story) callExternalFunction(name string, numberOfArgs int) error {
	f, ok := s.externalFunctions[name]
	if !ok {
		// If function isn't found, we might fallback?
		// In Ink, often there's a fallback function in the ink itself logic.
		// For now simple error or try fallback logic if implemented.
		return fmt.Errorf("external function '%s' not found", name)
	}

	// Pop arguments
	args := make([]any, numberOfArgs)
	for i := numberOfArgs - 1; i >= 0; i-- {
		obj := s.state.PopEvaluationStack()
		val, err := RuntimeObjectToNative(obj)
		if err != nil {
			return fmt.Errorf("failed to convert argument %d for function '%s': %v", i, name, err)
		}
		args[i] = val
	}

	// Call
	ret, err := f(args)
	if err != nil {
		return fmt.Errorf("error executing external function '%s': %v", name, err)
	}

	// Push result
	if ret != nil {
		rtObj, err := NativeToRuntimeObject(ret)
		if err != nil {
			return fmt.Errorf("failed to convert return value from function '%s': %v", name, err)
		}
		if rtObj != nil {
			s.state.PushEvaluationStack(rtObj)
		}
	}

	return nil
}

// RuntimeObjectToNative converts a RuntimeObject to a native Go value.
func RuntimeObjectToNative(obj RuntimeObject) (any, error) {
	switch v := obj.(type) {
	case *IntValue:
		return v.Value, nil
	case *FloatValue:
		return v.Value, nil
	case *StringValue:
		return v.Value, nil
	// case *BoolValue: // If implemented
	//	return v.Value, nil
	case *Void:
		return nil, nil
	default:
		return nil, fmt.Errorf("cannot convert runtime object of type %T to native", obj)
	}
}

// NativeToRuntimeObject converts a native Go value to a RuntimeObject.
func NativeToRuntimeObject(val any) (RuntimeObject, error) {
	switch v := val.(type) {
	case int:
		return NewIntValue(v), nil
	case int32:
		return NewIntValue(int(v)), nil
	case int64:
		return NewIntValue(int(v)), nil
	case float64:
		return NewFloatValue(v), nil
	case float32:
		return NewFloatValue(float64(v)), nil
	case string:
		return NewStringValue(v), nil
	case bool:
		// Assuming we implement BoolValue soon or re-use IntValue(1/0)
		if v {
			return NewIntValue(1), nil
		}
		return NewIntValue(0), nil
	case nil:
		return nil, nil
	}
	return nil, fmt.Errorf("cannot convert native value of type %T to runtime object", val)
}
