package test

import (
	"fmt"
	"math"
	"testing"

	"github.com/samdammers/ink-go/ink"
)

// The "Victory Lap" - A full integration test of the engine.
func TestVictoryLap(t *testing.T) {
	fmt.Println("Starting Victory Lap Integration Test...")

	// 1. The Complex JSON (Simulated Compilation)
	// Story:
	// - Starts with 0 gold.
	// - Calls External Function 'pow(10, 2)' to set gold to 100.
	// - Adds 'Sword' to Inventory (List).
	// - Enters a Tunnel ('shop').
	// - Inside Tunnel: Checks 'Inventory' list and prints it.
	// - Returns from Tunnel.
	// - Ends.
	jsonStory := `
    {
        "inkVersion": 21,
        "root": [
            ["^Start. ", "\n", "ev", 10, 2, {"x()": "pow", "exArgs": 2}, "/ev", {"VAR=": "gold", "re": true}, 
             "ev", {"list": {"Inventory": "Sword"}}, {"VAR=": "inv", "re": true}, "/ev",
             "->t-> shop", "^End. ", "done", 
             {"shop": ["^In Shop. ", "ev", {"VAR?": "inv"}, "out", "/ev", "\n", "->->", "done"]}
            ]
        ],
        "listDefs": {"Inventory": {"Sword": 1, "Shield": 2}}
    }`

	story, err := ink.NewStory(jsonStory)
	if err != nil {
		t.Fatalf("CRITICAL: Failed to load story: %v", err)
	}

	// 2. Bind External Function (The "Type Panic" Check)
	// We return a float64 to test your Safe-Guard against strict int requirements.
	// Updated signature to match Go implementation: (args []any) (any, error)
	story.BindExternalFunction("pow", func(args []any) (any, error) {
		// Args might be Int or Float depending on JSON parsing.
		// We handle both for robustness, though we patched JSON parser to prefer ints.
		var base, exp float64

		switch v := args[0].(type) {
		case int:
			base = float64(v)
		case float64:
			base = v
		default:
			return nil, fmt.Errorf("arg 0 expected number, got %T", v)
		}

		switch v := args[1].(type) {
		case int:
			exp = float64(v)
		case float64:
			exp = v
		default:
			return nil, fmt.Errorf("arg 1 expected number, got %T", v)
		}

		return math.Pow(base, exp), nil
	})

	// 3. Playthrough
	// Step 1: Start -> Tunnel Entry
	text, err := story.Continue()
	if err != nil {
		t.Fatalf("Runtime Error at Step 1: %v", err)
	}
	if text != "Start. \n" {
		t.Errorf("Step 1 Fail: Expected 'Start. \\n', got '%s'", text)
	}

	// Step 2: Inside Tunnel (The "Orphaned List" Check)
	// The list 'inv' is printed. If Origin is lost, this might crash or print raw values.
	text, err = story.Continue()
	if err != nil {
		t.Fatalf("Runtime Error at Step 2: %v", err)
	}

	if text != "In Shop. Inventory.Sword\n" && text != "In Shop. Sword\n" && text != "In Shop. Sword" {
		t.Errorf("Step 2 Fail: Expected 'In Shop. Inventory.Sword\\n' or 'In Shop. Sword', got %q", text)
	}

	// Step 3: Return from Tunnel (The "Dirty Stack" Check)
	text, err = story.Continue()
	if err != nil {
		t.Fatalf("Runtime Error at Step 3: %v", err)
	}
	if text != "End. " {
		t.Errorf("Step 3 Fail: Expected 'End. ', got '%s'", text)
	}

	// 4. Final Sanity Check
	if story.CanContinue() {
		t.Errorf("Story should be done, but CanContinue is true.")
	}

	fmt.Println("Victory Lap Passed. Engine is stable.")
}
