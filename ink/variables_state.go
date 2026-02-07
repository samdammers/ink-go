package ink

// VariableChangedFunc is a callback for variable changes.
type VariableChangedFunc func(variableName string, newValue RuntimeObject)

// VariablesState encompasses all the global variables in an ink Story.
type VariablesState struct {
	GlobalVariables        map[string]RuntimeObject
	DefaultGlobalVariables map[string]RuntimeObject
	CallStack              *CallStack
	ListDefsOrigin         *ListDefinitionsOrigin
	Patch                  *StatePatch

	batchObservingVariableChanges bool
	changedVariablesForBatchObs   map[string]struct{}
	variableChangedEvent          VariableChangedFunc
}

// NewVariablesState creates a new VariablesState.
func NewVariablesState(callStack *CallStack, listDefsOrigin *ListDefinitionsOrigin) *VariablesState {
	return &VariablesState{
		GlobalVariables: make(map[string]RuntimeObject),
		CallStack:       callStack,
		ListDefsOrigin:  listDefsOrigin,
	}
}

// SetCallStack sets the call stack.
func (vs *VariablesState) SetCallStack(callStack *CallStack) {
	vs.CallStack = callStack
}

// Assign assigns a value to a variable, handling global/temporary and pointers.
func (vs *VariablesState) Assign(varAss *VariableAssignment, value RuntimeObject) error {
	name := varAss.VariableName()
	contextIndex := -1
	setGlobal := false

	if varAss.IsNewDeclaration() {
		setGlobal = varAss.IsGlobal()
	} else {
		setGlobal = vs.GlobalVariableExistsWithName(name)

		// Fallback for loose JSON/Testing:
		// If variable doesn't exist, but assignment implies Key "VAR=" (Global), create it as global.
		if !setGlobal && varAss.IsGlobal() {
			// Check if it exists as local?
			// For now, assume global preference if VAR= is used.
			setGlobal = true
		}
	}

	// Constructing new variable pointer reference
	if varAss.IsNewDeclaration() {
		if varPtr, ok := value.(*VariablePointerValue); ok {
			fullyResolved, err := vs.ResolveVariablePointer(varPtr)
			if err != nil {
				return err
			}
			value = fullyResolved
		}
	} else {
		// Assign to existing variable pointer?
		var existingPointer *VariablePointerValue
		for {
			obj := vs.GetRawVariableWithName(name, contextIndex)
			existingPointer, _ = obj.(*VariablePointerValue)

			if existingPointer != nil {
				name = existingPointer.VariableName()
				contextIndex = existingPointer.ContextIndex()
				setGlobal = (contextIndex == 0)
			} else {
				break
			}
			// Safe check for infinite loop?
		}
	}

	if setGlobal {
		vs.SetGlobal(name, value)
	} else {
		return vs.CallStack.SetTemporaryVariable(name, value, varAss.IsNewDeclaration(), contextIndex)
	}
	return nil
}

// SetGlobal sets a global variable.
func (vs *VariablesState) SetGlobal(name string, value RuntimeObject) {
	oldValue, exists := vs.GlobalVariables[name]

	// TODO: ListValue logic (retain list origin)

	vs.GlobalVariables[name] = value

	if exists && oldValue != value { // TODO: Value equality check?
		if vs.variableChangedEvent != nil {
			vs.variableChangedEvent(name, value)
		}
		if vs.batchObservingVariableChanges {
			if vs.changedVariablesForBatchObs != nil {
				vs.changedVariablesForBatchObs[name] = struct{}{}
			}
		}
	}
}

// GetVariableWithName gets a variable value.
func (vs *VariablesState) GetVariableWithName(name string) RuntimeObject {
	return vs.GetVariableWithNameContext(name, -1)
}

// GetVariableWithNameContext gets a variable value with a specific callstack context.
func (vs *VariablesState) GetVariableWithNameContext(name string, contextIndex int) RuntimeObject {
	varValue := vs.GetRawVariableWithName(name, contextIndex)

	if varPtr, ok := varValue.(*VariablePointerValue); ok {
		resolved, _ := vs.ResolveVariablePointer(varPtr)
		return resolved
	}

	return varValue
}

// GetRawVariableWithName gets the raw object (potentially a pointer).
func (vs *VariablesState) GetRawVariableWithName(name string, contextIndex int) RuntimeObject {
	// 1. Context / Temp variables (Local Scope)
	varValue := vs.CallStack.GetTemporaryVariableWithName(name, contextIndex)
	if varValue != nil {
		return varValue
	}

	// 2. Globals (if allowed)
	if contextIndex == 0 || contextIndex == -1 {
		if vs.Patch != nil {
			if val, ok := vs.Patch.GetGlobals()[name]; ok {
				return val
			}
		}

		if val, ok := vs.GlobalVariables[name]; ok {
			return val
		}

		// if vs.ListDefsOrigin != nil {
		// TODO: List definitions access
		// }
	}

	return nil
}

// GlobalVariableExistsWithName checks if a global variable exists.
func (vs *VariablesState) GlobalVariableExistsWithName(name string) bool {
	_, ok := vs.GlobalVariables[name]
	// TODO: Check patches?
	return ok
}

// ResolveVariablePointer resolves a pointer to the target value.
func (vs *VariablesState) ResolveVariablePointer(varPtr *VariablePointerValue) (*VariablePointerValue, error) {
	contextIndex := varPtr.ContextIndex()

	if contextIndex == -1 {
		contextIndex = vs.GetContextIndexOfVariableNamed(varPtr.VariableName())
	}

	value := vs.GetRawVariableWithName(varPtr.VariableName(), contextIndex)

	if nextPtr, ok := value.(*VariablePointerValue); ok {
		return vs.ResolveVariablePointer(nextPtr)
	}

	// Ensure we store the context index
	return NewVariablePointerValue(varPtr.VariableName(), contextIndex), nil
}

// GetContextIndexOfVariableNamed returns the context index (0 for global, 1+ for stack).
func (vs *VariablesState) GetContextIndexOfVariableNamed(name string) int {
	if vs.GlobalVariableExistsWithName(name) {
		return 0
	}
	return vs.CallStack.ContextForVariableNamed(name)
}

// Copy creates a deep copy.
func (vs *VariablesState) Copy(newCallStack *CallStack) *VariablesState {
	cp := NewVariablesState(newCallStack, vs.ListDefsOrigin)
	for k, v := range vs.GlobalVariables {
		// TODO: Deep copy runtime object? Value.Copy()?
		// Assuming immutable or copy method exists
		// In Go-Ink, RuntimeObject doesn't strictly have Copy() yet in interface?
		// Value interface extends RuntimeObject.
		// VariablePointerValue has Copy().
		// We should probably rely on immutability for simple values or check specific types.
		// For now, simple assignment. Values should be treated as values.
		cp.GlobalVariables[k] = v // Shallow copy of map value (pointer)
		// If v is mutable, we might need v.Copy().
	}
	// TODO: Default globals, patch, etc.
	return cp
}
