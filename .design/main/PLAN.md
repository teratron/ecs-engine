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
