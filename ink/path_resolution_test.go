package ink

import (
	"testing"
)

func TestApproximatePathResolution(t *testing.T) {
	// Regression test for: https://github.com/samdammers/ink-go/issues/INK_GO_ISSUE.md
	// Scenario: A path implies a child exists inside a leaf node (StringValue),
	// but it actually exists as a sibling in the parent container.

	// Construct Path: Root -> Inner(0)
	// Inner contains:
	//   0: StringValue("You stand...")
	//   1: ChoicePoint
	// Inner has NamedContent:
	//   "c-0": Container (The Choice content)

	root := NewContainer()
	inner := NewContainer()
	str := NewStringValue("You stand...")
	cp := NewChoicePoint(true, false, false, false, false)

	// Add content to Inner
	_ = inner.AddContent(str) // Index 0
	_ = inner.AddContent(cp)  // Index 1

	// Add the choice content as named content to Inner
	choiceContent := NewContainer()
	choiceContent.SetName("c-0")
	_ = inner.AddNamedContent("c-0", choiceContent)

	_ = root.AddContent(inner)

	// The Story needs to be initialized so the root path is set up correctly
	story := &Story{MainContent: root}
	// We need to force path strings to be generated, which usually happens during Story Init or lazy loading
	// But GetPath() on objects solves this.
	_ = root.GetPath()

	// The problematic path: ".^.0.c-0"
	// Meaning: From current context (let's say we are at the ChoicePoint 0.1),
	// go up (to Inner), go to index 0 (StringValue), then find "c-0".
	//
	// In a strict filesystem, 0.0 is a file (StringValue), so 0.0/c-0 is impossible.
	// In Ink, 0.0 is context, and c-0 is looked up in the container that *holds* 0.0 (Inner).

	targetPathStr := "0.0.c-0"
	// The original issue description specified "Path: 0.0.c-0".
	// While typically relative paths start with dots, we test absolute path resolution here for clarity.

	targetPath := NewPathFromString(targetPathStr)

	ptr := story.PointerAtPath(targetPath)

	if ptr.IsNull() {
		t.Fatalf("PointerAtPath returned Null for path '%s'. Expected fuzzy resolution to sibling 'c-0'.", targetPathStr)
	}

	// PointerAtPath for a container typically returns a pointer to the start of the container (index 0).
	// If the container is empty, resolving this pointer yields nil.
	// We add content to the choice container to facilitate verification of the resolved object.
	_ = choiceContent.AddContent(NewStringValue("Inside Choice"))

	// Re-evaluate
	ptr = story.PointerAtPath(targetPath)
	resolvedObj := ptr.Resolve()

	if resolvedObj == nil {
		t.Fatalf("Resolved object is nil. Pointer: %v", ptr)
	}

	if val, ok := resolvedObj.(*StringValue); ok {
		if val.Value != "Inside Choice" {
			t.Errorf("Resolved wrong object. Expected 'Inside Choice', got '%s'", val.Value)
		}
	} else {
		t.Errorf("Resolved object is not StringValue. Type: %T", resolvedObj)
	}
}
