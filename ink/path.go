package ink

import (
	"strconv"
	"strings"
)

const (
	ParentID = "^"
)

// Component is a single part of a Path.
// It can be an integer index or a string name.
// In the Java code, this is a nested static class. In Go, we make it a
// separate struct in the same package.
type Component struct {
	Index int
	Name  string
}

// NewComponentWithIndex creates a new path component with an integer index.
func NewComponentWithIndex(index int) Component {
	return Component{Index: index, Name: ""}
}

// NewComponentWithName creates a new path component with a string name.
func NewComponentWithName(name string) Component {
	return Component{Index: -1, Name: name}
}

// IsIndex returns true if the component is an integer index.
func (c Component) IsIndex() bool {
	return c.Index >= 0
}

// IsParent returns true if the component represents a move to the parent container ('^').
func (c Component) IsParent() bool {
	return c.Name == ParentID
}

// String returns the string representation of the component.
func (c Component) String() string {
	if c.IsIndex() {
		return strconv.Itoa(c.Index)
	}
	return c.Name
}

// Path is the core data structure for navigating the story hierarchy.
// It's equivalent to a file path, but for story content.
type Path struct {
	Components []Component
	IsRelative bool

	componentsString string // cached string
}

// NewPath creates a new empty path.
func NewPath() *Path {
	return &Path{Components: []Component{}}
}

// NewPathFromString creates a new path from a string representation.
func NewPathFromString(str string) *Path {
	p := NewPath()
	p.SetComponentsString(str)
	return p
}

// NewPathWithComponents creates a new path with a given slice of components.
func NewPathWithComponents(components []Component) *Path {
	return &Path{Components: components}
}

// String implements the fmt.Stringer interface.
// This allows a Path to be used as a map key and improves debug output.
func (p *Path) String() string {
	if p.componentsString == "" {
		var sb strings.Builder
		if p.IsRelative {
			sb.WriteRune('.')
		}
		for i, comp := range p.Components {
			sb.WriteString(comp.String())
			if i < len(p.Components)-1 {
				sb.WriteRune('.')
			}
		}
		p.componentsString = sb.String()
	}

	return p.componentsString
}

// SetComponentsString parses a string to build the path components.
func (p *Path) SetComponentsString(str string) {
	p.Components = p.Components[:0] // Clear slice

	if str == "" {
		return
	}

	if str[0] == '.' {
		p.IsRelative = true
		str = str[1:]
	} else {
		p.IsRelative = false
	}

	if str == "" {
		return
	}

	parts := strings.Split(str, ".")
	for _, part := range parts {
		if index, err := strconv.Atoi(part); err == nil {
			p.Components = append(p.Components, NewComponentWithIndex(index))
		} else {
			p.Components = append(p.Components, NewComponentWithName(part))
		}
	}

	// Invalidate cache
	p.componentsString = ""
}

// Head returns the first component of the path.
func (p *Path) Head() *Component {
	if len(p.Components) > 0 {
		return &p.Components[0]
	}
	return nil
}

// Tail returns a new Path containing all but the first component.
func (p *Path) Tail() *Path {
	if len(p.Components) >= 2 {
		tailComps := p.Components[1:]
		return &Path{Components: tailComps}
	}
	// In the Java version, this returns Path.getSelf(), a relative path.
	// Let's replicate that.
	return &Path{IsRelative: true}
}

// PathByAppendingPath appends a path to the current path.
func (p *Path) PathByAppendingPath(pathToAppend *Path) *Path {
	if !pathToAppend.IsRelative {
		return pathToAppend
	}

	newPath := NewPath()
	upwardMoves := 0
	for i := 0; i < len(pathToAppend.Components); i++ {
		if pathToAppend.Components[i].IsParent() {
			upwardMoves++
		} else {
			break
		}
	}

	for i := 0; i < len(p.Components)-upwardMoves; i++ {
		newPath.Components = append(newPath.Components, p.Components[i])
	}

	for i := upwardMoves; i < len(pathToAppend.Components); i++ {
		newPath.Components = append(newPath.Components, pathToAppend.Components[i])
	}

	return newPath
}

// PathByAppendingComponent appends a component to the current path.
func (p *Path) PathByAppendingComponent(c Component) *Path {
	newPath := NewPath()
	newPath.Components = append(newPath.Components, p.Components...)
	newPath.Components = append(newPath.Components, c)
	return newPath
}

// ComponentToParent returns a component that represents a move to the parent container.
func ComponentToParent() Component {
	return NewComponentWithName(ParentID)
}
