package ink

import (
	"testing"
)

// Task 7: Advanced Flow Control Verification
func TestTunnelFlow(t *testing.T) {
	// Structure: root -> tunnel -> content -> return -> next content
	// Replaced "-> tunnel" with {"->t->": "tunnel"} to match JSON parser implementation
	// Flattened root structure to ensure "tunnel" is accessible at root level
	jsonStr := `{"root": [{"->t->": "tunnel"}, "^After tunnel.", "\n", "done", {"tunnel": ["^In tunnel.", "\n", "->->", "done"]}], "inkVersion": 21}`

	story, err := NewStory(jsonStr)
	if err != nil {
		t.Fatalf("Failed to create story: %v", err)
	}

	// 1. Enter Tunnel
	text, err := story.ContinueMaximally()
	if err != nil {
		t.Fatalf("Continue failed: %v", err)
	}

	// Expect "In tunnel." followed by "After tunnel."
	// The engine processes the tunnel, returns, and continues until 'done'.
	// Detailed execution flow:
	// -> tunnel (Process)
	// ^In tunnel. (Output)
	// \n (Output)
	// ->-> (Process Pop)
	// ^After tunnel. (Output)
	// \n (Output)
	// done (Stop)

	// So we expect "In tunnel.\nAfter tunnel.\n" if everything executes in one go.

	expected := "In tunnel.\nAfter tunnel.\n"
	if text != expected {
		t.Errorf("Tunnel flow mismatch.\nGot: '%s'\nWant: '%s'", text, expected)
	}
}
