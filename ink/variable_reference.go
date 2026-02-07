package ink

import "fmt"

// VariableReference represents a named reference to a variable.
type VariableReference struct {
	*BaseRuntimeObject
	Name         string
	PathForCount *Path
}

// NewVariableReference creates a new VariableReference.
func NewVariableReference(name string) *VariableReference {
	return &VariableReference{
		BaseRuntimeObject: NewBaseRuntimeObject(),
		Name:              name,
	}
}

func (vr *VariableReference) String() string {
	if vr.Name != "" {
		return fmt.Sprintf("VAR?(%s)", vr.Name)
	} else if vr.PathForCount != nil {
		return fmt.Sprintf("CNT?(%s)", vr.PathForCount.String())
	}
	return "VAR?"
}
