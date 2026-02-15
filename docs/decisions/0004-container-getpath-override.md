# Decisions: Overriding GetPath for Proper Container Resolution

## Context

In implementation languages like Java and C#, `Container` typically inherits from a base `RTObject` class. When a method like `getPath()` is called on a `Container`, the `this` reference inside `RTObject.getPath()` still refers to the `Container` instance.

In Go, we use struct embedding (`BaseRuntimeObject` embedded in `Container`). While this mimics inheritance, method promotion behaves differently. When a method is promoted from the embedded struct, its receiver is the _embedded struct_ (`BaseRuntimeObject`), not the outer struct (`Container`).

## The Problem

`BaseRuntimeObject.GetPath()` calculates the path by traversing up the parent hierarchy. It calls `parent.GetPathForContent(child)`.

When `GetPath()` was called on a `Container` via the promoted method:

1. The receiver `child` was the embedded `*BaseRuntimeObject`.
2. `parent.GetPathForContent(child)` checked if `child` implemented `INamedContent`.
3. Since `BaseRuntimeObject` does not implement `INamedContent` (only `Container` does), the check failed.
4. The parent could not find the child by name, leading to incorrect path resolution (falling back to indices or failure).

This caused a regression where nested choices (which rely on accurate relative paths) would resolve to incorrect targets.

## The Decision

**We override `GetPath()` specifically on `*Container` to ensure the correct type identity is preserved.**

By implementing `func (c *Container) GetPath() *Path`, we ensure that:

1. The receiver is the full `*Container` instance.
2. When calling `parent.GetPathForContent(c)`, the parent receives an object that satisfies `INamedContent`.
3. Path resolution correctly uses the container's name.

## Alignment with Reference Implementation

This change aligns the behavior with the reference Java/C# implementations. In those languages, virtual dispatch ensures that `this` always refers to the runtime type (the `Container`), allowing `instanceof INamedContent` checks to succeed even within base class methods. Our Go override manually enforces this same behavior.
