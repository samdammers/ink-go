package ink

import "testing"

func TestThreadHubPattern(t *testing.T) {
	// Scenario: A conversation hub where you can ask two questions, and the story remembers you asked them.
	// JSON Input similar to described in prompt:
	// {
	//     "root": [
	//         ["^Start.", "<- threadA", "<- threadB", "^End.", "done",
	//         {"threadA": ["^Thread A.", "done"]},
	//         {"threadB": ["^Thread B.", "done"]}
	//         ]
	//     ],
	//     "inkVersion": 21
	// }
	//
	// Notes:
	// "<- threadA" usually implies StartThread command (19) + Divert.
	// We need to construct a JSON that mirrors what the compiler produces.
	// Compiler: ControlCommandStartThread (19) -> Divert(Target).
	//
	// Manual JSON construction for "<- threadA":
	// "cmd": "thread" (ControlCommandStartThread)
	// {"->": "threadA"} (Divert)

	jsonStr := `{"root": [
		"^Start.", "\n",
		"thread", {"->": "threadA"},
		"thread", {"->": "threadB"},
		"^End.", "\n",
		"done",
		{
			"threadA": ["^Thread A.", "done"],
			"threadB": ["^Thread B.", "done"]
		}
	], 
	"inkVersion": 21}`

	story, err := NewStory(jsonStr)
	if err != nil {
		t.Fatalf("Failed to create story: %v", err)
	}

	// Execution
	text, err := story.ContinueMaximally()
	if err != nil {
		t.Fatalf("ContinueMaximally failed: %v", err)
	}

	// Expected behavior:
	// 1. Start.
	// 2. StartThread -> Fork Main. Main at threadA Divert.
	// 3. Thread A (Fork) runs.
	// 4. Thread A output: Thread A.
	// 5. Thread A done -> Pops.
	// 6. Main resumes. Main was advanced past threadA Divert?
	//    Yes, logic implemented in Divert handler advances prevThread.
	// 7. Main runs. Sees 19 (StartThread). Sets flag.
	// 8. Main sees Divert(threadB). Forks.
	// 9. Thread B runs. Output: Thread B.
	// 10. Thread B done -> Pops.
	// 11. Main resumes. Advanced past threadB Divert.
	// 12. Main runs "End.".
	// 13. Main done.

	// Order: Start. \n Thread A. Thread B. End. \n
	// Note: Thread A and B outputs might or might not have newlines depending on content.
	// We put "^Thread A." (no newline).
	// But usually standard flow accumulates text.

	expected := "Start.\nThread A.Thread B.End.\n"

	if text != expected {
		t.Errorf("Thread Hub mismatch.\nGot: '%s'\nWant: '%s'", text, expected)
	}
}

func TestSimultaneousChoices(t *testing.T) {
	// Scenario:
	// A "Quest Log" choice interaction running in parallel with "Dialogue" choices.
	// Structure:
	// Root -> Launches Thread QuestLog
	// Root -> Launches Thread Dialogue
	// Root -> done (Wait for choices)
	//
	// QuestLog: * [Check Objectives] -> ...
	// Dialogue: * [Say Hello] -> ...
	//
	// We expect BOTH choices to be available.

	jsonStr := `{"root": [
		"thread", {"->": "threadQuest"},
		"thread", {"->": "threadDialog"},
		"done",
		{
			"threadQuest": [
				"ev", "str", "^Check Objectives", "/str", "/ev", {"*": ".^.c-0", "flg": 20},
				"\n", "done",
				{"c-0": ["^Objectives Checked.", "done"]}
			],
			"threadDialog": [
				"ev", "str", "^Say Hello", "/str", "/ev", {"*": ".^.c-0", "flg": 20},
				"\n", "done",
				{"c-0": ["^Hello said.", "done"]}
			]
		}
	], "inkVersion": 21}`

	story, err := NewStory(jsonStr)
	if err != nil {
		t.Fatalf("Failed to create story: %v", err)
	}

	// Run
	_, err = story.ContinueMaximally()
	if err != nil {
		t.Fatalf("Continue failed: %v", err)
	}

	// Verify Choices
	choices := story.GetCurrentChoices()
	if len(choices) != 2 {
		t.Errorf("Simultaneous Choice Failure. Got %d choices, expected 2.", len(choices))
		for i, c := range choices {
			t.Logf("Choice %d: %s", i, c.Text)
		}
	}
}
