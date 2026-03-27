# Implementation Plan — ECS Engine

**Status:** Phase 0: Architecture & Specifications (Active)
**STOP FACTOR:** No architectural expansion (L1/L2 specs) beyond P0 is permitted until Phase 1 (POC) is validated via code in `examples/`.

## Phase 0: Foundation & Specifications (CURRENT)

- [ ] Finalize Go-specific implementation specs (World-Go, Query-Go)
- [ ] Reach "Stable" status for the POC specification set
- [ ] Transition to POC implementation (tracked in .design/main/tasks/)

## Phase 1: High-Performance Foundations (Backlog)

- [ ] Implement `TaskPool` with `TryCooperate` main-thread assistance
- [ ] Implement `ParallelFor` with adaptive batching

### 1.2 Component System & Data Layout

- [ ] Implement Sparse-Set storage for core components
- [ ] Implement `AssociatedDataMap` for system-owned technical caches
- [ ] Implement `ComponentPool` for short-lived components (command-like)
- [ ] **Task 3: Archetype & Table Layout** (Assignee: Antigravity, Deadline: 2026-03-30, Deps: Task 2)
  - Implement `Column` (byte slice) and `Table` (map of columns).
  - Implement `Archetype` as a set of component IDs.
- [ ] **Task 4: Sparse-Set Fallback** (Assignee: Antigravity, Deadline: 2026-03-31, Deps: Task 3)
  - Implement `SparseSet` for components marked with `StorageSparseSet`.

### 1.3 System Scheduling & Discovery

- [ ] Implement DAG-based scheduler with parallel execution tracks

## Phase 2: Render SubApp & Pipeline

### 2.1 Multi-Phase Pipeline

- [ ] Implement Render SubApp with async extraction
- [ ] Implement `RenderFeature` extension points

### 2.2 Asset/Content Management

- [ ] Implement `FileSystemProvider` and `ZipProvider` for VFS

## Phase 3: Entity lifecycle & Scene tree

- [ ] Implement `SceneSpawner` with GUID-based prefab overrides
- [ ] Implement Entity cloning (Deep Copy)

## Phase 4: Advanced Engine Extensions (Backlog)

### 4.1 Physics Server

- [ ] Implement `PhysicsSubApp` with extraction/writeback cycle
- [ ] `RigidBody` component with axis-lock solver constraints
- [ ] `Collider` (Model B) & parent body syncing
- [ ] `PhysicsServer` query APIs (Ray/Shape/Point/Overlap)
- [ ] Implement `Joint` entity-based constraints & motors
- [ ] Implement `CollisionEvent` diffing & manifold snapshots
- [ ] Implement `PhysicsMaterial` assets & combine logic
- [ ] Implement `CharacterController` kinematic sweep & step-up
- [ ] `CollisionGroups` bitfield filtering
- [ ] Trajectory prediction logic (ShapeCast)
- [ ] Debug visualisation for physics (Gizmos)
- [ ] Implement `ImpulseBackend` (pure Go deterministic solver)

### 4.2 Scripting & Automation

- [ ] Implement Lua bridge with `TypeRegistry` binding generator
- [ ] Implement `Go Hot-Reload` via `-overlay` or plugin package
- [ ] **Task 1: Generational Entity ID** (Assignee: Antigravity, Deadline: 2026-03-28, Deps: None)
  - Implement `EntityID` as a `uint64` (32-bit index, 32-bit generation).
  - Implement `EntityAllocator` with a free-list and generation increment on reuse.
- [ ] **Task 2: World-Registry Basic** (Assignee: Antigravity, Deadline: 2026-03-29, Deps: Task 1)
  - Implement `World` struct with basic entity spawning and life-checks.

### 4.3 Diagnostics & Networking

- [ ] Implement `ProfilingProtocol` with Tracy span exporter
- [ ] Implement `NetworkPrimitives` for state snapshotting

### 4.4 Quality & Documentation

- [ ] Establish `ADR` (Architecture Decision Records) process
- [ ] Implement `Property-based testing` for core ECS invariants
- [ ] Implement `DevTools` state export (JSON) and graph visualization

### 4.5 Testing & Validation Scenarios
- [ ] Implement `Dino-Sapiens` test case (complex 50+ entity hierarchy with nested transforms)
- [ ] Implement `Stress-Physics` (10,000 colliding primitives)
