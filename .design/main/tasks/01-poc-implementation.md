# Task Breakdown — POC (Phase 1)

This document breaks down the Phase 1 goals into atomic, actionable tasks for AI assistance and core development.

## 1. Entity Lifecycle (P1.1)

- [ ] **Task 1: Generational Entity ID** (Assignee: Antigravity, Deadline: 2026-03-28, Deps: None)
  - Implement `EntityID` as a `uint64` (32-bit index, 32-bit generation).
  - Implement `EntityAllocator` with a free-list and generation increment on reuse.
- [ ] **Task 2: World-Registry Basic** (Assignee: Antigravity, Deadline: 2026-03-29, Deps: Task 1)
  - Implement `World` struct with basic entity spawning and life-checks.

## 2. Component Storage (P1.2)

- [ ] **Task 3: Archetype & Table Layout** (Assignee: Antigravity, Deadline: 2026-03-30, Deps: Task 2)
  - Implement `Column` (byte slice) and `Table` (map of columns).
  - Implement `Archetype` as a set of component IDs.
- [ ] **Task 4: Sparse-Set Fallback** (Assignee: Antigravity, Deadline: 2026-03-31, Deps: Task 3)
  - Implement `SparseSet` for components marked with `StorageSparseSet`.

## 3. Query System (P1.3)

- [ ] **Task 5: Basic Query Fetch** (Assignee: Antigravity, Deadline: 2026-04-01, Deps: Task 3)
  - Implement `QueryState1[T]` with archetype matching logic.
  - Implement `Iter` with `unsafe.Pointer` access to table data.

## 4. Concurrency & Scheduling (P1.4)

- [ ] **Task 6: DAG Build** (Assignee: Antigravity, Deadline: 2026-04-02, Deps: Task 5)
  - Implement topological sort for `Schedule` systems based on `Access` requirements.
- [ ] **Task 7: Sequential Executor** (Assignee: Antigravity, Deadline: 2026-04-03, Deps: Task 6)
  - Implement a basic executor that runs compatible systems in a single goroutine for validation.

## 5. Tooling & Quality (P1.5)

- [ ] **Task 8: Benchmark Entry**
  - Create `BenchmarkSpawn` and `BenchmarkIter1` in `pkg/ecs/ecs_test.go`.
- [ ] **Task 9: ADR-001**
  - Create the first Architecture Decision Record documenting the choice of Table-based storage for the hot-path.

## Strategy for AI

For each task, provide the corresponding specification file as context. The AI should generate implementation code in `internal/ecs/` or `pkg/ecs/` following the Go development rules.
