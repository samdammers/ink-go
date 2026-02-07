package ink

import "fmt"

// Container is a node in the story hierarchy. It can contain other
// RuntimeObjects.
type Container struct {
	BaseRuntimeObject

	// Name of the container.
	name string

	// Content is the list of child objects.
	Content []RuntimeObject

	// NamedContent is a map of child objects keyed by name.
	NamedContent map[string]RuntimeObject

	// Flags for visit counting and turn counting.
	VisitsShouldBeCounted    bool
	TurnIndexShouldBeCounted bool
	CountingAtStartOnly      bool
}

// NewContainer creates a new empty Container.
func NewContainer() *Container {
	return &Container{
		NamedContent: make(map[string]RuntimeObject),
		Content:      []RuntimeObject{},
	}
}

// Name returns the name of the container.
func (c *Container) Name() string {
	return c.name
}

// SetName sets the name of the container.
func (c *Container) SetName(name string) {
	c.name = name
}

// HasValidName returns true if the container has a valid name.
func (c *Container) HasValidName() bool {
	return c.name != ""
}

// AddContent adds a RuntimeObject to the container's content list.
func (c *Container) AddContent(content RuntimeObject) error {
	if content.GetParent() != nil {
		return fmt.Errorf("content is already in %v", content.GetParent())
	}

	c.Content = append(c.Content, content)
	content.SetParent(c)

	c.tryAddNamedContent(content)
	return nil
}

func (c *Container) tryAddNamedContent(content RuntimeObject) {
	if namedContent, ok := content.(INamedContent); ok && namedContent.HasValidName() {
		c.addToNamedContentOnly(namedContent)
	}
}

func (c *Container) addToNamedContentOnly(namedContent INamedContent) {
	if runtimeObj, ok := namedContent.(RuntimeObject); ok {

		runtimeObj.SetParent(c)

		c.NamedContent[namedContent.Name()] = runtimeObj
	}
}

// AddNamedContent adds a named RuntimeObject to the container.
func (c *Container) AddNamedContent(name string, obj RuntimeObject) error {
	if obj.GetParent() != nil {
		return fmt.Errorf("object already has a parent")
	}
	if _, ok := c.NamedContent[name]; ok {
		return fmt.Errorf("named content '%s' already exists", name)
	}

	obj.SetParent(c)
	c.NamedContent[name] = obj
	return nil
}

// ContentAtPathComponent returns the object at the given path component.
func (c *Container) ContentAtPathComponent(component Component) (RuntimeObject, error) {
	if component.IsIndex() {
		if component.Index >= 0 && component.Index < len(c.Content) {
			return c.Content[component.Index], nil
		}
		return nil, fmt.Errorf("index out of bounds: %d", component.Index)
	}
	if found, ok := c.NamedContent[component.Name]; ok {
		return found, nil
	}
	// Fallback: Check if C# implementation checks offsets etc.
	// For now simple lookup.
	return nil, fmt.Errorf("content not found for name: %s", component.Name)
}

// GetPathForContent returns the path component for a child object.
func (c *Container) GetPathForContent(content RuntimeObject) (Component, error) {
	// 1. Check NamedContent
	// If the object has a name, check if it's stored under that name.
	if named, ok := content.(INamedContent); ok && named.HasValidName() {
		if found, ok := c.NamedContent[named.Name()]; ok && found.GetBase() == content.GetBase() {
			return NewComponentWithName(named.Name()), nil
		}
	}

	// 2. Scan Content
	targetBase := content.GetBase()
	for i, obj := range c.Content {
		if obj.GetBase() == targetBase {
			return NewComponentWithIndex(i), nil
		}
	}

	// Not found
	return Component{}, fmt.Errorf("child not found in container")
}
