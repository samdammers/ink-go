package ink

// Flow represents a flow execution context.
type Flow struct {
	Name           string
	CallStack      *CallStack
	OutputStream   []RuntimeObject
	CurrentChoices []*Choice
}

// NewFlow creates a new Flow.
func NewFlow(name string, story *Story) *Flow {
	f := &Flow{
		Name:           name,
		CallStack:      NewCallStack(story.MainContent), // Assuming Story.MainContent is accessible
		OutputStream:   make([]RuntimeObject, 0),
		CurrentChoices: make([]*Choice, 0),
	}
	return f
}

// Copy creates a deep copy of the Flow.
func (f *Flow) Copy() *Flow {
	copy := &Flow{
		Name:           f.Name,
		CallStack:      f.CallStack.Copy(),
		OutputStream:   make([]RuntimeObject, len(f.OutputStream)),
		CurrentChoices: make([]*Choice, len(f.CurrentChoices)),
	}
	// Copy OutputStream (RuntimeObjects are shared except simple values which are effectively immutable)
	for i, obj := range f.OutputStream {
		copy.OutputStream[i] = obj
	}
	// Copy Choices
	for i, choice := range f.CurrentChoices {
		// Choices references should be fine? Or do we deep copy choices?
		// Choices are RuntimeObjects.
		copy.CurrentChoices[i] = choice
	}
	return copy
}
