package ink

// INamedContent is an interface for objects that have a name.
type INamedContent interface {
	Name() string
	HasValidName() bool
}
