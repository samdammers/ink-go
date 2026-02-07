package test

import (
	"testing"

	"github.com/samdammers/go-ink/ink"
)

// TestFloatTypeErasure demonstrates the loss of type fidelity for whole-number floats
// during the Save/Load cycle. This is an accepted behavior documented in
// docs/decisions/0001-number-type-fidelity.md.
func TestFloatTypeErasure(t *testing.T) {
	// Minimal valid JSON
	jsonStr := `{"inkVersion":21,"root":[["done",{"#f":5}],"done",null],"listDefs":{}}`

	t.Run("Original Story Preserves Float", func(t *testing.T) {
		s, err := ink.NewStory(jsonStr)
		if err != nil {
			t.Fatalf("Failed to create story: %v", err)
		}

		// Manually inject a float variable
		s.State().VariablesState.GlobalVariables["x"] = ink.NewFloatValue(5.0)

		// Check internal type of x
		val := s.State().VariablesState.GetVariableWithName("x")
		if _, ok := val.(*ink.FloatValue); !ok {
			t.Errorf("Expected x to be FloatValue initially, got %T with value %v", val, val)
		}
	})

	t.Run("Loaded Story Erases Type", func(t *testing.T) {
		s, err := ink.NewStory(jsonStr)
		if err != nil {
			t.Fatalf("Failed to create story: %v", err)
		}

		// Manually inject a float variable
		s.State().VariablesState.GlobalVariables["x"] = ink.NewFloatValue(5.0)

		// Save
		savedJSON, err := s.ToJSON()
		if err != nil {
			t.Fatalf("ToJSON failed: %v", err)
		}

		// Load into new story
		s2, err := ink.NewStory(jsonStr)
		if err != nil {
			t.Fatalf("NewStory for load failed: %v", err)
		}

		err = s2.LoadState(savedJSON)
		if err != nil {
			t.Fatalf("LoadState failed: %v", err)
		}

		// Check internal type of x
		val := s2.State().VariablesState.GetVariableWithName("x")

		if _, ok := val.(*ink.FloatValue); !ok {
			t.Logf("Confirmed: x converted to %T after load (%v)", val, val)

			// If it is IntValue, verify value is 5
			if iv, ok := val.(*ink.IntValue); ok {
				if iv.Value != 5 {
					t.Errorf("Expected 5, got %d", iv.Value)
				}
			} else {
				t.Errorf("Unexpected type %T", val)
			}
		} else {
			t.Log("Valid: x remained FloatValue (Unexpected for current implementation)")
		}
	})
}
