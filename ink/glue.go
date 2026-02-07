package ink

// Glue is a special RuntimeObject that joins two pieces of text together
// with no space in between.
type Glue struct {
	*BaseRuntimeObject
}

func NewGlue() *Glue {
	return &Glue{
		BaseRuntimeObject: NewBaseRuntimeObject(),
	}
}
