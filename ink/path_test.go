package ink

import (
	"reflect"
	"testing"
)

func TestPathCreation(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedComps []Component
		isRelative    bool
	}{
		{"FromString", "hello.world.123", []Component{NewComponentWithName("hello"), NewComponentWithName("world"), NewComponentWithIndex(123)}, false},
		{"FromStringRelative", ".hello.world", []Component{NewComponentWithName("hello"), NewComponentWithName("world")}, true},
		{"FromStringEmpty", "", []Component{}, false},
		{"FromStringRoot", ".", []Component{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := NewPathFromString(tt.input)
			if path.IsRelative != tt.isRelative {
				t.Errorf("got isRelative %v, want %v", path.IsRelative, tt.isRelative)
			}
			if !reflect.DeepEqual(path.Components, tt.expectedComps) {
				t.Errorf("got components %v, want %v", path.Components, tt.expectedComps)
			}
		})
	}
}

func TestPathToString(t *testing.T) {
	tests := []struct {
		name     string
		path     *Path
		expected string
	}{
		{"Simple", NewPathFromString("hello.world"), "hello.world"},
		{"Relative", NewPathFromString(".hello.world"), ".hello.world"},
		{"WithIndex", NewPathFromString("knot.1.stitch"), "knot.1.stitch"},
		{"Empty", NewPathFromString(""), ""},
		{"Parent", NewPathFromString("a.^.b"), "a.^.b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.path.String() != tt.expected {
				t.Errorf("got %q, want %q", tt.path.String(), tt.expected)
			}
		})
	}
}

func TestPathAppending(t *testing.T) {
	tests := []struct {
		name           string
		start          *Path
		append         any // *Path or Component
		expectedString string
	}{
		{"AppendPath", NewPathFromString("a.b"), NewPathFromString(".c.d"), "a.b.c.d"},
		{"AppendPathWithParent", NewPathFromString("a.b.c"), NewPathFromString(".^.d"), "a.b.d"},
		{"AppendMultipleParents", NewPathFromString("a.b.c.d"), NewPathFromString(".^.^.e"), "a.b.e"},
		{"AppendComponent", NewPathFromString("a.b"), NewComponentWithName("c"), "a.b.c"},
		{"AppendParentComponent", NewPathFromString("a.b"), ComponentToParent(), "a.b.^"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result *Path
			switch v := tt.append.(type) {
			case *Path:
				result = tt.start.PathByAppendingPath(v)
			case Component:
				result = tt.start.PathByAppendingComponent(v)
			}

			if result.String() != tt.expectedString {
				t.Errorf("got %q, want %q", result.String(), tt.expectedString)
			}
		})
	}
}

func TestPathHeadTail(t *testing.T) {
	path := NewPathFromString("a.b.c")

	if path.Head().String() != "a" {
		t.Errorf("Head() got %q, want 'a'", path.Head().String())
	}

	tail := path.Tail()
	if tail.String() != "b.c" {
		t.Errorf("Tail() got %q, want 'b.c'", tail.String())
	}

	if path.String() != "a.b.c" {
		t.Error("Original path should not be modified")
	}

	single := NewPathFromString("a")
	if single.Tail().String() != "." {
		t.Errorf("Tail of single element path should be relative empty, got %q", single.Tail().String())
	}
}
