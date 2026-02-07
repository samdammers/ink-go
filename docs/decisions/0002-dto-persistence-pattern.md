# Decisions: DTO Pattern for State Persistence

## Context

The Ink runtime involves complex, interconnected data structures:

- **Circular References**: Stories point to States, States point to Flows, Flows point to CallStacks, which point back to Story Containers.
- **Runtime-Only Fields**: Some fields are transient caches or optimization pointers that should not be saved.
- **Strict JSON Format**: The save format is dictated by the upstream Inkle C# implementation. It uses terse keys (`"evalStack"`, `"currentDivertTarget"`) and specific nesting that does not match idiomatic Go struct tagging.

Attempting to annotate the core `StoryState`, `Flow`, and `CallStack` structs with `json:"..."` tags leads to:

1.  **Polluted Domain Models**: Core logic is cluttered with serialization concerns.
2.  **Inflexible Serialization**: You cannot easily change the internal structure without breaking the save format.
3.  **Impossible Mappings**: Some concepts (like a `Pointer` to a specific instruction) must be transformed into a different representation (a string path) to be serialized.

## The Decision

**We utilize a dedicated Data Transfer Object (DTO) layer for all serialization operations.**

We introduced `persistence_dto.go` containing mirror structs (`StoryStateDto`, `FlowDto`, etc.) that strictly adhere to the external JSON schema.

### Implementation Pattern

1.  **Serialization (`ToJSON`)**:
    - The engine traverses the live Runtime Objects.
    - It maps them to DTOs.
    - Complex objects (Pointers) are resolved to string paths.
    - `json.Marshal` is called on the DTO.

2.  **Deserialization (`LoadState`)**:
    - `json.Unmarshal` decodes into the DTO.
    - The engine iterates through the DTO.
    - It reconstructs the live Runtime Objects.
    - String paths are resolved back into live `Pointer` references using the `Story` context.

## Benefits

- **Conformance**: We can guarantee 100% compatibility with the standard Ink JSON format without compromising Go naming conventions or internal architecture.
- **Clean Architecture**: The core runtime logic (stepping, jumping, evaluating) remains completely unaware of JSON or storage concerns.
- **Migration Safety**: If we refactor the internal `Flow` struct, we only need to update the `ToDto/FromDto` mappers; the persisted data format remains stable.
