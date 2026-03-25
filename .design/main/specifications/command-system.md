# Command System

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

Commands are deferred mutations to the World. Systems cannot directly perform structural changes (spawn, despawn, add/remove components) because these would invalidate active query iterators. Instead, systems enqueue commands into a buffer, which is applied at defined synchronization points.

## Related Specifications

- [world-system.md](world-system.md) — Commands are applied to the World
- [entity-system.md](entity-system.md) — Spawn/despawn commands create/destroy entities
- [component-system.md](component-system.md) — Insert/remove commands modify entity components
- [system-scheduling.md](system-scheduling.md) — ApplyDeferred sync points in schedule

## 1. Motivation

During parallel system execution, direct World mutation would cause data races and iterator invalidation. Commands solve this by:
- Buffering all structural changes.
- Applying them atomically at safe synchronization points.
- Allowing systems to "fire and forget" mutations without worrying about ordering.

## 2. Constraints & Assumptions

- Commands are per-system buffers, merged and applied sequentially.
- Command application order within a buffer is FIFO (first-in, first-out).
- Commands from different systems are applied in system execution order.
- Command application can trigger component hooks and observers.

## 3. Core Invariants

- **INV-1**: Commands are never applied during system execution — only at ApplyDeferred points.
- **INV-2**: Command buffers are flushed completely at each sync point — no partial application.
- **INV-3**: Entity IDs reserved by Commands are valid immediately (can be used in subsequent commands in the same buffer).
- **INV-4**: A command that references a despawned entity is a no-op (not an error).

## 4. Detailed Design

### 4.1 Built-in Commands

| Command | Description |
| :--- | :--- |
| `Spawn(components...)` | Create new entity with components |
| `SpawnEmpty()` | Create entity with no components |
| `Despawn(entity)` | Destroy entity and all its components |
| `Insert(entity, component)` | Add/overwrite component on entity |
| `Remove[T](entity)` | Remove component T from entity |
| `InsertResource(resource)` | Insert or overwrite a global resource |
| `RemoveResource[T]()` | Remove a resource |
| `SendEvent(event)` | Enqueue an event |
| `Trigger(event)` | Trigger an observer event |
| `RunSystem(system)` | Schedule a one-shot system to run |

### 4.2 Entity Commands (Builder Pattern)

```
commands.Spawn(PlayerBundle{...})
    .Insert(Health{100})
    .Insert(Name{"Player 1"})
    .WithChildren(func(parent) {
        parent.Spawn(WeaponBundle{...})
    })
```

`Spawn` returns a handle to the not-yet-created entity. Subsequent `.Insert()` calls append to the same command buffer entry. The entity ID is reserved immediately and can be stored.

### 4.3 Custom Commands

Users can define custom commands implementing a `Command` interface:

```
interface Command {
    Apply(world *World)
}
```

Custom commands have full `&mut World` access when applied. Use for complex operations that touch multiple entities or resources atomically.

### 4.4 Command Queue Lifecycle

1. System runs → writes commands to its per-system `CommandBuffer`.
2. System completes → buffer is sealed.
3. `ApplyDeferred` sync point reached → all pending buffers applied in system order.
4. Each command's `Apply(world)` runs sequentially.
5. Buffers cleared after application.

### 4.5 Entity Reservation

When a system calls `commands.Spawn(...)`, the entity ID is reserved immediately from the allocator. This means:
- The ID can be used in subsequent commands within the same system.
- The entity does not physically exist in the World until commands are applied.
- Other systems cannot see the entity until after ApplyDeferred.

## 5. Open Questions

- Should commands support rollback/undo for editor integration?
- Priority commands: should some commands apply before others regardless of system order?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft |
