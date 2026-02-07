package ink

import "testing"

func TestChoiceLogicGates(t *testing.T) {
	// Gate 1: Fall-Through Logic
	// ensure engine executes choice op, but continues to next content until DONE
	t.Run("FallThrough", func(t *testing.T) {
		// ^Start. \n ev...Choice...pop... *... ev...Choice...pop... *... \n ^End of block. \n done
		jsonStr := `{"root": [["^Start.", "\n", "ev", "str", "^Opt1", "/str", "/ev", {"*": ".^.c-0"}, "\n", "^End of block.", "\n", "done", {"c-0": ["^Chose 1", "done"]}]], "inkVersion": 21}`

		s, err := NewStory(jsonStr)
		if err != nil {
			t.Fatalf("Failed to load story: %v", err)
		}

		text, err := s.ContinueMaximally()
		if err != nil {
			t.Fatalf("Continue failed: %v", err)
		}

		// It should contain "Start." AND "End of block."
		// Because the engine falls through the choice definition.
		if text == "" || len(s.GetCurrentChoices()) != 1 {
			t.Errorf("Fall-through failed. Text: '%s', Choices: %d (want 1)", text, len(s.GetCurrentChoices()))
		}
	})

	// Gate 2: Cleanup Logic
	t.Run("Cleanup", func(t *testing.T) {
		jsonStr := `{"root": [["ev", "str", "^Opt1", "/str", "/ev", {"*": ".^.c-0"}, "done", {"c-0": ["^Res", "done"]}]], "inkVersion": 21}`
		s, _ := NewStory(jsonStr)
		s.Continue()

		if len(s.GetCurrentChoices()) == 0 {
			t.Fatal("Setup failed: no choices")
		}

		err := s.ChooseChoiceIndex(0)
		if err != nil {
			t.Fatalf("Choose failed: %v", err)
		}

		if len(s.GetCurrentChoices()) != 0 {
			t.Errorf("Dangling choice bug: CurrentChoices not empty after selection")
		}
	})

	// Gate 3: Infinite Loop / Pointer Advance
	// Choice points to a gathered point containing the choice itself?
	// Simpler: Knot -> Choice -> Divert to Knot
	// This relies on Step() entering the target container correctly.
	t.Run("PointerAdvance", func(t *testing.T) {
		// root -> [ (loop) ^Start. *Opt -> c-0 ]
		// c-0 -> -> loop
		// JSON structure for a loop is complex to hand-write, using a simple recursive divert
		// root: [ limit: ... ]
		// We'll trust the mechanism verified in Loop check:
		// Target is c-0. c-0 contains divert to root.

		// Map: root:[ mark, choice, done, c-0:[ -> mark ] ] ??
		// Simpler: root is a loop.
		// [ ^Start. * Opt. -> c-0 ]  c-0: [ -> root ] is hard without named knots.
		// Let's use internal check: Verify pointer moves to c-0 content.

		jsonStr := `{"root": [["^Start.", "ev", "str", "^Opt", "/str", "/ev", {"*": ".^.c-0"}, "done", {"c-0": ["^Content", "done"]}]], "inkVersion": 21}`
		s, _ := NewStory(jsonStr)
		s.Continue()

		choice := s.GetCurrentChoices()[0]
		// targetPath := choice.TargetPath.String() // likely root.0.c-0
		_ = choice

		s.ChooseChoiceIndex(0)

		// NOW: Pointer should be AT start of c-0 content, OR at c-0 container waiting to resolve.
		// If we call Continue(), it should produce "Content".

		res, _ := s.Continue()
		if res != "Content" {
			t.Errorf("Pointer advance logic failed? Got '%s' instead of 'Content'", res)
		}
	})
}
