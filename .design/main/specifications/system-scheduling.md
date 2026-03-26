# System Scheduling

**Version:** 0.2.0
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

### 4.10 Automatic System Discovery

Systems can be auto-discovered and instantiated when the first entity matching their requirements appears, rather than requiring explicit registration:

```plaintext
// Component declares its default processor via attribute/tag
ComponentDescriptor {
    ...
    DefaultProcessor: TypeID   // optional — auto-create this system when component first appears
}
```

When the world encounters a component type for the first time:
1. Check if `DefaultProcessor` is set on its descriptor.
2. If set, instantiate the processor (if not already registered).
3. Recursively check the processor's required types for their own default processors.

This enables truly modular plugins: adding a component automatically brings in the system that processes it. Explicit registration (`AddSystems`) is still supported and takes precedence over auto-discovery.

### 4.11 Dual-Phase System Registration

Adding entities can trigger cascading processor discovery (§4.10), which can itself add more entities. To prevent ordering issues, system registration uses a two-phase approach:

```plaintext
addEntityLevel: int   // reentrancy counter

Phase 1 — Entity Addition (addEntityLevel > 0):
  New processors go into pendingProcessors list
  No immediate registration or initialization

Phase 2 — Flush (addEntityLevel returns to 0):
  RegisterPendingProcessors()
  - Sort by priority
  - Initialize each processor
  - Match existing entities against new processor
```

This prevents a processor from receiving half-initialized entities or from interfering with the entity-addition process that triggered its creation. The `addEntityLevel` counter handles recursive entity additions (e.g., a processor's initialization spawns more entities).

### 4.12 System Dependency Graph

Systems declare component type dependencies that the scheduler uses for automatic ordering and revalidation:

```plaintext
SystemDescriptor {
    MainComponentType:  TypeID       // primary component this system processes
    RequiredTypes:      []TypeID     // additional required components (entity must have ALL)
    DependentTypes:     []TypeID     // component types whose changes trigger revalidation
    Order:              int          // explicit priority (lower = earlier)
}
```

The scheduler builds a mapping: `ComponentType → []System`, tracking both direct processors and dependents. When a component changes on an entity:
1. Direct processors are notified to recheck the entity.
2. Dependent processors are notified to revalidate their cached data.

This automates what would otherwise require manual `After`/`Before` constraints for component-driven systems. Type matching supports interface satisfaction — a system requiring `Renderable` matches any component implementing that interface.

### 4.13 Flexible System Registration

An alternative registration model where components declare which system handles them via interface implementation:

```plaintext
// Component declares its processor type via interface
IProcessable[TSystem] interface {
    // Marker interface — presence triggers system auto-creation
}

// System is created on-demand when first IProcessable component appears
// System is destroyed when last component is removed (reference-counted)
```

This complements the traditional model (§4.1–4.6): both can coexist in the same world. The flexible model is better for optional, self-contained features (e.g., a particle effect component that brings its own processor). The traditional model is better for core engine systems with complex inter-system ordering.

## 5. Open Questions

- Should the engine support dynamic system addition/removal at runtime?
- How to handle system panics — skip and continue, or halt the schedule?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft |
| 0.2.0 | 2026-03-26 | Added automatic system discovery, dual-phase registration, dependency graph, flexible registration |
