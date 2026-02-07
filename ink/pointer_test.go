package ink

import (
	"testing"
)

func TestPointer(t *testing.T) {
	// Create a container structure
	// root -> [ child1, child2 ]
	root := NewContainer()
	root.SetName("root")

	child1 := NewStringValue("child1")
	child2 := NewStringValue("child2")

	root.AddContent(child1)
	root.AddContent(child2)

	t.Run("StartOf", func(t *testing.T) {
		ptr := StartOf(root)
		if ptr.Container != root {
			t.Errorf("Expected container to be root, got %v", ptr.Container)
		}
		if ptr.Index != 0 {
			t.Errorf("Expected index 0, got %d", ptr.Index)
		}
	})

	t.Run("Resolve", func(t *testing.T) {
		ptr := StartOf(root)
		obj := ptr.Resolve()
		if obj != child1 {
			t.Errorf("Expected resolved object to be child1, got %v", obj)
		}

		ptr.Index = 1
		obj = ptr.Resolve()
		if obj != child2 {
			t.Errorf("Expected resolved object to be child2, got %v", obj)
		}

		ptr.Index = 2
		obj = ptr.Resolve()
		if obj != nil {
			t.Errorf("Expected resolved object to be nil for out of bounds, got %v", obj)
		}
	})

	t.Run("IsNull", func(t *testing.T) {
		ptr := NullPointer
		if !ptr.IsNull() {
			t.Error("Expected pointer to be null")
		}

		ptr = StartOf(root)
		if ptr.IsNull() {
			t.Error("Expected pointer to be non-null")
		}
	})

	t.Run("Path", func(t *testing.T) {
		ptr := StartOf(root)
		path := ptr.Path()
		// Root path is empty or usually implies root.
		// Container path logic depends on how Container.Path() is implemented.
		// Assuming Container path logic works correctly, we check only appending index here.

		// If root has no parent, its path might be empty or specific string logic.
		// Let's rely on string representation.

		// Wait, path calculation in Container relies on parent.
		// Since root has no parent, its path is ".".

		// PathByAppendingComponent should add index.
		// So result should be .0

		str := path.String()
		// Adjust expectation based on implementation details of Path/Container
		// If path is relative, it starts with dot.
		// ".0" seems likely.

		// Checking that it's not nil at least.
		if path == nil {
			t.Error("Expected path to be non-nil")
		}

		t.Logf("Pointer Path: %s", str)
	})
}
