# Component System

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

Components are pure data attached to entities. They carry no logic — systems process components. The component system manages registration, storage strategy selection, lifecycle hooks, required component dependencies, clone behavior, and the type registry that maps runtime type information to component metadata.

## Related Specifications

- [world-system.md](world-system.md) — Components stored in the World
- [entity-system.md](entity-system.md) — Components attached to entities
- [query-system.md](query-system.md) — Systems access components via queries
- [change-detection.md](change-detection.md) — Component mutations tracked by ticks

## 1. Motivation

The component model defines how game data is structured. Key requirements:
- Efficient memory layout for cache-friendly iteration.
- Flexible storage: dense iteration vs. fast add/remove tradeoff.
- Automatic dependency management (adding Transform auto-adds GlobalTransform).
- Lifecycle hooks for cross-cutting concerns (physics body registration on add, cleanup on remove).

## 2. Constraints & Assumptions

- Components are plain data structures with no methods (logic lives in systems).
- Each component type has a unique ComponentID assigned at registration time.
- Component registration is one-time and immutable after first use.
- A component cannot be both Table-stored and SparseSet-stored — the choice is per-type.

## 3. Core Invariants

- **INV-1**: Each component type has exactly one storage strategy, declared at registration.
- **INV-2**: Required components are transitively resolved — if A requires B, and B requires C, then adding A adds B and C.
- **INV-3**: Component hooks (OnAdd, OnInsert, OnReplace, OnRemove) fire exactly once per event, in deterministic order.
- **INV-4**: Immutable components cannot be mutated after insertion. Attempting mutation is an error.
- **INV-5**: The type registry is the single source of truth for component metadata at runtime.

## 4. Detailed Design

### 4.1 Storage Strategies

#### Table Storage (Default)

- Column-oriented (Structure of Arrays): each component type = one contiguous array.
- Entities with the same archetype share a table.
- **Pros**: Excellent cache locality for iteration, SIMD-friendly.
- **Cons**: Moving an entity between archetypes (add/remove component) requires data copy.

#### Sparse Set Storage

- Entity-indexed sparse array with a dense data array.
- **Pros**: O(1) add/remove without archetype changes, good for frequently toggled components.
- **Cons**: Iteration less cache-friendly (random access by entity index).

Selection guideline: Use Table (default) for most components. Use SparseSet for components that are frequently added/removed (e.g., status effects, temporary markers).

### 4.2 Component Registration

Every component type must be registered before use:

```
ComponentDescriptor {
    Name:             string
    TypeID:           unique type identifier
    Size:             bytes
    Alignment:        bytes
    StorageType:      Table | SparseSet
    RequiredComponents: []ComponentID
    CloneBehavior:    Clone | Ignore | Custom
    Hooks:            ComponentHooks
    Immutable:        bool
}
```

Registration happens automatically on first use or explicitly during plugin setup.

### 4.3 Required Components

A component can declare dependencies:

```
Transform requires GlobalTransform
Mesh3D requires Transform, Visibility
Camera requires Transform, GlobalTransform, Frustum
```

When component A with requirements is inserted on an entity, the engine:
1. Resolves the full transitive dependency set.
2. Inserts any missing required components with their default values.
3. Fires hooks for all inserted components in dependency order.

Circular dependencies are forbidden and detected at registration time.

### 4.4 Component Hooks

Lifecycle callbacks invoked during component mutations:

| Hook | Trigger | DeferredWorld Access |
| :--- | :--- | :--- |
| **OnAdd** | First time component type appears on entity | Yes |
| **OnInsert** | Every insertion (including overwrite) | Yes |
| **OnReplace** | Component value overwritten (old value available) | Yes |
| **OnRemove** | Component about to be removed | Yes |

Hooks receive a DeferredWorld — they can read/write other components but cannot perform structural changes (spawn/despawn/add/remove). Structural changes must go through Commands.

### 4.5 Immutable Components

Components marked as immutable cannot be mutated after insertion:
- `Query<&mut ImmutableComponent>` is a compile-time or registration-time error.
- The only way to change the value is to remove and re-insert.
- Use case: entity identifiers, UUID components, archetype-defining tags.

### 4.6 Clone Behavior

When duplicating entities, each component defines how it is cloned:

- **Clone** (default): Deep copy of the component data.
- **Ignore**: Component is not copied to the new entity.
- **Custom**: User-defined clone function (e.g., assign new unique ID).

### 4.7 Bundles

Bundles are groups of components for convenient spawning:

```
PlayerBundle = { Transform, Velocity, Health, Player }
world.Spawn(PlayerBundle{...})
```

- A bundle is NOT a component — it dissolves into individual components on spawn.
- Bundles can contain other bundles (flattened recursively).
- Spawning a bundle triggers all individual component hooks.

### 4.8 Type Registry

The type registry maps runtime type information to engine metadata:

- ComponentID → ComponentDescriptor
- Type name → ComponentID (for serialization, debug, editor)
- Reflection data for dynamic component access (scene loading, inspector)

The registry is a World resource, populated during plugin initialization.

## 5. Open Questions

- Should components support inheritance / composition beyond required components?
- Maximum component count per archetype — is there a practical limit to enforce?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft from Bevy analysis |
