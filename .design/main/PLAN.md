# Implementation Plan — ECS Engine

**Status:** Phase 0: Architecture & Specifications (Active)
**STOP FACTOR:** No architectural expansion (L1/L2 specs) beyond **P1-P3 (Foundation)** is permitted until Phase 1 (POC) is validated via code in `examples/`.

## Phase 0: Foundation & Specifications (CURRENT)

- [ ] Finalize Go-specific implementation specs (World-Go, Query-Go)
- [ ] Reach "Stable" status for the POC specification set
- [ ] Transition to POC implementation (tracked in .design/main/tasks/)

## Phase 1: POC (Entity/Component/System) — GO 1.26.1

- [ ] Implement Generational Entity ID & Allocator
- [ ] Implement World-Registry & Basic Spawning
- [ ] Implement Archetype & Table Layout (Chunk-based)
- [ ] Implement Sparse-Set fallback storage
- [ ] Implement Basic Query System with bitmask matching
- [ ] Implement DAG-based system scheduler
- [ ] Implement Performance benchmarks & CI setup

## Phase 2: High-Performance Extensions

- [ ] Implement `TaskPool` with multi-threaded work-stealing
- [ ] Implement `ParallelFor` with adaptive batching
- [ ] Implement `AssociatedDataMap` for system-owned technical caches
- [ ] Implement `ComponentPool` for short-lived data

## Phase 3: Render SubApp & Pipeline

- [ ] Implement Render SubApp with async extraction
- [ ] Implement `RenderFeature` extension points
- [ ] Implement FileSystem/Asset Management (VFS, Zip)

## Phase 4: Physics, Scenes & Logic

- [ ] Implement Scene tree & Entity cloning
- [ ] Implement Physics Server (Collision, RigidBody, Constraints)
- [ ] Implement Character Controller (Kinematic Sweep)
- [ ] Implement Collision Events & Physics Material assets

## Phase 5: Diagnostics & Tools

- [ ] Implement Tracy Profiling exporter
- [ ] Implement Hot-Reload orchestrator (code hot-restart + shader hot-swap)
- [ ] Implement ADR process for architecture tracking
- [ ] Implement DevTools JSON export & Visualization
- [ ] Implement Dino-Sapiens & Stress-Physics test scenarios
