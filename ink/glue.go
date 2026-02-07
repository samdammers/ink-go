package ink

// Glue is a special RuntimeObject that joins two pieces of text together
// with no space in between.
type Glue struct {
	*BaseRuntimeObject
}

// NewGlue creates a new Glue object.
func NewGlue() *Glue {
	return &Glue{
		BaseRuntimeObject: NewBaseRuntimeObject(),
	}
}
