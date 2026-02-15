package test

import (
	"strings"
	"testing"

	"github.com/samdammers/ink-go/ink"
)

// TestNestedChoicePathResolution verifies that choices in nested containers
// resolve their paths correctly relative to their containers.
func TestNestedChoicePathResolution(t *testing.T) {
	// Generic Ink structure:
	// root -> Choice A (leads to 'nested') | Choice B (leads to 'fail_root')
	// nested -> Choice A (leads to 'fail_nested') | Choice B (leads to 'success')
	// If path resolution is broken, Choosing B in 'nested' might jump to 'fail_root' (Choice B of root)
	jsonStr := `{"inkVersion":21,"root":[["^Root content.","\n","ev","str","^Go to nested","/str","/ev",{"*":"0.c-0","flg":20},"ev","str","^Go to fail_root","/str","/ev",{"*":"0.c-1","flg":20},{"c-0":["^ ",{"->":"nested"},"\n",{"->":"0.g-0"},{"#f":5}],"c-1":["^ ",{"->":"fail_root"},"\n",{"->":"0.g-0"},{"#f":5}],"g-0":["done",null]}],"done",{"nested":[["^Nested content.","\n","ev","str","^Go to fail_nested","/str","/ev",{"*":".^.c-0","flg":20},"ev","str","^Go to success","/str","/ev",{"*":".^.c-1","flg":20},{"c-0":["^ ",{"->":"fail_nested"},"\n",{"#f":5}],"c-1":["^ ",{"->":"success"},"\n",{"#f":5}]}],null],"fail_nested":["^Fail nested.","\n","done",null],"success":["^Success content.","\n","done",null],"fail_root":["^Fail root.","\n","done",null]}],"listDefs":{}}`

	story, err := ink.NewStory(jsonStr)
	if err != nil {
		t.Fatalf("Failed to load story: %v", err)
	}

	// 1. Start
	_, err = story.ContinueMaximally()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// 2. Choose "Go to nested" (Choice 0)
	err = story.ChooseChoiceIndex(0)
	if err != nil {
		t.Fatalf("ChooseChoiceIndex(0) failed: %v", err)
	}

	text, err := story.ContinueMaximally()
	if err != nil {
		t.Fatalf("ContinueMaximally failed after Enter: %v", err)
	}

	if !strings.Contains(text, "Nested content") {
		t.Errorf("Expected 'Nested content', got '%s'", text)
	}

	// 3. Choose "Go to success" (Choice 1)
	if len(story.GetCurrentChoices()) < 2 {
		t.Fatalf("Expected at least 2 choices, got %d", len(story.GetCurrentChoices()))
	}

	err = story.ChooseChoiceIndex(1)
	if err != nil {
		t.Fatalf("ChooseChoiceIndex(1) failed: %v", err)
	}

	text, err = story.ContinueMaximally()
	if err != nil {
		t.Fatalf("ContinueMaximally failed after Walk: %v", err)
	}

	expectedSnippet := "Success content"
	if !strings.Contains(text, expectedSnippet) {
		t.Errorf("Path resolution failure: Expected '%s', got '%s'", expectedSnippet, text)
	}
}
