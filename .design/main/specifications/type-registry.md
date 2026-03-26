# Type Registry

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Related Specifications

- [component-system.md](component-system.md) — Components register in the type registry
- [scene-system.md](scene-system.md) — Scene serialization uses registry for dynamic types
- [definition-system.md](definition-system.md) — Definition System resolves JSON type names through the registry

## Overview

The Type Registry provides dynamic introspection capabilities for the engine, allowing the engine to manipulate types at runtime without compile-time knowledge of their specific structures.

## 1. Motivation

A static language like Go requires a centralized dynamic metadata store for systems like GUI editors, scene serialization, and scripting to work. Without it, every new component would need manual boilerplate for serialization and UI mapping.

## 2. Constraints & Assumptions

- **Go standard lib**: Use `reflect` but minimize overhead for performance.
- **Registration**: All types wishing to be serializable or visible in the editor MUST register.
- **Safety**: Runtime type manipulation must still respect Go's type system at the boundaries.

## 3. Core Invariants

- **INV-1**: Every registered type must have a globally unique String identifier (e.g., `namespace/TypeName`).
- **INV-2**: The registry MUST support complex types, including nested structs, slices, and maps.
- **INV-3**: Registration must be possible from anywhere (ideally via an `init()` hook or similar engine entry point).


## 4. Detailed Design

### 4.1 Registry Store

The central `TypeRegistry` stores the following for each type:
1. **Type ID**: The human-readable unique string.
2. **Reflect.Type**: The Go reflect type handle.
3. **Field Metadata**: Names, types, offsets, and custom Go tags (e.g., `editor:"hidden"`, `range:"0,10"`).
4. **Hooks**: Per-type functions for cloning, default initialization, and serialization.

### 4.2 Attribute System

The registry supports attaching metadata (attributes) to fields or types:
- **Min/Max Range**: For numeric fields.
- **Visibility**: Whether it should be shown in the editor.
- **Tooltips**: Documentation strings.


### 4.3 Proxy Operations

The registry provides a generic `DynamicObject` wrapper (proxy) to:
- Get/Set fields by name.
- Iterate fields of an arbitrary struct.
- Construct new instances of a type by its string ID.

## 5. Open Questions

- Performance penalty for using `reflect` in hot paths (e.g., per-frame UI updates).
- Handling of type versioning (renaming fields in different versions of the engine).

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft |
