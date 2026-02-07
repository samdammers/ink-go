# Go Implementation Architecture

This document details the specific architectural decisions and patterns used in the Go implementation of the Ink Runtime. It is intended for core maintainers and senior engineers.

## 1. Type System & Polymorphism

Ink is a dynamically typed language implemented in a statically typed host (Go).

### The `RuntimeObject` Interface

The base unit of content is the `RuntimeObject` interface.

```go
type RuntimeObject interface {
    DebugMetadata() *DebugMetadata
    Owns() bool // Memory management hint
    Copy() RuntimeObject
}
```

### Generic Values

To avoid the overhead and unsafety of raw `interface{}`, we use Go 1.18+ Generics for value types.

- **Base**: `value[T]` embedding `BaseRuntimeObject`.
- **Implementations**: `IntValue` (`value[int]`), `FloatValue` (`value[float64]`), `ListValue`, etc.
- **Execution**: The `EvaluationStack` is a slice of `RuntimeObject`. Operations must cast these objects to `Value` interfaces to perform math.

## 2. Execution Model

### The Step Loop (`Story.step`)

The core heartbeat of the engine is the `step()` method in `story.go`.

1.  **Check Pointer**: Is the `CurrentPointer` null? If so, try to return from the call stack.
2.  **Read Object**: Fetch the object at the current pointer.
3.  **Process**: Switch on the type of object.
    - **Value**: Push to Evaluation Stack.
    - **ControlCommand**: Update internal state (Push/Pop/Duplicate stack).
    - **Divert**: Change the `CurrentPointer`.
    - **VariableAssignment**: Update `VariablesState`.
4.  **Advance**: Move the pointer to the next index in the container.

### Pointer Resolution

Pointers in Ink are essentially paths (`root.0.thread.3`).

- **Design**: We cache resolved pointers where possible, but Serialization forces us to rely on string-path resolution.
- **Resolution**: `PointerAtPath` traverses the containment hierarchy using `ContentAtPath` to find the target object.

## 3. Persistence (Save/Load)

We employ a **DTO (Data Transfer Object)** pattern to handle JSON serialization.

- **Separation**: `StoryState` (Runtime) <-> `StoryStateDto` (JSON).
- **Reasoning**:
  - Go struct tags (`json:"..."`) were insufficient for the complex, deeply nested, and polymorphic nature of Ink's save format.
  - It allows us to "flatten" circular references (Pointers) into String Paths during save and re-resolve them during load.
- **Logic**: Located in `ink/persistence.go` (Save) and `ink/persistence_loader.go` (Load).

## 4. Variable Observation

The engine supports an Observer pattern for variable changes.

- **Mechanism**: Is handled in `VariablesState`.
- **Implementation**: When `Assign()` is called, we check if `variableChangedEvent` is registered.
- **Syncing**: Changes are propagated immediately unless in a batch context (not fully implemented in this port yet).

## 5. Control Flow & Glue

Ink's "Glue" (`<>`) behavior allows joining text across multiple lines/logic blocks.

- **Implementation**: `Story.step` logic checks for `Glue` objects.
- **Effect**: It prevents the emission of a newline character that would normally be generated when a content block ends.

## 6. Directory Structure

- `/ink`: The core library. Flat package structure `package ink` to allow shared internal access to unexported fields (like `state`) while exposing a clean public API (`Story`).
- `/cmd/ink-cli`: A reference CLI implementation used for testing and running .json files.
- `/test`: Integration and conformance tests.

## 7. Known Divergences from C#

- **Numeric Types**: We prioritize `int` for whole numbers during deserialization to maintain index safety, whereas C# might preserve `float` loosely. Usage of `math.Trunc` checks handles this conversion logic.
