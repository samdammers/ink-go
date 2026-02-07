package ink

import "testing"

func TestContainerContent(t *testing.T) {
	c := NewContainer()

	// 1. Add a StringValue
	s := NewStringValue("hello")
	c.AddContent(s)

	// 2. Add an IntValue
	i := NewIntValue(123)
	c.AddContent(i)

	// 3. Add a nested Container
	nestedC := NewContainer()
	c.AddContent(nestedC)

	if len(c.Content) != 3 {
		t.Fatalf("Expected 3 items in content, got %d", len(c.Content))
	}

	// Check types
	if _, ok := c.Content[0].(*StringValue); !ok {
		t.Error("Expected first item to be a *StringValue")
	}
	if _, ok := c.Content[1].(*IntValue); !ok {
		t.Error("Expected second item to be an *IntValue")
	}
	if _, ok := c.Content[2].(*Container); !ok {
		t.Error("Expected third item to be a *Container")
	}

	// Check parent pointers
	if s.GetParent() != c {
		t.Error("StringValue parent not set correctly")
	}
	if i.GetParent() != c {
		t.Error("IntValue parent not set correctly")
	}
	if nestedC.GetParent() != c {
		t.Error("Nested Container parent not set correctly")
	}
}

func TestContainerNamedContent(t *testing.T) {
	c := NewContainer()

	// Add a named nested container
	namedC := NewContainer()
	namedC.SetName("nested")
	c.AddContent(namedC)

	if len(c.NamedContent) != 1 {
		t.Fatalf("Expected 1 item in NamedContent, got %d", len(c.NamedContent))
	}

	found, ok := c.NamedContent["nested"]
	if !ok {
		t.Fatal("Could not find 'nested' in NamedContent")
	}

	if found != namedC {
		t.Error("NamedContent points to wrong object")
	}

	// Add an unnamed value, should not be in named content
	s := NewStringValue("unnamed")
	c.AddContent(s)

	if len(c.NamedContent) != 1 {
		t.Errorf("Unnamed content should not be added to NamedContent. Expected 1, got %d", len(c.NamedContent))
	}

	// Test AddNamedContent helper
	helperC := NewContainer()
	err := c.AddNamedContent("helper", helperC)
	if err != nil {
		t.Fatalf("AddNamedContent failed: %v", err)
	}
	if len(c.NamedContent) != 2 {
		t.Errorf("Expected 2 items in NamedContent after helper, got %d", len(c.NamedContent))
	}
	if c.NamedContent["helper"] != helperC {
		t.Error("AddNamedContent set the wrong object")
	}
	if helperC.GetParent() != c {
		t.Error("AddNamedContent did not set parent")
	}
}

func TestContainerContentOrdering(t *testing.T) {
	c := NewContainer()

	s1 := NewStringValue("one")
	s2 := NewStringValue("two")
	s3 := NewStringValue("three")

	c.AddContent(s1)
	c.AddContent(s2)
	c.AddContent(s3)

	if c.Content[0] != s1 {
		t.Error("Content order incorrect at index 0")
	}
	if c.Content[1] != s2 {
		t.Error("Content order incorrect at index 1")
	}
	if c.Content[2] != s3 {
		t.Error("Content order incorrect at index 2")
	}
}

func TestAddContentWithExistingParent(t *testing.T) {
	c1 := NewContainer()
	c2 := NewContainer()

	s := NewStringValue("hello")

	err := c1.AddContent(s)
	if err != nil {
		t.Fatalf("Initial AddContent failed: %v", err)
	}

	err = c2.AddContent(s)
	if err == nil {
		t.Fatal("Expected an error when adding content with an existing parent, but got nil")
	}

	if len(c2.Content) != 0 {
		t.Errorf("Container should not be modified on failed AddContent. Expected 0 items, got %d", len(c2.Content))
	}
}

func TestAddNamedContentWithExistingParent(t *testing.T) {
	c1 := NewContainer()
	c2 := NewContainer()

	s := NewStringValue("hello")

	err := c1.AddNamedContent("s", s)
	if err != nil {
		t.Fatalf("Initial AddNamedContent failed: %v", err)
	}

	err = c2.AddNamedContent("s", s)
	if err == nil {
		t.Fatal("Expected an error when adding named content with an existing parent, but got nil")
	}

	if len(c2.NamedContent) != 0 {
		t.Errorf("Container should not be modified on failed AddNamedContent. Expected 0 items, got %d", len(c2.NamedContent))
	}
}
