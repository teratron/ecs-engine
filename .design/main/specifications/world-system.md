# World System

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

The World is the central data store of the ECS engine. It owns all entities, components, resources, and schedules. Every game operation reads from or writes to a World instance. The World orchestrates entity allocation, archetype management, resource storage, and change tracking.

## Related Specifications

- [entity-system.md](entity-system.md) — Entity allocation lives inside the World
- [component-system.md](component-system.md) — Component storage managed by the World
- [query-system.md](query-system.md) — Queries operate on World data
- [command-system.md](command-system.md) — Commands are deferred mutations applied to the World
- [change-detection.md](change-detection.md) — World maintains the global change tick

## 1. Motivation

Every ECS needs a single owner for all game data. Without a centralized World:
- Entity and component lifetimes become ambiguous.
- Parallel system execution cannot be safely coordinated.
- Change detection has no global reference point.

The World is the answer: one struct that owns everything, enabling safe parallel access through controlled borrowing.

## 2. Constraints & Assumptions

- A single World instance is the norm. Multiple Worlds are permitted for isolation (e.g., render world, test world).
- The World is NOT thread-safe by default. Concurrent access is mediated by the scheduler.
- All entity and component operations go through the World or through Commands (which are applied to the World).
- Resources are global singletons keyed by type — one instance per type per World.

## 3. Core Invariants

- **INV-1**: Every Entity belongs to exactly one World. Entities cannot migrate between Worlds.
- **INV-2**: Component data for an Entity is always stored in the archetype corresponding to that Entity's component set.
- **INV-3**: The World's change tick is monotonically increasing and never wraps within a single run.
- **INV-4**: Clearing trackers (change detection reset) happens exactly once per update cycle, at a well-defined point.
- **INV-5**: A DeferredWorld provides limited, re-entrant-safe access for hooks and observers.

## 4. Detailed Design

### 4.1 World Structure

The World is composed of:

- **Entity Allocator** — Manages entity IDs and generations (see entity-system).
- **Archetypes** — Collection of all archetype definitions. An archetype = unique set of component types.
- **Tables** — Column-oriented storage for Table-stored components. One table per archetype.
- **Sparse Sets** — Entity-indexed storage for SparseSet-stored components.
- **Resources** — Type-keyed singleton storage.
- **Schedules** — Named schedule definitions registered with the World.
- **Change Tick** — Global monotonic counter incremented on every system run.
- **Component Registry** — Metadata for all registered component types.

### 4.2 Entity Operations

```
World.Spawn(components...) -> Entity
World.SpawnEmpty() -> Entity
World.Despawn(entity)
World.Entity(entity) -> EntityRef        // read-only access
World.EntityMut(entity) -> EntityMut     // read-write access
World.Contains(entity) -> bool
```

### 4.3 Resource Operations

```
World.InsertResource(resource)
World.InitResource[T]()                  // insert default
World.GetResource[T]() -> *T
World.RemoveResource[T]()
World.ContainsResource[T]() -> bool
```

### 4.4 Schedule Execution

```
World.AddSchedule(label, schedule)
World.RunSchedule(label)                 // runs all systems in the named schedule
World.RunSystemOnce(system)              // run a single system immediately
```

### 4.5 DeferredWorld

A restricted view of the World available inside component hooks and observers. Prevents re-entrant archetype mutations that would invalidate internal pointers.

Allowed: read/write components on existing entities, send events, access resources.
Forbidden: spawn, despawn, add/remove components (these must go through Commands).

### 4.6 Multiple Worlds

Sub-applications (e.g., render pipeline) run their own World. Data transfer between Worlds happens via an explicit "extract" phase — a function that copies selected data from the main World to the sub-World.

### 4.7 Change Tick Management

- Each World has a `last_change_tick` and `current_change_tick`.
- When a system runs, it receives the range `[last_run_tick, current_tick]`.
- `ClearTrackers()` advances the tick and resets per-frame change flags.
- Called once per update cycle, typically at the end of the frame.

## 5. Open Questions

- Should the World expose direct archetype iteration for advanced use cases, or keep it behind Query?
- Memory budget: should the World support pre-allocation hints for expected entity counts?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft |
