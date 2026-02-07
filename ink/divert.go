package ink

import "fmt"

// Divert represents a divert to another part of the story.
type Divert struct {
	*BaseRuntimeObject
	TargetPath         *Path
	VariableDivertName string
	PushesToStack      bool
	StackPushType      PushPopType
	IsExternal         bool
	IsConditional      bool
	ExternalArgs       int
}

// NewDivert creates a new Divert.
func NewDivert() *Divert {
	return &Divert{
		BaseRuntimeObject: NewBaseRuntimeObject(),
		PushesToStack:     false,
	}
}

// NewDivertWithPushType creates a new Divert that pushes to the stack.
func NewDivertWithPushType(stackPushType PushPopType) *Divert {
	d := NewDivert()
	d.PushesToStack = true
	d.StackPushType = stackPushType
	return d
}

func (d *Divert) GetTargetPath() *Path {
	// If path is relative, and we have a target pointer, resolve it?
	// Java doesn't do much here except returning the path.
	// But it does allow setting target content directly too.
	// For now, simple getter.
	return d.TargetPath
}

func (d *Divert) SetTargetPath(path *Path) {
	d.TargetPath = path
}

func (d *Divert) HasVariableTarget() bool {
	return d.VariableDivertName != ""
}

func (d *Divert) GetVariableDivertName() string {
	return d.VariableDivertName
}

func (d *Divert) SetVariableDivertName(name string) {
	d.VariableDivertName = name
}

func (d *Divert) String() string {
	if d.HasVariableTarget() {
		return fmt.Sprintf("Divert(variable: %s)", d.VariableDivertName)
	}
	if d.TargetPath == nil {
		return "Divert(null)"
	}
	return fmt.Sprintf("Divert(%s)", d.TargetPath.String())
}
