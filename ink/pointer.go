package ink

import "fmt"

// Pointer is used to point to a particular / current point in the story.
// Where Path is a set of components that make content fully addressable,
// this is a reference to the current container, and the index of the current
// piece of content within that container.
type Pointer struct {
	Container *Container
	Index     int
}

// NewPointer creates a new Pointer.
func NewPointer(container *Container, index int) Pointer {
	return Pointer{
		Container: container,
		Index:     index,
	}
}

// NullPointer returns a "null" pointer (-1 index, nil container).
var NullPointer = Pointer{Container: nil, Index: -1}

// StartOf returns a pointer to the start of the container.
func StartOf(container *Container) Pointer {
	return Pointer{
		Container: container,
		Index:     0,
	}
}

// Resolve returns the RuntimeObject this pointer refers to.
func (p Pointer) Resolve() RuntimeObject {
	if p.Index < 0 {
		return p.Container
	}

	if p.Container == nil {
		return nil
	}

	content := p.Container.Content
	if len(content) == 0 {
		return p.Container
	}

	if p.Index >= len(content) {
		return nil
	}

	return content[p.Index]
}

// IsNull returns true if the pointer is null.
func (p Pointer) IsNull() bool {
	return p.Container == nil
}

// Next returns the pointer to the next instruction (incremented index).
func (p Pointer) Next() Pointer {
	if p.IsNull() {
		return p
	}
	return Pointer{
		Container: p.Container,
		Index:     p.Index + 1,
	}
}

// Path returns the path to the pointed object.
func (p Pointer) Path() *Path {
	if p.IsNull() {
		return nil
	}

	if p.Index >= 0 {
		return p.Container.GetPath().PathByAppendingComponent(NewComponentWithIndex(p.Index))
	}
	return p.Container.GetPath()
}

// String returns the string representation.
func (p Pointer) String() string {
	if p.Container == nil {
		return "Ink Pointer (null)"
	}

	return fmt.Sprintf("Ink Pointer -> %s -- index %d", p.Container.GetPath().String(), p.Index)
}

// Assign copies the values from another pointer.
// Note: Since Pointer is a struct in Go, assignment is just `p1 = p2`.
// This method is provided for compatibility with Java semantics if needed via pointer receiver.
func (p *Pointer) Assign(other Pointer) {
	p.Container = other.Container
	p.Index = other.Index
}
