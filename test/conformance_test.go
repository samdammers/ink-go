package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samdammers/go-ink/ink"
)

// Tests a sample of the standard ink conformance tests to gauge compliance.
// We do not run the full suite automatically as extracting expectations from Java code is non-trivial.
func TestConformanceSample(t *testing.T) {
	basePath := "testdata/"

	cases := []struct {
		Name     string
		File     string
		Choices  []int  // Sequence of choices to make. -1 means "run maximally".
		Expected string // Expected full output text after all interactions
	}{
		{
			Name:     "One Line",
			File:     "basictext/oneline.ink.json",
			Expected: "Line.\n",
		},
		{
			Name:     "Two Lines",
			File:     "basictext/twolines.ink.json",
			Expected: "Line.\nOther line.\n",
		},
		{
			Name:     "Tunnel Onwards Divert Override",
			File:     "tunnels/tunnel-onwards-divert-override.ink.json",
			Expected: "This is A\nNow in B.\n",
		},
		{
			Name:     "List Basic Operations",
			File:     "lists/basic-operations.ink.json",
			Expected: "b, d\na, b, c, e\nb, c\nfalse\ntrue\ntrue\n",
		},
		// Multi-Choice requires specialized runner logic (resetting or branching),
		// sticking to linear or single-branch validation for this simple loop.
		// choice/multi-choice.ink expects:
		// Run 1 (Choice 0): Hello, world!\nHello back!\nGoodbye\nHello back!\nNice to hear from you\n
		// Run 2 (Choice 1): Hello, world!\nHello back!\nGoodbye\nGoodbye\nSee you later\n
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			path := filepath.Join(basePath, tc.File)
			// Read file manually or passed to NewStory? NewStory takes string content.
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", path, err)
			}

			// Remove BOM if present? Ink compiler might output BOM. Go strings might not like it.
			// Standard Ink JSON usually doesn't have BOM, but let's be safe if json.Unmarshal complains.
			jsonStr := string(content)

			story, err := ink.NewStory(jsonStr)
			if err != nil {
				t.Fatalf("Failed to create story: %v", err)
			}

			// Execution
			var sb strings.Builder

			// Initial run
			text, err := story.ContinueMaximally()
			if err != nil {
				t.Fatalf("Continue failed: %v", err)
			}
			sb.WriteString(text)

			// Loops through choices if provided (not used in current simple cases)
			for _, choiceIdx := range tc.Choices {
				if len(story.GetCurrentChoices()) == 0 {
					t.Fatalf("Expected choices but found none.")
				}
				err := story.ChooseChoiceIndex(choiceIdx)
				if err != nil {
					t.Fatalf("Choose failed: %v", err)
				}
				text, err = story.ContinueMaximally()
				if err != nil {
					t.Fatalf("Continue after choice failed: %v", err)
				}
				sb.WriteString(text)
			}

			if sb.String() != tc.Expected {
				t.Errorf("Output Mismatch.\nGot:\n%q\nWant:\n%q", sb.String(), tc.Expected)
			}
		})
	}
}

func TestConformanceMultiChoice(t *testing.T) {
	// Specific test for the multi-choice branching behavior
	path := "testdata/choices/multi-choice.ink.json"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	jsonStr := string(content)

	// Branch 0
	t.Run("Choice 0", func(t *testing.T) {
		s, _ := ink.NewStory(jsonStr)
		out, _ := s.ContinueMaximally()
		// "Hello, world!\n"
		s.ChooseChoiceIndex(0)
		out2, _ := s.ContinueMaximally()

		full := out + out2
		expected := "Hello, world!\nHello back!\nNice to hear from you\n"
		if full != expected {
			t.Errorf("Got: %q\nWant: %q", full, expected)
		}
	})

	// Branch 1
	t.Run("Choice 1", func(t *testing.T) {
		s, _ := ink.NewStory(jsonStr)
		out, _ := s.ContinueMaximally()
		s.ChooseChoiceIndex(1)
		out2, _ := s.ContinueMaximally()

		full := out + out2
		expected := "Hello, world!\nGoodbye\nSee you later\n"
		if full != expected {
			t.Errorf("Got: %q\nWant: %q", full, expected)
		}
	})
}
