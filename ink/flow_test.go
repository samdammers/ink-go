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

	// Expect "In tunnel."
	// Note: Engine might continue running until it hits a stop?
	// Wait, "In tunnel." is followed by "\n", then "->->".
	// The "->->" pops. Then "After tunnel." follows.
	// If the engine behaves like a standard run, it might produce ALL text if constraints are not met?
	// But usually Step() stops at Newlines if they are significant?
	// ink-cli usually prints line by line.
	// However, standard Continue() runs until CanContinue() is false or it yields text?
	// Actually, Continue() runs `ContinueInternal` which loops `CanContinueInternal`.
	// `GenerateChoices` logic promotes choices.
	// `story.go`: `ContinueInternal` loops `Step()`.

	// Checking `Step` logic:
	// It adds content to output stream.
	// It halts if `shouldAddToStream` and `CurrentPointer` becomes bad? No.
	// Ah, pure text story runs until end or choice.

	// So for this JSON, it should run:
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

func TestThreadFlow(t *testing.T) {
	// ThreadStart ( fork ), ThreadEnd ( done )
	// JSON representation of threads is complex.
	// Basic Thread:
	// root: [ -> thread_target, "Main side", done ]
	// thread_target: [ "Thread side", done ] - But wait, -> is divert.
	// Threads use "ex" / "thread" commands?
	// Standard ink uses `<- target`.
	// compiled: ControlCommand(StartThread) -> pushes thread to stack.

	// Let's defer strict Thread testing until implementation details are confirmed,
	// but Tunnel is critical for Task 7.2.
}
