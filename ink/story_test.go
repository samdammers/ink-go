package ink

import "testing"

func TestHelloWorld(t *testing.T) {
	json := `{"root": [["^Hello World", "\n", "done", null]], "inkVersion": 21}`

	story, err := NewStory(json)
	if err != nil {
		t.Fatalf("NewStory failed: %v", err)
	}

	result, err := story.Continue()
	if err != nil {
		t.Fatalf("Continue failed: %v", err)
	}

	// The user's request specified "Hello World", but the ink model includes
	// the newline in the content stream. The Continue method correctly
	// concatenates this.
	expected := "Hello World\n"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}
