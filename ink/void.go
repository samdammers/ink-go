package ink

// Void represents a void value (e.g. from a function with no return).
type Void struct {
	*BaseRuntimeObject
}

func NewVoid() *Void {
	return &Void{
		BaseRuntimeObject: NewBaseRuntimeObject(),
	}
}

func (v *Void) String() string {
	return "void"
}
