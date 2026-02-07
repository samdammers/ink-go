package ink

// VariableAssignment represents an assignment to a variable in the story.
// The value to be assigned is popped off the evaluation stack, so no need to keep it here.
type VariableAssignment struct {
	*BaseRuntimeObject
	variableName     string
	isNewDeclaration bool
	isGlobal         bool
}

// NewVariableAssignment creates a new VariableAssignment.
func NewVariableAssignment(variableName string, isNewDeclaration bool) *VariableAssignment {
	return &VariableAssignment{
		BaseRuntimeObject: NewBaseRuntimeObject(),
		variableName:      variableName,
		isNewDeclaration:  isNewDeclaration,
		isGlobal:          false, // Default
	}
}

func (v *VariableAssignment) VariableName() string {
	return v.variableName
}

func (v *VariableAssignment) SetVariableName(name string) {
	v.variableName = name
}

func (v *VariableAssignment) IsNewDeclaration() bool {
	return v.isNewDeclaration
}

func (v *VariableAssignment) SetIsNewDeclaration(isNew bool) {
	v.isNewDeclaration = isNew
}

func (v *VariableAssignment) IsGlobal() bool {
	return v.isGlobal
}

func (v *VariableAssignment) SetIsGlobal(isGlobal bool) {
	v.isGlobal = isGlobal
}

func (v *VariableAssignment) String() string {
	return "VarAssign to " + v.variableName
}
