# Entity System

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

Entities are lightweight identifiers for game objects. An Entity carries no data and no behavior — it is simply an ID that components are attached to. The entity system manages allocation, recycling, generational safety, and lifecycle states.

## Related Specifications

- [world-system.md](world-system.md) — Entities exist within a World
- [component-system.md](component-system.md) — Components are attached to Entities
- [command-system.md](command-system.md) — Entity spawn/despawn via deferred commands
- [hierarchy-system.md](hierarchy-system.md) — Parent-child relationships between entities

## 1. Motivation

Games constantly create and destroy objects. A robust entity system must:
- Provide cheap, copyable identifiers.
- Safely detect stale references to destroyed entities.
- Support high-throughput allocation and deallocation.
- Enable entity disabling without destroying data.

## 2. Constraints & Assumptions

- Entity IDs are 64-bit values combining an index and a generation counter.
- Generation prevents ABA problems: reusing an index increments the generation.
- Entity allocation is NOT thread-safe — it runs on the main thread or under exclusive access.
- Maximum entity count is bounded by index size (32-bit index = ~4 billion entities).

## 3. Core Invariants

- **INV-1**: An Entity ID is unique within its World for the lifetime of that entity.
- **INV-2**: After despawn, the same index may be reused but with an incremented generation.
- **INV-3**: Any operation on a despawned Entity (stale generation) must fail gracefully, not panic.
- **INV-4**: Entity allocation and deallocation are O(1) amortized.

## 4. Detailed Design

### 4.1 Entity ID Layout

```
Entity = { Index: uint32, Generation: uint32 }
```

- **Index**: Slot in the entity allocator. Reused after despawn.
- **Generation**: Incremented each time the slot is reused. Detects stale references.
- Packed into a single `uint64` for efficient storage and comparison.

### 4.2 Entity Lifecycle (5 Stages)

```mermaid
stateDiagram-v2
    [*] --> Unallocated
    Unallocated --> Allocated: Reserve
    Allocated --> Spawned: Spawn (add components)
    Spawned --> Despawned: Despawn
    Despawned --> Freed: Flush
    Freed --> Unallocated: Recycle (generation++)
```

1. **Unallocated** — Slot is in the free list.
2. **Allocated** — ID reserved but no components yet (used by Commands for deferred spawn).
3. **Spawned** — Entity is alive with components, queryable by systems.
4. **Despawned** — Marked for removal. Components still exist until flush.
5. **Freed** — Components removed, slot returned to free list with incremented generation.

### 4.3 Entity Allocator

- Free list of available indices (LIFO stack for cache locality).
- `Reserve()` — pops from free list or extends the arena. Returns Entity with current generation.
- `Free(entity)` — pushes index back to free list, increments generation for that slot.
- `IsAlive(entity)` — checks if stored generation matches the entity's generation.

### 4.4 Entity Disabling

Entities can be temporarily disabled without despawning. Disabled entities:
- Retain all their components and data.
- Are excluded from default queries (like a built-in `Without[Disabled]` filter).
- Can be explicitly included via `Query.IncludeDisabled()`.
- Useful for object pooling, pause mechanics, and LOD systems.

### 4.5 Entity Collections

Typed collections for efficient entity storage:
- **EntitySet** — Unordered unique set with O(1) insert/remove/contains.
- **EntityHashMap[V]** — Entity-keyed hash map.
- **EntityVec** — Ordered list of entities.

### 4.6 Entity References

- **EntityRef** — Read-only view of an entity's components (borrowed from World).
- **EntityMut** — Read-write view of an entity's components.
- **EntityWorldMut** — Full mutable access to entity + World (for structural changes).

### 4.7 Placeholder and Special Entities

- `Entity.PLACEHOLDER` — A sentinel value (index=MAX, generation=0) used as a default. Never valid in a World.
- No global "null entity" — use `Option[Entity]` / pointer-to-nil patterns instead.

## 5. Open Questions

- Should the engine support remote entity allocation (reserving IDs from multiple threads)?
- Entity names: built-in `Name` component or separate debug-only system?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft |
