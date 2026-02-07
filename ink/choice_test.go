package ink

import "testing"

func TestChoiceCreation(t *testing.T) {
	choice := NewChoice()

	choice.Text = "My Choice"
	choice.Index = 1
	choice.SetPathStringOnChoice("my.target.path")
	choice.SourcePath = "original.path"

	if choice.Text != "My Choice" {
		t.Errorf("Text not set correctly")
	}

	if choice.Index != 1 {
		t.Errorf("Index not set correctly")
	}

	if choice.PathStringOnChoice() != "my.target.path" {
		t.Errorf("PathStringOnChoice not set correctly")
	}

	if choice.SourcePath != "original.path" {
		t.Errorf("SourcePath not set correctly")
	}

	if choice.TargetPath.String() != "my.target.path" {
		t.Errorf("TargetPath not set correctly")
	}
}

func TestChoiceInteraction(t *testing.T) {
	jsonStr := `{"root": [["^Start.", "\n", "ev", "str", "^Option A", "/str", "/ev", {"*": ".^.c-0","flg": 20}, "ev", "str", "^Option B", "/str", "/ev", {"*": ".^.c-1","flg": 20}, "\n", "done", {"c-0": ["\n", "^You chose A.", "\n", "done", {"#n": ""}], "c-1": ["\n", "^You chose B.", "\n", "done", {"#n": ""}]}], null], "inkVersion": 21}`

	story, err := NewStory(jsonStr)
	if err != nil {
		t.Fatalf("Failed to create story: %v", err)
	}

	// 1. Initial Continue
	text, err := story.ContinueMaximally()
	if err != nil {
		t.Fatalf("First continue failed: %v", err)
	}
	expectedText := "Start.\n\n"
	if text != expectedText {
		t.Errorf("Expected '%s', got '%s'", expectedText, text)
	}

	// 2. Verify Choices
	choices := story.GetCurrentChoices()
	if len(choices) != 2 {
		t.Fatalf("Expected 2 choices, got %d", len(choices))
	}

	if choices[0].Text != "Option A" {
		t.Errorf("Choice 0 text incorrect: %s", choices[0].Text)
	}

	// Debug
	t.Logf("Choice 0 Target Path: %s", choices[0].TargetPath.String())
	t.Logf("Choice 0 Source Path: %s", choices[0].SourcePath)

	// DEBUG: Verify Structure
	root := story.MainContent
	if len(root.Content) == 0 {
		t.Fatalf("Root has no content")
	}
	inner, ok := root.Content[0].(*Container)
	if !ok {
		t.Fatalf("Root content 0 is not container")
	}
	t.Logf("Inner container content count: %d", len(inner.Content))
	if _, ok := inner.NamedContent["c-0"]; !ok {
		t.Errorf("Inner container missing c-0")
		// Debug keys
		for k := range inner.NamedContent {
			t.Logf("Found key: %s", k)
		}
	} else {
		t.Logf("Found c-0 in inner container")
		c0 := inner.NamedContent["c-0"].(*Container)
		t.Logf("c-0 content count: %d", len(c0.Content))
		if len(c0.Content) > 0 {
			t.Logf("c-0[0] type: %T", c0.Content[0])
		}
	}

	// 3. Choose
	// Verify Pointer Resolution first
	ptr := story.PointerAtPath(choices[0].TargetPath)
	if ptr.IsNull() {
		t.Fatalf("PointerAtPath failed to resolve %s", choices[0].TargetPath.String())
	}
	t.Logf("Resolved pointer: %v", ptr)

	err = story.ChooseChoiceIndex(0) // Choose A
	if err != nil {
		t.Fatalf("Choose failed: %v", err)
	}

	// 4. Continue after choice
	text, err = story.ContinueMaximally()
	if err != nil {
		t.Fatalf("Second continue failed: %v", err)
	}

	// Output expectation: \nYou chose A.\n based on c-0 content.
	// c-0: ["\n", "^You chose A.", "\n", "done", ...]
	expectedText2 := "\nYou chose A.\n"
	if text != expectedText2 {
		t.Errorf("Expected '%s', got '%s'", expectedText2, text)
	}
}
