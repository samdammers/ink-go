# Developer Guide: Workflow & Testing

This guide is for developers contributing to `go-ink`. It covers how to build, test, and debug the engine.

## Prerequisites

- **Go 1.18+**: Required for Generics support.
- **Git**: For version control.

## Building

The project is a standard Go module.

```bash
# Build the core library
go build ./ink/...

# Build the CLI tool
go build -o ink-player ./cmd/ink-cli
```

## Testing

Reliability is paramount for a runtime engine. We use two types of tests:

### 1. Unit & Integration Tests

These are Go tests located in the `test/` directory.

```bash
# Run all tests
go test ./test/...

# Run a specific persistence test (useful for Save/Load work)
go test ./test/persistence_test.go
```

### 2. Conformance Tests (`conformance_test.go`)

This is the most critical suite. It iterates through the `test/testdata` folder, identifying pairs of `.ink.json` (compiled story) and `.txt` (expected output).

- It loads the JSON.
- It runs the story.
- It auto-selects choices based on the `.txt` content.
*   It compares the engine output *exactly* against the expected text.

**Note:** Our conformance test data is sourced from the [blade-ink-java](https://github.com/bladecoder/blade-ink) project, which in turn mirrors the standard Ink runtime tests. When updating the engine or tracking upstream changes, we should ensure new tests from the reference implementation are imported into `test/testdata`.

**If you break a conformance test, you have broken compatibility with standard Ink.**

## Debugging Tips

### Tracing Execution

Since the engine runtime is a tight loop (`step()`), debugging can be verbose.

1.  **Breakpoints**: Set a breakpoint in `ink/story.go` inside the `Continue()` method.
2.  **Inspect State**:
    - `story.state.CurrentPointer`: Where are we absolutely?
    - `story.state.CurrentPointer.Resolve()`: What instruction is this?
    - `story.state.EvaluationStack`: What values are currently waiting to be operated on?

### Common Issues

- **"Stack Empty" Panic**: Usually means a mismatch between `EvalStart` and `EvalEnd` markers, or a DTO persistence failure dropping stack items.
- **Missing Text**: Check if a `Glue` (`<>`) object accidentally consumed a newline you expected.
- **Variable Mismatch**: Use `story.State().VariablesState.GetVariableWithName("x")` to inspect values at runtime.

## Adding Features

1.  **Porting**: If porting a feature from C#, start by finding the equivalent class in the C# source (e.g., `Story.cs`).
2.  **DTOs**: If existing state fields change, **you must update `persistence_dto.go`**. Adding a field to `StoryState` without adding it to the DTO means it won't be saved!
