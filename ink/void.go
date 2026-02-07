package ink

// Void represents a void value (e.g. from a function with no return).
type Void struct {
	*BaseRuntimeObject
}

// VoidName is the string representation of Void.
const VoidName = "void"

// NewVoid creates a new Void object.
func NewVoid() *Void {
	return &Void{
		BaseRuntimeObject: NewBaseRuntimeObject(),
	}
}

func (v *Void) String() string {
	return VoidName
}
