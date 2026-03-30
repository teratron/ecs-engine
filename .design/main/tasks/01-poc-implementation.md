# Task Breakdown — POC (Phase 1)

This document breaks down the Phase 1 goals into atomic, actionable tasks for AI assistance and core development.

## Sprint 1: Foundation (P1.1)

- [ ] **Task 1: Generational Entity ID** (Deadline: 2026-03-28, Deps: None)
  - Implement `EntityID` as a `uint64` (32-bit index, 32-bit generation).
  - Implement `EntityAllocator` with a free-list and generation increment on reuse.
- [ ] **Task 2: World-Registry Basic** (Deadline: 2026-03-29, Deps: Task 1)
  - Implement `World` struct with basic entity spawning and life-checks.

## Sprint 2: Context & Data (P1.2)

- [ ] **Task 3: Archetype & Table Layout** (Deadline: 2026-03-30, Deps: Task 2)
  - Implement `Column` (byte slice) and `Table` (map of columns).
  - Implement `Archetype` as a set of component IDs.
- [ ] **Task 4: Sparse-Set Fallback** (Deadline: 2026-03-31, Deps: Task 3)
  - Implement `SparseSet` for components marked with `StorageSparseSet`.

## Sprint 3: Execution & API (P1.3)

- [ ] **Task 5: Basic Query Fetch** (Deadline: 2026-04-01, Deps: Task 3)
  - Implement `QueryState1[T]` with archetype matching logic.
  - Implement `Iter` with `unsafe.Pointer` access to table data.

## Sprint 4: Advanced & Topology (P1.4)

- [ ] **Task 6: DAG Build** (Deadline: 2026-04-02, Deps: Task 5)
  - Implement topological sort for `Schedule` systems based on `Access` requirements.
- [ ] **Task 7: Sequential Executor** (Deadline: 2026-04-03, Deps: Task 6)
  - Implement a basic executor that runs compatible systems in a single goroutine for validation.

## 5. Tooling & Quality (P1.5)

- [ ] **Task 8: Benchmark Entry** (Deadline: 2026-04-03, Deps: Task 5)
  - Create `BenchmarkSpawn` and `BenchmarkIter1` in `pkg/ecs/ecs_test.go`.
- [ ] **Task 9: CI Infrastructure** (Deadline: 2026-04-05, Deps: None)
  - Setup GitHub Actions for `go test`, `go vet`, and `golangci-lint`.
  - Integrate markdown-lint for specification consistency.
- [ ] **Task 10: ADR-001** (Deadline: 2026-04-06, Deps: Task 3)
  - Create the first Architecture Decision Record documenting the choice of Table-based storage for the hot-path.
- [ ] **Task 11: Scaffold pkg/ boundaries** (Deadline: TBD, Deps: None)
  - Create pkg/editor/ and pkg/protocol/ with initial interface boundaries.

## Strategy for AI

For each task, provide the corresponding specification file as context. The AI should generate implementation code in `internal/ecs/` or `pkg/ecs/` following the Go development rules.
