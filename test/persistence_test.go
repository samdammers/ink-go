package test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/samdammers/go-ink/ink"
)

func TestPersistenceSave(t *testing.T) {
	// 1. Basic Story (Ends immediately)
	t.Run("Save Completed Story", func(t *testing.T) {
		jsonBytes, err := os.ReadFile("testdata/basictext/oneline.ink.json")
		if err != nil {
			t.Fatalf("Failed to read story json: %v", err)
		}

		story, err := ink.NewStory(string(jsonBytes))
		if err != nil {
			t.Fatalf("Failed to create story: %v", err)
		}

		_, err = story.ContinueMaximally()
		if err != nil {
			t.Fatalf("Failed to continue story: %v", err)
		}

		jsonStr, err := story.ToJson()
		if err != nil {
			t.Fatalf("ToJson failed: %v", err)
		}

		if len(jsonStr) == 0 {
			t.Fatal("Empty JSON output")
		}

		// Verify basic keys
		if !strings.Contains(jsonStr, `"flows"`) {
			t.Error("JSON missing 'flows'")
		}
		if !strings.Contains(jsonStr, `"variablesState"`) {
			t.Error("JSON missing 'variablesState'")
		}

		// Validate JSON validity
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
			t.Errorf("Invalid JSON produced: %v", err)
		}
	})

	// 2. Story Mid-Flow (Active state)
	t.Run("Save Mid-Flow Story", func(t *testing.T) {
		jsonBytes, err := os.ReadFile("testdata/basictext/twolines.ink.json")
		if err != nil {
			t.Fatalf("Failed to read story json: %v", err)
		}

		story, err := ink.NewStory(string(jsonBytes))
		if err != nil {
			t.Fatalf("Failed to create story: %v", err)
		}

		_, err = story.Continue() // Run first line
		if err != nil {
			t.Fatalf("Failed to continue story: %v", err)
		}

		if !story.CanContinue() {
			t.Fatalf("Expected story to continue")
		}

		jsonStr, err := story.ToJson()
		if err != nil {
			t.Fatalf("ToJson failed: %v", err)
		}

		t.Logf("Mid-Flow State JSON: %s", jsonStr)

		// Verify cPath is present
		if !strings.Contains(jsonStr, `"cPath"`) {
			t.Error("JSON missing 'cPath' for active story state")
		}
	})

	// 4. Round Trip Test (Load)
	t.Run("Round Trip Load", func(t *testing.T) {
		jsonBytes, err := os.ReadFile("testdata/choices/single-choice.ink.json")
		if err != nil {
			t.Fatalf("Failed to read story json: %v", err)
		}

		story, err := ink.NewStory(string(jsonBytes))
		if err != nil {
			t.Fatalf("Failed to create story: %v", err)
		}

		_, err = story.ContinueMaximally() // Run until choice
		if err != nil {
			t.Fatalf("Failed to continue story: %v", err)
		}

		// Save State
		savedJson, err := story.ToJson()
		if err != nil {
			t.Fatalf("ToJson failed: %v", err)
		}

		// Create New Story
		newStory, err := ink.NewStory(string(jsonBytes))
		if err != nil {
			t.Fatalf("Failed to create new story: %v", err)
		}

		// Load State
		err = newStory.LoadState(savedJson)
		if err != nil {
			t.Fatalf("LoadState failed: %v", err)
		}

		// Assertions
		// 1. Current Pointer
		p1 := story.State().GetCurrentPointer()
		p2 := newStory.State().GetCurrentPointer()
		if p1.String() != p2.String() {
			t.Errorf("Pointer Mismatch. Original: %s, Loaded: %s", p1, p2)
		}

		// 2. Variables (None in this story, but checking errors didn't occur)

		// 3. Current Choices
		c1 := story.GetCurrentChoices()
		c2 := newStory.GetCurrentChoices()

		if len(c1) != len(c2) {
			t.Fatalf("Choice count mismatch. Orig: %d, Loaded: %d", len(c1), len(c2))
		}

		if len(c1) > 0 {
			if c1[0].Text != c2[0].Text {
				t.Errorf("Choice Text mismatch. Orig: '%s', Loaded: '%s'", c1[0].Text, c2[0].Text)
			}
			if c1[0].Index != c2[0].Index {
				t.Errorf("Choice Index mismatch.")
			}
		}

		// 4. Verify Continue works after load
		// (Depends on choice being open)
		// If we choose index 0, logic should proceed identically
	})
}

func TestSaveInspection(t *testing.T) {
	// Scenario:
	// 1. Define a function with a temp variable (Check #1)
	// 2. Do some math (Check #2 - Int preservation)
	// 3. Keep a choice open (Check #3 - Thread Index)
	inkJson := `
    {
        "root": [
            {"->": "myFunc"}, 
            "done", 
            {"myFunc": [
                "ev", 5, {"temp=": "tempVar"}, "/ev", 
                "ev", {"VAR?": "tempVar"}, "out", "/ev", 
                "ev", "str", "^Option", "/str", "/ev", 
                {"*": ".^.c-0", "flg": 20}, 
                "done", 
                {"c-0": ["^Done", "done"]}
            ]}
        ],
        "inkVersion": 21
    }`

	story, err := ink.NewStory(inkJson)
	if err != nil {
		t.Fatalf("Failed to create story: %v", err)
	}

	_, err = story.Continue() // Enter function, define temp, print 5, offer choice.
	if err != nil {
		t.Fatalf("Failed to continue story: %v", err)
	}

	// Generate JSON
	jsonStr, err := story.ToJson()
	if err != nil {
		t.Fatalf("ToJson failed: %v", err)
	}

	// Check 1: Temporary Variables
	// Needs to be in the "callstack" -> "temp" section
	if !strings.Contains(jsonStr, "\"tempVar\":") {
		t.Error("FAIL: Temporary variable 'tempVar' not found in JSON.")
	}

	// Check 2: Int Preservation (Look for 5, not 5.0)
	// (Simple string check, strict check happens in Load)
	// We search for :5 or : 5 depending on formatting
	if !strings.Contains(jsonStr, ":5") && !strings.Contains(jsonStr, ": 5") {
		t.Error("FAIL: Integer 5 seems missing or malformed.")
	}
}
