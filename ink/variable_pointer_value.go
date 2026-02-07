package ink

import "fmt"

// VariablePointerValue represents a reference to a variable.
type VariablePointerValue struct {
	*BaseRuntimeObject
	variableName string
	contextIndex int // -1 = unknown, 0 = global, 1+ = stack frame
}

// NewVariablePointerValue creates a new VariablePointerValue.
func NewVariablePointerValue(variableName string, contextIndex int) *VariablePointerValue {
	if contextIndex == 0 {
		contextIndex = -1 // Default to unknown/generic if 0 passed? Java default is -1.
		// Wait, Java constructor:
		// public VariablePointerValue(String variableName) { this(variableName, -1); }
		// public VariablePointerValue(String variableName, int contextIndex) { ... }
	}
	return &VariablePointerValue{
		BaseRuntimeObject: NewBaseRuntimeObject(),
		variableName:      variableName,
		contextIndex:      contextIndex,
	}
}

// VariableName returns the name of the variable this pointer references.
func (v *VariablePointerValue) VariableName() string {
	return v.variableName
}

// SetVariableName sets the name of the variable this pointer references.
func (v *VariablePointerValue) SetVariableName(name string) {
	v.variableName = name
}

// ContextIndex returns the context index of the variable (e.g. stack frame).
func (v *VariablePointerValue) ContextIndex() int {
	return v.contextIndex
}

// SetContextIndex sets the context index.
func (v *VariablePointerValue) SetContextIndex(idx int) {
	v.contextIndex = idx
}

// Implement Value interface

// GetValueType returns the type of the value (ValueTypeVariablePointer).
func (v *VariablePointerValue) GetValueType() ValueType {
	return ValueTypeVariablePointer
}

// IsTruthy returns true, as variable pointers are considered truthy in Ink.
func (v *VariablePointerValue) IsTruthy() bool {
	// Variable pointers are usually truthy unless maybe the name is empty?
	// Java doesn't override, so it uses basic object truthiness (true).
	// Exception: int(0), float(0.0), bool(false), empty string?
	// A pointer to a variable should probably be truthy.
	return true
}

// Cast returns the value cast to the specified type.
func (v *VariablePointerValue) Cast(newType ValueType) (Value, error) {
	if newType == v.GetValueType() {
		return v, nil
	}
	// TODO: Can we cast a variable pointer to something else?
	// Java says: throw BadCastException(newType);
	return nil, fmt.Errorf("cannot cast VariablePointerValue to type %v", newType)
}

// GetValueObject returns the value of the object.
func (v *VariablePointerValue) GetValueObject() any {
	return v.variableName
}

func (v *VariablePointerValue) String() string {
	return fmt.Sprintf("VariablePointerValue(%s)", v.variableName)
}

// Copy creates a deep copy.
func (v *VariablePointerValue) Copy() RuntimeObject {
	return NewVariablePointerValue(v.variableName, v.contextIndex)
}
