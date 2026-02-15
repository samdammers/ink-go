package ink

// RuntimeObject defines the base interface for all ink runtime content.
// It allows for polymorphic collections of story content.
type RuntimeObject interface {
	GetParent() RuntimeObject
	SetParent(parent RuntimeObject)
	GetPath() *Path
	GetBase() *BaseRuntimeObject
}

// BaseRuntimeObject is the base struct for all ink runtime content.
// It holds the basic state (parent, path) and implements the RuntimeObject interface.
// Other runtime types will embed this struct.
type BaseRuntimeObject struct {
	// parent is the parent object in the story hierarchy.
	// Note that this is a RuntimeObject interface, which allows for polymorphism.
	// In practice, parents are almost always *Container.
	parent RuntimeObject

	// TODO: Add DebugMetadata once that is ported.
	// debugMetadata *DebugMetadata

	// path is the Path to this object in the story hierarchy.
	// It's lazily initialized.
	path *Path
}

// NewBaseRuntimeObject creates a new BaseRuntimeObject.
func NewBaseRuntimeObject() *BaseRuntimeObject {
	return &BaseRuntimeObject{}
}

// GetParent returns the parent object.
func (r *BaseRuntimeObject) GetParent() RuntimeObject {
	return r.parent
}

// SetParent sets the parent object.
func (r *BaseRuntimeObject) SetParent(parent RuntimeObject) {
	r.parent = parent
}

// GetPath gets the path of this object in the story hierarchy.
// It is lazily initialized.
//
//nolint:dupl // Logic matches Container.GetPath but operates on *BaseRuntimeObject.
func (r *BaseRuntimeObject) GetPath() *Path {
	if r.path == nil {
		if r.parent == nil {
			r.path = NewPath()
		} else {
			// Find component of this object in parent
			parent, ok := r.parent.(*Container)
			var comp Component
			if ok && parent != nil {
				// We expect the parent to be a container
				// But we need to use type assertion carefully if circular deps existed (none here)
				c, err := parent.GetPathForContent(r)
				if err == nil {
					comp = c
				}
				// If err != nil, comp remains empty. This implies the object is attached but
				// not found in parent's content list. We fallback to empty/root component to avoid crashing.
			}

			// Path = ParentPath + Component
			if parent != nil {
				r.path = parent.GetPath().PathByAppendingComponent(comp)
			} else {
				r.path = NewPath()
			}
		}
	}
	return r.path
}

// GetBase returns the base runtime object.
func (r *BaseRuntimeObject) GetBase() *BaseRuntimeObject {
	return r
}
