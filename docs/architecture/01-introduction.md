# Introduction to go-ink Architecture

Welcome to the `go-ink` project! This document provides a high-level overview of the engine's architecture, designed to help new contributors understand how the pieces fit together.

## What is go-ink?

`go-ink` is a Go port of the **Ink Runtime**. Ink is a narrative scripting language created by Inkle Studios.

- **Ink Compiler**: (Not part of this project) Converts `.ink` text files into an intermediate `.ink.json` format.
- **Ink Runtime**: (This project) Loads that JSON and "plays" the story, handling logic, variables, text generation, and choices.

## High-Level Concepts

The engine is built around a few core objects that model the structure of a story.

### 1. The Story (`ink.Story`)

The `Story` is the main entry point. It represents the entire runtime engine for a single game session.

- **Input**: You initialize it with the JSON string content of a compiled story.
- **Responsibility**: It manages the global state, the current position in the text, and the interface for the player (getting text, making choices).

### 2. The Content Hierarchy (Containers & Objects)

Internally, an Ink story is a tree of objects.

- **Container**: Think of this as a folder or a compiled "function". It holds a list of instructions. The root of the story is a Container. Knots (`=== Name ===`) are Containers.
- **RuntimeObject**: Everything inside a container is a `RuntimeObject`. This includes:
  - **Text**: `"Hello World"`
  - **Control Commands**: Signal the engine to do something (e.g., "Begin String", "Pop Stack").
  - **Diverts**: Jumps to another part of the story (`-> nearby_knot`).
  - **Variable Assignments**: `~ x = 5`

### 3. The State (`ink.StoryState`)

While `Story` holds the static content (the text/logic instructions), `StoryState` holds the dynamic data for a playthrough.

- **Variables**: Global variables defined in Ink.
- **CallStack**: Tracks where we are. If you jump into a "Knot" (function), the CallStack remembers where to return to.
- **Evaluation Stack**: Used for math and logic. To calculate `time + 5`, we push `time`, push `5`, and then the "Add" command pops them both and pushes the result.

## The Loop: How it Runs

The lifecycle of an interaction usually looks like this:

1.  **Usage**: The player calls `story.Continue()`.
2.  **Execution**:
    - The engine looks at its **current pointer** (where it is in the Container tree).
    - It executes the instruction there.
    - If it's Text, it adds it to the output buffer.
    - If it's Logic (math, conditions), it runs it silently.
    - If it's a **Choice**, it adds it to the list of `currentChoices`.
3.  **Pause**: The engine stops when it hits a "Wait for Choice" marker or runs out of content.
4.  **Interaction**: The player picks a choice index (0, 1, 2...).
5.  **Resume**: The engine "chooses" that path and the loop starts again at step 1.

## Getting Started

If you are just starting to look at the code:

1.  Start at `ink/story.go` - look at `Continue()` and `NewStory()`.
2.  Look at `ink/container.go` to see how content is structured.
3.  Check `ink/value.go` to see how we handle dynamic types like Ints and Strings.
