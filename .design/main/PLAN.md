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

## Backlog

- [ ] l1-2d-rendering.md
- [ ] l1-ai-assistant-system.md
- [ ] l1-animation-system.md
- [ ] l1-app-framework.md
- [ ] l1-asset-formats.md
- [ ] l1-asset-system.md
- [ ] l1-audio-system.md
- [ ] l1-build-tooling.md
- [ ] l1-camera-and-visibility.md
- [ ] l1-change-detection.md
- [ ] l1-character-controller.md
- [ ] l1-client-prediction.md
- [ ] l1-collider.md
- [ ] l1-collision-events.md
- [ ] l1-command-system.md
- [ ] l1-compatibility-policy.md
- [ ] l1-component-system.md
- [ ] l1-definition-system.md
- [ ] l1-diagnostic-system.md
- [ ] l1-ecs-lifecycle-patterns.md
- [ ] l1-entity-system.md
- [ ] l1-error-core.md
- [ ] l1-event-system.md
- [ ] l1-examples-framework.md
- [ ] l1-hierarchy-system.md
- [ ] l1-hot-reload.md
- [ ] l1-input-system.md
- [ ] l1-joints.md
- [ ] l1-lockstep.md
- [ ] l1-materials-and-lighting.md
- [ ] l1-math-system.md
- [ ] l1-mesh-and-image.md
- [ ] l1-multi-repo-architecture.md
- [ ] l1-network-diagnostics.md
- [ ] l1-networking-system.md
- [ ] l1-physics-materials.md
- [ ] l1-physics-query.md
- [ ] l1-physics-system.md
- [ ] l1-platform-system.md
- [ ] l1-post-processing.md
- [ ] l1-profiling-protocol.md
- [ ] l1-query-system.md
- [ ] l1-render-core.md
- [ ] l1-replication.md
- [ ] l1-rigid-body.md
- [ ] l1-rpc.md
- [ ] l1-scene-system.md
- [ ] l1-scripting-system.md
- [ ] l1-snapshot-interpolation.md
- [ ] l1-state-system.md
- [ ] l1-system-scheduling.md
- [ ] l1-task-system.md
- [ ] l1-time-system.md
- [ ] l1-transport.md
- [ ] l1-type-registry.md
- [ ] l1-ui-system.md
- [ ] l1-window-system.md
- [ ] l1-world-system.md
- [ ] l2-app-framework-go.md
- [ ] l2-benchmark-spec.md
- [ ] l2-change-detection-go.md
- [ ] l2-codegen-tools.md
- [ ] l2-command-system-go.md
- [ ] l2-component-system-go.md
- [ ] l2-entity-system-go.md
- [ ] l2-event-system-go.md
- [ ] l2-hierarchy-system-go.md
- [ ] l2-input-system-go.md
- [ ] l2-query-system-go.md
- [ ] l2-state-system-go.md
- [ ] l2-system-scheduling-go.md
- [ ] l2-time-system-go.md
- [ ] l2-type-registry-go.md
- [ ] l2-world-system-go.md
