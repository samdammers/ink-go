# Glossary: Ink Concepts to Go Code

This document maps the narrative terminology used by Ink writers to the specific Go types and structures used in `go-ink`.

## Narrative Structure

| Ink Concept      | Description                                                                  | Go Type / Struct            |
| :--------------- | :--------------------------------------------------------------------------- | :-------------------------- |
| **Story**        | The entire compiled game script.                                             | `ink.Story`                 |
| **Knot**         | A major section of content (like a chapter/function). Syntax: `=== Name ===` | `ink.Container` (Top-level) |
| **Stitch**       | A subsection within a Knot. Syntax: `= Name`                                 | `ink.Container` (Nested)    |
| **Gather**       | A point where flow coalesces. Syntax: `-`                                    | `ink.Container`             |
| **Choice**       | An option presented to the player. Syntax: `* Hello`                         | `ink.Choice`                |
| **Choice Point** | The internal marker instruction for a choice.                                | `ink.ChoicePoint`           |

## Flow Control

| Ink Concept  | Description                                          | Go Type / Struct                                 |
| :----------- | :--------------------------------------------------- | :----------------------------------------------- |
| **Divert**   | A jump to another location. Syntax: `-> name`        | `ink.Divert`                                     |
| **Tunnel**   | A subroutine call that returns. Syntax: `-> knot ->` | Implemented via `Divert` pushing to `CallStack`. |
| **Function** | A knot that returns a value.                         | `ink.Container` using return instructions.       |
| **Glue**     | Runs content together without newlines. Syntax: `<>` | `ink.Glue`                                       |
| **Thread**   | Parallel flow execution.                             | `ink.Thread`, managed by `CallStack`.            |

## Data & Logic

| Ink Concept    | Description                                      | Go Type / Struct                      |
| :------------- | :----------------------------------------------- | :------------------------------------ |
| **Variable**   | Global storage state. Syntax: `VAR x = 1`        | `ink.VariablesState` entry.           |
| **Temp**       | Local temporary variable. Syntax: `~ temp x = 1` | Stored in `CallStack` frame.          |
| **List**       | A multi-state enum flag set.                     | `ink.ListValue` / `ink.InkList`       |
| **Evaluation** | Inline math/logic. Syntax: `{ x + 1 }`           | `ink.ControlCommand` (EvalStart/End). |

## Internal Engine Terms

| Term                 | Description                                                       |
| :------------------- | :---------------------------------------------------------------- |
| **Pointer**          | A `{Container, Index}` tuple pointing to the current instruction. |
| **Path**             | A string representation of a pointer (e.g., `root.knot.0`).       |
| **Visit Count**      | Tracks how many times a container has been played.                |
| **CallStack**        | The stack of frames tracking function calls and tunnel recursion. |
| **Evaluation Stack** | The temporary stack used for math operations (RPN style).         |
