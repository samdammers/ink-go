package ink

// SearchResult is returned when looking up content within the story.
// When looking up content within the story (e.g. in Container.ContentAtPath),
// the result is generally found, but if the story is modified, then when loading
// up an old save state, then some old paths may still exist. In this case we
// try to recover by finding an approximate result by working up the story hierarchy
// in the path to find the closest valid container.
type SearchResult struct {
	Object      RuntimeObject
	Approximate bool
}

// GetContainer returns the container object if the result object is a container.
func (s SearchResult) GetContainer() *Container {
	if s.Object == nil {
		return nil
	}
	if c, ok := s.Object.(*Container); ok {
		return c
	}
	if p, ok := s.Object.GetParent().(*Container); ok {
		return p
	}
	return nil
}
