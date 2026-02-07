package ink

// ListDefinitionsOrigin stores definitions of lists.
type ListDefinitionsOrigin struct {
	Lists map[string]*ListDefinition
}

// NewListDefinitionsOrigin creates a new ListDefinitionsOrigin.
func NewListDefinitionsOrigin(lists []*ListDefinition) *ListDefinitionsOrigin {
	origin := &ListDefinitionsOrigin{
		Lists: make(map[string]*ListDefinition),
	}
	for _, l := range lists {
		origin.Lists[l.Name] = l
	}
	return origin
}
