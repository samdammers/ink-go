package ink

import (
	"testing"
)

func TestPathParsing(t *testing.T) {
	p := NewPathFromString("0.2.s")
	if len(p.Components) != 3 {
		t.Errorf("Expected 3 components, got %d", len(p.Components))
	}
	if p.String() != "0.2.s" {
		t.Errorf("Expected '0.2.s', got '%s'", p.String())
	}
	t.Logf("Components: %v", p.Components)
}
