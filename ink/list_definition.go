package ink

// ListDefinition represents a defined list in Ink.
type ListDefinition struct {
	Name  string
	Items map[string]int
}

// NewListDefinition creates a new list definition.
func NewListDefinition(name string, items map[string]int) *ListDefinition {
	return &ListDefinition{
		Name:  name,
		Items: items,
	}
}

// ValueForItem returns the integer value of an item in this list definition.
func (ld *ListDefinition) ValueForItem(itemName string) (int, bool) {
	val, ok := ld.Items[itemName]
	return val, ok
}

// ContainsItem checks if the definition contains an item.
func (ld *ListDefinition) ContainsItem(itemName string) bool {
	_, ok := ld.Items[itemName]
	return ok
}
