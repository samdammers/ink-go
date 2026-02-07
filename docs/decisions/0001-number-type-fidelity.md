# Decisions: Numeric Type Fidelity in Persistence

## Context

In Ink, variables are dynamically typed. A variable can check if it is explicitly a float or an int.

```ink
VAR x = 5.0
```

Internally, the `go-ink` engine represents this as a `*FloatValue` with value `5.0`.

When saving the game state, the engine serializes this data to JSON. The JSON specification (and Go's standard `encoding/json` library) does not distinguish between distinct integer and float types; it simply has a `Number` type.

This creates a scenario known as "Type Erasure" for whole-number floats:

1.  **Runtime**: `x` is `5.0` (Float).
2.  **Serialization**: `x` is saved as `5` in the JSON string (standard JSON behavior for whole numbers to save space).
3.  **Deserialization**: When `LoadState` runs, it sees `5`.

## The Decision

**We have chosen to deserialize all whole-number JSON values as `int` (`*IntValue`), even if they were originally `float` (`*FloatValue`).**

If the JSON contains `5`, the engine loads it as `int(5)`.
If the JSON contains `5.1`, the engine loads it as `float64(5.1)`.

### Why?

1.  **JSON Ambiguity**: Without complex metadata sidecars (which would break format compatibility with other Ink engines), we cannot know if `5` was originally intended to be `5` or `5.0`.
2.  **Engine Logic**: Ink uses integers heavily for list logic, pointers, and counters. Go's standard unmarshaller defaults everything to `float64`. If we kept everything as `float64`, we would break internal list indexing and equality checks. Converting whole numbers to `int` is the safest default for the _internal_ engine logic.

## Impact on Developers

This decision has practically zero impact on standard Ink logic. Ink is permissive: `5 == 5.0` is true, and `5 + 2.5` becomes `7.5`.

**However, this has a Critical Impact on External Functions (Go Bindings).**

If you bind a Go function that expects a `float64`, and you rely on strict Go type assertions, your game may crash after loading a save file.

### The Antipattern (Do Not Do This)

```go
// ❌ Dangerous: This will PANIC if 'val' was loaded from a saved game as an int
story.BindExternalFunction("my_func", func(args []any) (any, error) {
    val := args[0].(float64) // Panic: interface conversion: interface {} is int, not float64
    return val * 2.0, nil
})
```

### The Solution (Robust Unboxing)

Developers must assume that any numeric argument passed to an external function could be either `int` or `float64`.

```go
// ✅ Safe: Handles both types gracefully
story.BindExternalFunction("my_func", func(args []any) (any, error) {
    var val float64

    switch v := args[0].(type) {
    case int:
        val = float64(v)
    case float64:
        val = v
    default:
        return nil, fmt.Errorf("expected number")
    }

    return val * 2.0, nil
})
```

## Summary

- **Behavior**: Whole number floats (`5.0`) become integers (`5`) after a Save/Load cycle.
- **Responsibility**: Developers binding External Functions must manually check for both `int` and `float64` to prevent runtime panics.
