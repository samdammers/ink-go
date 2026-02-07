package ink

// Choice is a generated Choice from the story. A single ChoicePoint in the
// Story could potentially generate different Choices dynamically dependent on
// state, so they're separated.
type Choice struct {
	BaseRuntimeObject

	// The main text to be presented to the player for this Choice.
	Text string

	// The original index into currentChoices list on the Story when this Choice
	// was generated, for convenience.
	Index int

	// The target path that the Story should be diverted to if this choice is chosen.
	TargetPath *Path

	// The path to the original choice point.
	SourcePath string
	
	// Used in JSON deserialisation
	OriginalThreadIndex int
	IsInvisibleDefault  bool
	
	Tags []string
}

// NewChoice creates a new choice.
func NewChoice() *Choice {
	return &Choice{}
}

// PathStringOnChoice gets the path to the original choice point.
// In the Java implementation, this is stored as a string, but we can
// just return the TargetPath's string representation.
func (c *Choice) PathStringOnChoice() string {
	if c.TargetPath != nil {
		return c.TargetPath.String()
	}
	return ""
}

// SetPathStringOnChoice sets the path to the original choice point.
func (c *Choice) SetPathStringOnChoice(path string) {
	c.TargetPath = NewPathFromString(path)
}
