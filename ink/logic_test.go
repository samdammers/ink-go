package ink

import (
	"testing"
)

func TestMathAndLogic(t *testing.T) {
	jsonLimit := `{"root": [["ev", 5, 3, "-", "/ev", "out", "\n", "done", null]], "inkVersion": 21}`

	story, err := NewStory(jsonLimit)
	if err != nil {
		t.Fatalf("Failed to create story: %v", err)
	}

	output, err := story.Continue()
	if err != nil {
		t.Fatalf("Story error: %v", err)
	}

	expected := "2\n"
	if output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, output)
	}
}

func TestComplexLogic(t *testing.T) {
	// (2 * 3) + 5 = 11
	json := `{"root": [["ev", 2, 3, "*", 5, "+", "/ev", "out", "\n", "done", null]], "inkVersion": 21}`

	story, err := NewStory(json)
	if err != nil {
		t.Fatalf("Failed to create story: %v", err)
	}

	output, err := story.Continue()
	if err != nil {
		t.Fatalf("Story error: %v", err)
	}

	expected := "11\n"
	if output != expected {
		t.Errorf("Expected '%s', got '%s'", output, expected)
	}
}
func TestLogicEdgeCases(t *testing.T) {
	// 1. Mixed Type Equality (1 == 1.0)
	// Note: We use 1.0 in JSON to ensure it's a float.
	jsonEq := `{"root": [["ev", 1, 1.1, 0.1, "-", "==", "/ev", "out", "done", null]], "inkVersion": 21}`
	// Actually simpler: 1 vs 1.0.
	// But JSON unmarshal might treat 1.0 as 1 if it fits in integer?
	// To be safe in test data, let's trust the unmasrshaler or use a float that is clearly float but equal?
	// No, 1.0 is fine in standard JSON readers usually, but let's see.
	// Let's use logic that forces float: 0.5 + 0.5 == 1

	jsonEq = `{"root": [["ev", 1, 0.5, 0.5, "+", "==", "/ev", "out", "done", null]], "inkVersion": 21}`

	story, err := NewStory(jsonEq)
	if err != nil {
		t.Fatalf("Failed to create story: %v", err)
	}
	output, err := story.Continue()
	if err != nil {
		t.Fatalf("Story error: %v", err)
	}
	// True is usually printed as "1"? Or "true"?
	// NativeFunctionCall Equal returns NewIntValue(1).
	// IntValue stringifies to "1".
	if output != "1" {
		t.Errorf("Mixed Equality failed: Expected '1', got '%s'", output)
	}

	// 2. String Arithmetic ("val: " + 1)
	jsonAdd := `{"root": [["ev", "str", "^val: ", "/str", 1, "+", "/ev", "out", "done", null]], "inkVersion": 21}`
	story2, err := NewStory(jsonAdd)
	if err != nil {
		t.Fatalf("Failed to create story 2: %v", err)
	}
	output2, err := story2.Continue()
	if err != nil {
		t.Fatalf("Story 2 error: %v", err)
	}
	if output2 != "val: 1" {
		t.Errorf("String Arithmetic failed: Expected 'val: 1', got '%s'", output2)
	}

	// 2b. Reverse String Arithmetic (1 + " val")
	// Ensure it works when the Int is first on the stack
	jsonAddReverse := `{"root": [["ev", 1, "str", "^ val", "/str", "+", "/ev", "out", "done", null]], "inkVersion": 21}`
	story3, err := NewStory(jsonAddReverse)
	if err != nil {
		t.Fatalf("Failed to create story 3: %v", err)
	}
	output3, err := story3.Continue()
	if err != nil {
		t.Fatalf("Story 3 error: %v", err)
	}
	if output3 != "1 val" {
		t.Errorf("Reverse String Arithmetic failed: Expected '1 val', got '%s'", output3)
	}
}
