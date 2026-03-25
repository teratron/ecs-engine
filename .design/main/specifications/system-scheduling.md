# System Scheduling

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

Systems are functions that process components. The scheduling system organizes systems into named schedules, resolves ordering dependencies, detects conflicts, and executes systems — either sequentially or in parallel based on their data access patterns.

## Related Specifications

- [world-system.md](world-system.md) — Systems operate on World data
- [query-system.md](query-system.md) — Systems declare access via queries
- [command-system.md](command-system.md) — Systems enqueue deferred commands
- [event-system.md](event-system.md) — Systems communicate via events/messages
- [app-framework.md](app-framework.md) — Schedules registered in the App

## 1. Motivation

A game engine runs dozens to hundreds of systems per frame. The scheduler must:
- Maximize parallelism by running independent systems concurrently.
- Guarantee deterministic ordering where dependencies exist.
- Detect cycles and ambiguities at build time, not at runtime.
- Support conditional execution (run-if predicates).

## 2. Constraints & Assumptions

- Systems are plain functions with typed parameters (dependency injection).
- The scheduler builds a DAG (Directed Acyclic Graph) from ordering constraints.
- Cycle detection uses Tarjan's SCC algorithm at schedule build time.
- Systems within the same schedule share a single World.

## 3. Core Invariants

- **INV-1**: Systems with conflicting data access (write vs. read/write on same component) are never executed simultaneously.
- **INV-2**: Explicit ordering constraints (`Before`, `After`, `Chain`) are always respected.
- **INV-3**: Cycles in the dependency graph are detected at build time and reported as errors.
- **INV-4**: Deferred commands are applied at defined synchronization points, not during system execution.
- **INV-5**: A system that is skipped by a run condition still counts as "executed" for ordering purposes.

## 4. Detailed Design

### 4.1 System Types

#### Function Systems

Plain functions whose parameter types define their World access:

```
fn movement(query: Query[(&mut Position, &Velocity)], time: Res[Time]) {
    for (pos, vel) in query.Iter() {
        pos.x += vel.x * time.delta
    }
}
```

Parameters are automatically resolved via dependency injection from the World.

#### Exclusive Systems

Systems that take `&mut World` — full exclusive access. Cannot run in parallel with anything. Used for structural changes that cannot be deferred (e.g., batch entity migrations).

#### System Parameters

| Parameter | Access | Description |
| :--- | :--- | :--- |
| `Query[...]` | Read/Write | Component access |
| `Res[T]` | Read | Shared resource access |
| `ResMut[T]` | Write | Exclusive resource access |
| `Commands` | Write (deferred) | Deferred mutation buffer |
| `EventReader[T]` | Read | Read events from previous frame |
| `EventWriter[T]` | Write | Send events for next frame |
| `MessageReader[T]` | Read | Read system-to-system messages |
| `MessageWriter[T]` | Write | Send system-to-system messages |
| `Local[T]` | Write | Per-system local state (not shared) |

### 4.2 Schedules

A Schedule is a named, ordered collection of systems:

- **Main Schedule Order**: `First` → `PreUpdate` → `StateTransition` → `RunFixedMainLoop` → `Update` → `PostUpdate` → `Last`
- **Startup Schedules**: `PreStartup` → `Startup` → `PostStartup` (run once)
- **Fixed Timestep**: `FixedFirst` → `FixedPreUpdate` → `FixedUpdate` → `FixedPostUpdate` → `FixedLast` (run inside `RunFixedMainLoop`, may tick 0..N times per frame)

Users typically add systems to `Update` (gameplay) or `FixedUpdate` (physics).

### 4.3 System Sets

Named groups for organizing systems:

```
schedule.ConfigureSets(
    PhysicsSet.After(InputSet),
    RenderPrepSet.After(PhysicsSet),
)

schedule.AddSystems(Update,
    apply_gravity.InSet(PhysicsSet),
    detect_collisions.InSet(PhysicsSet).After(apply_gravity),
)
```

Sets support: `Before`, `After`, `Chain` (sequential ordering of all members), `RunIf` (conditional execution for entire set).

### 4.4 Ordering Constraints

- **Before(system/set)** — This system runs before the target.
- **After(system/set)** — This system runs after the target.
- **Chain()** — All systems in a tuple run sequentially in declaration order.
- **Ambiguity**: Two systems with conflicting access and no ordering = ambiguity warning.

### 4.5 DAG Building and Cycle Detection

At schedule build time:
1. Collect all systems and their ordering constraints.
2. Build a directed graph (system → dependency edges).
3. Run Tarjan's SCC algorithm to detect cycles.
4. Report cycles as build errors with the involved system names.
5. Topological sort for execution order.
6. Check for ambiguities: conflicting access without ordering.

### 4.6 Executors

#### Single-Threaded Executor

Runs systems sequentially in topological order. Used for debugging, deterministic replay, and platforms without threading.

#### Multi-Threaded Executor

Runs non-conflicting systems in parallel:
1. Maintain a ready queue of systems whose dependencies are satisfied.
2. For each ready system, check if its access set conflicts with currently running systems.
3. If no conflict, spawn the system on a worker thread.
4. When a system completes, update the dependency graph and check for newly ready systems.

### 4.7 Run Conditions

Systems can have conditions that control execution:

```
system.RunIf(resource_exists[GameState])
system.RunIf(in_state(GameState.Playing))
system.RunIf(on_timer(Duration.Seconds(1)))
```

Conditions are evaluated before the system runs. If false, the system is skipped but ordering constraints are still respected.

### 4.8 Apply Deferred (Sync Points)

Command buffers accumulate during system execution. They are applied at:
- Explicit `ApplyDeferred` sync points inserted between system sets.
- Automatically between sets that have structural dependencies.
- At the end of each schedule run.

During apply, all pending commands execute sequentially against the World.

### 4.9 System Stepping (Debug)

A debug tool that pauses schedule execution and allows stepping through systems one at a time. Useful for:
- Inspecting World state between system runs.
- Identifying which system introduces a bug.
- Not available in release builds.

## 5. Open Questions

- Should the engine support dynamic system addition/removal at runtime?
- How to handle system panics — skip and continue, or halt the schedule?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft |
