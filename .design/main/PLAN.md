# Implementation Plan — ECS Engine core

## Phase 1: High-Performance Foundations (Active)

### 1.1 Task System & Concurrency

- [ ] Implement `TaskPool` with `TryCooperate` main-thread assistance
- [ ] Implement `ParallelFor` with adaptive batching

### 1.2 Component System & Data Layout

- [ ] Implement Sparse-Set storage for core components
- [ ] Implement `AssociatedDataMap` for system-owned technical caches

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
- [ ] `CollisionGroups` bitfield filtering
- [ ] Trajectory prediction logic (ShapeCast)
- [ ] Debug visualisation for physics (Gizmos)
- [ ] Implement `ImpulseBackend` (pure Go deterministic solver)

### 4.2 Scripting & Automation

- [ ] Implement Lua bridge with `TypeRegistry` binding generator
- [ ] Implement `Go Hot-Reload` via `-overlay` or plugin package

### 4.3 Diagnostics & Networking

- [ ] Implement `ProfilingProtocol` with Tracy span exporter
- [ ] Implement `NetworkPrimitives` for state snapshotting
