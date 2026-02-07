# Project Map: Navigating the Codebase

This document maps the physical file structure to the logical architecture. Use this as a guide to locate the code responsible for specific engine features.

## Directory Structure

```text
/
├── cmd/
│   └── ink-cli/       # Reference implementation of a console player
├── ink/               # The core runtime library (package ink)
├── test/              # Integration and conformance tests
├── test_data/         # External ink files used for verification
└── docs/              # Architecture and decision records
```

## The Core Library (`/ink`)

The `ink` package is flat, but the files can be grouped by their responsibility:

### 1. The Brain (Orchestration)

- **`story.go`**: The API entry point. Contains the main `Continue()` loop, `ChooseChoiceIndex()`, and initial setup.
- **`story_state.go`**: Holds the dynamic state (variables, callstack, evaluation stack) for a single session.
- **`story_impl.go`** (if applicable): Often internal step logic lives here or within `story.go`.

### 2. The Skeleton (Static Content)

These files define the structure of the compiled story.

- **`container.go`**: Defines `Container`, the node type that holds lists of other objects.
- **`object.go`**: Defines the `RuntimeObject` interface that all content must implement.
- **`value.go`**: Defines `IntValue`, `StringValue`, `FloatValue` (the generic `value[T]`).
- **`control_command.go`**: Defines metadata markers like `BeginString`, `NoOp`, `Pop`.
- **`divert.go`**: Defines jumps/gotos.

### 3. The Nervous System (navigating)

- **`pointer.go`**: A lightweight struct `{Container, Index}` pointing to a specific instruction.
- **`path.go`**: Logic for string paths (e.g., `root.sub.1`) used to find objects in the tree.

### 4. The Memory (Dynamic State)

- **`variables_state.go`**: Manages global variables (`VAR x = 1`).
- **`call_stack.go`**: Manages the stack of "Threads". Handles function calls and returns.
- **`choice.go`** & **`choice_point.go`**: Manages the list of options presented to the player.

### 5. Disk I/O (Save/Load)

- **`persistence.go`**: `ToJSON` implementation.
- **`persistence_loader.go`**: `LoadState` implementation.
- **`persistence_dto.go`**: Struct definitions for the JSON format.

## Key Interactions

### The `Continue()` Sequence

When you call `story.Continue()`, the code touches these files in roughly this order:

1.  **`story.go`**: Starts the loop.
2.  **`pointer.go`**: Resolves where we are.
3.  **`story.go`** (Internal step): Fetches the next object.
4.  **`runtime_object.go`**: Executed.
    - If Text -> Buffer.
    - If Divert -> **`path.go`** calculates new location -> **`pointer.go`** updates.
    - If Variable -> **`variables_state.go`** updates.

### The `ChooseChoiceIndex()` Sequence

1.  **`story.go`**: Validates the index.
2.  **`choice.go`**: Retrieves the `Choice` object.
3.  **`call_stack.go`**: The thread associated with that choice is revived.
4.  **`pointer.go`**: The instruction pointer is moved to the choice's destination.
