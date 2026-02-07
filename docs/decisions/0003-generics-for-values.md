# Decisions: Generic Value Implementation

## Context

The Ink language allows variables and the evaluation stack to hold values of various types: Integers, Floats, Strings, Booleans, Divert Targets, Lists, etc.

In the original C# implementation, this is handled via class inheritance (`Value<T> : Value`). In pre-1.18 Go, this would typically be handled via `interface{}` validation or code generation, leading to substantial type-checking boilerplate or reduced safety.

## The Decision

**We leverage Go 1.18+ Generics to implement a base `value[T]` struct.**

```go
type value[T any] struct {
    BaseRuntimeObject
    Value T
}
```

Concrete types embed this generic base:

```go
type IntValue struct {
    value[int]
}

type FloatValue struct {
    value[float64]
}
```

## Implications

1.  **Type Erasure in Collections**: Since Go does not support covariance in slices, the evaluation stack is stored as `[]RuntimeObject` (an interface), not a generic collection.
2.  **Casting**: We define a `Cast(newType ValueType) (Value, error)` interface method. This allows polymorphic runtime conversion (e.g., adding an `Int` and a `Float` results in the `Int` being cast to `Float` first).

## Benefits

- **Code Reduction**: Accessor methods like `GetValueObject()` are implemented once on the generic receiver `*value[T]` rather than repeated for every type.
- **Compile-Time Safety**: Within the concrete types, `v.Value` is guaranteed to be of type `T`, preventing accidental assignment of the wrong underlying primitive.
