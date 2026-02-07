# Java vs Go: Implementation Differences

This document outlines the key architectural divergences between the reference Java/C# implementations (like `blade-ink-java`) and this Go port. Understanding these differences is crucial for developers porting features or debugging behavior differences.

## 1. Inheritance vs Composition

The most significant structural difference stems from Go's lack of classical class inheritance.

### Java/C# Approach

The reference implementations rely heavily on deep inheritance hierarchies.

- `Object` -> `RuntimeObject` -> `Value<T>` -> `StringValue`
- Shared logic is often placed in abstract base classes.
- Polymorphism is handled via `virtual` method overrides.

### Go Approach

We use **Composition** and **Interfaces**.

- **Interface**: `RuntimeObject` is an `interface`, not a class.
- **Embedding**: struct Reuse is achieved by embedding base structs.
  ```go
  type IntValue struct {
      value[int]        // Embeds generic behavior
      BaseRuntimeObject // Embeds metadata fields
  }
  ```
- **Impact**: You cannot cast a `StringValue` to a `BaseRuntimeObject`. You must access the interface methods. Type assertions (`obj.(*IntValue)`) are used frequently where Java would use `instanceof`.

## 2. Persistence: The DTO Pattern

### Java/C# Approach

Java implementation often map JSON directly to domain objects, using libraries (like Jackson) that support complex annotations, private field injection, or custom deserializers mixed into the class.

### Go Approach

We enforce a strict separation using **Data Transfer Objects (DTOs)**.

- **Why**: Go's `encoding/json` relies on exported (Capitalized) fields. Ink's JSON format uses lowercase, specific keys (`currentChoices`, `evalStack`).
- **Solution**: We do not pollute the core logic with JSON tags. Instead, `persistence_dto.go` defines a mirror world of structs solely for saving/loading.
  - _Save_: Runtime -> DTO -> JSON
  - _Load_: JSON -> DTO -> Runtime

## 3. Numeric Fidelity

### Java

Java's numeric handling (via `Double` vs `Integer` objects) allows for distinct storage of `5.0` vs `5`, though behavior depends heavily on the specific JSON library config.

### Go

Go's `encoding/json` unmarshals **all** numbers as `float64` by default.

- **Adaptation**: We implemented a "Best Fit" strategy in the loader. If a float `5.0` has no fractional part (`val == math.Trunc(val)`), we explicitly convert it to an `int` RuntimeObject.
- **Consequence**: See [0001-number-type-fidelity.md](../decisions/0001-number-type-fidelity.md).

## 4. Generics and Collections

### Java

Java uses Type Erasure but allows `List<RuntimeObject>`.

### Go

We use Go 1.18+ Generics for the `Value` types (`value[T]`), but the containers (Stacks/Lists) hold `interface{}` or `RuntimeObject` interfaces.

- **Impact**: This requires explicit casting (boxing/unboxing) when popping from the stack to perform operations.
  ```go
  // Go requires casting interface back to concrete type
  val := stack.Pop()
  if iVal, ok := val.(*IntValue); ok { ... }
  ```

## 5. Package Structure

### Java

Deeply nested package structure (`com.blade.ink.runtime...`).

### Go

Flat, idiomatic package structure (`package ink`).

- We expose a minimal public API (`Story`, `Trace`).
- Internal types (`storyState`, `callStack`) are exposed to the package but hidden from consumers where appropriate, though currently many are exported to support the white-box testing required for the complex engine logic.
