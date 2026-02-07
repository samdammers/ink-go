package ink

import (
	"testing"
)

func TestGlueBehavior(t *testing.T) {
	// Setup
	mainContent := NewContainer()
	s := &Story{
		MainContent: mainContent,
	}
	s.state = NewStoryState(s)

	// Ensure CurrentFlow is initialized
	if s.state.CurrentFlow == nil {
		s.state.CurrentFlow = &Flow{
			OutputStream: []RuntimeObject{},
			Name:         "DEFAULT_FLOW",
			CallStack:    NewCallStack(s.MainContent),
		}
	}

	// Case 1: [Newline, Glue] -> "" (Preceding/Backward Glue)
	// Add "A", "\n", Glue, "B"
	s.state.PushToOutputStream(NewStringValue("A"))
	s.state.PushToOutputStream(NewStringValue("\n"))
	s.state.PushToOutputStream(NewGlue())
	s.state.PushToOutputStream(NewStringValue("B"))

	output := s.CurrentText()
	if output != "AB" {
		t.Errorf("Case 1 (Backward Glue) Failed. Expected 'AB', got '%s'", output)
	}

	// Case 2: [Glue, Newline] -> "" (Following/Forward Glue)
	// Reset
	s.state.CurrentFlow.OutputStream = []RuntimeObject{}
	s.state.PushToOutputStream(NewStringValue("C"))
	s.state.PushToOutputStream(NewGlue())
	s.state.PushToOutputStream(NewStringValue("\n"))
	s.state.PushToOutputStream(NewStringValue("D"))

	output = s.CurrentText()
	if output != "CD" {
		t.Errorf("Case 2 (Forward Glue) Failed. Expected 'CD', got '%s'", output)
	}

	// Case 3: [Newline, Space, Glue] -> " " (Newline removed, Space kept)
	// The glue should "look through" the space to kill the newline,
	// but KEEP the space.
	s.state.CurrentFlow.OutputStream = []RuntimeObject{}
	s.state.PushToOutputStream(NewStringValue("E"))
	s.state.PushToOutputStream(NewStringValue("\n"))
	s.state.PushToOutputStream(NewStringValue(" ")) // Inline whitespace
	s.state.PushToOutputStream(NewGlue())
	s.state.PushToOutputStream(NewStringValue("F"))

	output = s.CurrentText()
	// Ink standard: Glue kills the newline, but preserves the space.
	// Result should be "E F"
	if output != "E F" {
		t.Errorf("Case 3 (Whitespace Transparency) Failed. Expected 'E F', got '%s'", output)
	}
}
