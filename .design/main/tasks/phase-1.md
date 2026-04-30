---
phase: 1
name: "ECS Core POC"
status: In Progress
subsystem: "internal/ecs"
requires: []
provides:
  - "World runtime (entities, components, archetypes, resources)"
  - "Bitmask-matched query layer (128-bit archetype masks)"
  - "DAG scheduler with sequential executor"
  - "Command buffer + event bus"
  - "Type registry for runtime introspection"
  - "examples/ecs/poc/ — C29 unblock entry"
key_files:
  created: []
  modified: []
patterns_established: []
duration_minutes: ~
bootstrap: true
---

# Stage 1 Tasks — ECS Core POC

**Phase:** 1
**Status:** In Progress
**Strategic Goal:** Land a runnable ECS runtime in `internal/ecs` and `pkg/ecs`, validated end-to-end by `examples/ecs/poc/`. Successful completion unblocks C29 and promotes the P1 spec cohort `Draft → Stable`.

## Track Overview

| Track | Domain | Critical-Path | Tasks |
| :--- | :--- | :---: | :--- |
| A | Entity (`internal/ecs/entity/`) | — | T-1A01..03 |
| B | Component (`internal/ecs/component/`) | **Yes** | T-1B01..03 |
| C | World (`internal/ecs/world/`) | Yes | T-1C01..03 |
| D | Query (`internal/ecs/query/`) | — | T-1D01..03 |
| E | Scheduler (`internal/ecs/scheduler/`) | — | T-1E01..03 |
| F | Command (`internal/ecs/command/`) | — | T-1F01..02 |
| G | Event (`internal/ecs/event/`) | — | T-1G01..02 |
| H | Type Registry (`internal/ecs/typereg/`) | — | T-1H01..02 |
| I | Lifecycle Patterns (cross-cutting) | — | T-1I01..02 |
| T | Validation (`pkg/ecs/`, `examples/ecs/poc/`) | — | T-1T01..05 |

Critical path: **B → C → D**. Tracks A, E, F, G, H, I are file-independent and parallelizable.

## Atomic Checklist

### Track A — Entity

- [x] [T-1A01] Implement `EntityID` (uint64, 32-bit index / 32-bit generation) and `Entity` value type. — `internal/ecs/entity/entity.go` + tests (100% coverage). [Bootstrap]
- [x] [T-1A02] Implement `EntityAllocator`: free-list + generation increment on reuse. — `internal/ecs/entity/allocator.go` + tests (98.6% coverage). [Bootstrap]
- [x] [T-1A03] Implement `EntitySet` / `EntityMap` helpers and `DisabledTag` empty-struct component. — `internal/ecs/entity/{set,tags}.go` + tests (99.2% coverage). Track A complete. [Bootstrap]

### Track B — Component (Critical Path)

- [x] [T-1B01] Implement `ComponentRegistry`: type→`ComponentID` allocation, deterministic ordering for archetype hashing. — `internal/ecs/component/{component,registry}.go` + tests (97.6% coverage). [Bootstrap]
- [x] [T-1B02] Implement storage strategies: chunk-based `Table` (16 KB blocks) primary path; `SparseSet` fallback for `StorageSparseSet`-tagged components. — `internal/ecs/component/{column,sparseset,table}.go` + tests (97.2% coverage). [Bootstrap]
- [x] [T-1B03] Implement `OnAdd`/`OnRemove` hooks, required-component graph, bundle insertion. — `internal/ecs/component/{hooks,bundle,required}.go` + tests (95.7% pkg coverage). Track B complete. [Bootstrap]

### Track C — World

- [x] [T-1C01] Implement `World` struct: entities, components, `ResourceMap` (typed singletons), monotonic change tick. — `internal/ecs/world/{world,resource}.go` + tests (100% coverage). [Bootstrap]
- [x] [T-1C02] Implement `DeferredWorld` view + apply points (consumed by Track F). — `internal/ecs/world/deferred.go` + tests (100% coverage). [Bootstrap]
- [x] [T-1C03] Implement archetype graph + entity migration on component add/remove. — `internal/ecs/world/{archetype,entity_ops}.go` + Table/Registry extras (`CellPtrByID`, `RowValues`, `SetCellByID`, `RegisterByType`). Spawn/Insert/Remove/Get with archetype migration; required components auto-injected. component 96.6%, world 94.7% coverage. Track C complete. [Bootstrap]

### Track D — Query

- [x] [T-1D01] Implement `QueryState` with 128-bit bitmask archetype matching and `Access` tracking (read/write/exclusive). — `internal/ecs/query/{mask,access,query}.go` + tests (100% coverage). [Bootstrap]
- [x] [T-1D02] Implement multi-arity generics `Query1[T]`, `Query2[T,U]`, `Query3[T,U,V]` with `iter.Seq2` traversal. — `internal/ecs/query/{tuple,resolver,query1,query2,query3}.go` + tests (97.5% pkg coverage). Cross-pkg cache via `world.ArchetypeStore.Each/EachFrom/At` and `world.SparseSet(id)` accessors. [Bootstrap]
- [x] [T-1D03] Implement filters (`With`, `Without`, `Added`, `Changed`) and `ParIter` scaffold (work-stealing deferred to Phase 3). — `internal/ecs/query/{filter,par}.go` + tests (96.4% pkg coverage, `-race` clean). Track D complete. [Bootstrap]

### Track E — Scheduler

- [x] [T-1E01] Implement `System` interface + `Schedule` (DAG topology built from `Access`). — `internal/ecs/scheduler/{system,dag,schedule}.go` + tests (98.1% pkg coverage). Kahn's algorithm; deterministic order on ties; explicit Before/After + implicit edges from `query.Access` conflicts. [Bootstrap]
- [x] [T-1E02] Implement sequential executor (single-goroutine) sufficient for POC validation. — `internal/ecs/scheduler/executor.go` + tests (98.2% pkg coverage). Panic recovery via `ErrSystemPanic`; `ErrScheduleNotBuilt` guard; `Schedule.Run(world)` convenience wrapper. [Bootstrap]
- [x] [T-1E03] Implement `RunCondition` predicates and `SystemSet` grouping. — `internal/ecs/scheduler/{condition,set}.go` + tests (98.9% pkg coverage). `RunCondition`/Not/And/Or combinators; `Schedule.ConfigureSet` with set-level RunIf+Before+After expanding to pairwise edges at Build; executor evaluates own + set conditions before each system. Track E complete. [Bootstrap]

### Track F — Command

- [x] [T-1F01] Implement `Command` interface + `CommandBuffer` with `sync.Pool` reuse (C27). — `internal/ecs/command/{command,buffer,builtin,param}.go` + tests (100% pkg coverage). Built-ins: SpawnEmpty/Spawn/Despawn/Insert/Remove/Custom; `Commands`/`EntityCommands`/`ChildSpawner` builder API; `AcquireBuffer`/`ReleaseBuffer` pool round-trip. World extended with `SpawnWithEntity`/`SpawnWithEntityAndData`/`RemoveByID` (world 96.1% coverage, `-race` clean). **BenchmarkCommandFlush: 0 B/op, 0 allocs/op** — C27 ≤1 alloc/op satisfied.
- [ ] [T-1F02] Implement entity reservation (pre-allocated IDs) and flush-at-apply-point semantics.

### Track G — Event

- [ ] [T-1G01] Implement `EventBus` + typed `MessageChannel` (double-buffered, drained per tick).
- [ ] [T-1G02] Implement observer registration + entity event bubbling along `ChildOf` chains.

### Track H — Type Registry

- [ ] [T-1H01] Implement `TypeRegistry` + `FieldInfo` via `reflect`; cache lookups.
- [ ] [T-1H02] Implement `DynamicObject` + serialization-hook contract (no actual codecs in Phase 1).

### Track I — Lifecycle Patterns

- [ ] [T-1I01] Implement bitmask tagging utilities + cached query views (subscribe to archetype graph deltas).
- [ ] [T-1I02] Implement object pool primitives for short-lived components and command payloads (`sync.Pool` wrappers).

### Track T — Validation

- [ ] [T-1T01] `pkg/ecs/ecs_test.go`: `BenchmarkSpawn`, `BenchmarkIter1`, `BenchmarkIter3` with `-benchmem`; baseline thresholds documented.
- [ ] [T-1T02] Race tests for Scheduler + EventBus + CommandBuffer (CI gate `go test -race ./...`).
- [ ] [T-1T03] Fuzz: `FuzzComponentRegistry` (registration ordering), `FuzzEntityID` (encode/decode round-trip).
- [ ] [T-1T04] Golden test: deterministic archetype migration order across 1000-entity churn.
- [ ] [T-1T05] **C29 unblock** — `examples/ecs/poc/` end-to-end: spawn 10k entities, run a 3-system schedule for N ticks, assert outcomes; documented in `Document History` of every P1 spec.

## Detailed Tracking

### [T-1A01] EntityID layout

- **Spec:** [l2-entity-system-go.md](../specifications/l2-entity-system-go.md) §3
- **Status:** Done [Bootstrap]
- **Assignment:** Agent
- **Handoff:** Required by T-1A02 (allocator), T-1B01 (registry component-of-entity check), T-1C01 (world).
- **Notes:** Use `uint64` packed; expose `Index()`/`Generation()` accessors. Stack-only struct, no heap allocations.
- **Changes:** Added `internal/ecs/entity/entity.go` with `EntityID` (uint64 packed), `Entity` value type, `NewEntityID`/`NewEntity`/`FromID` constructors and accessors. Tests: 100% coverage, fuzz target `FuzzEntityIDRoundTrip`, size assertion (8 bytes). `-race` deferred — local toolchain lacks CGO/gcc; CI gate (T-1T02) will enforce.

### [T-1B02] Storage strategies (Critical Path)

- **Spec:** [l2-component-system-go.md](../specifications/l2-component-system-go.md), [l1-ecs-lifecycle-patterns.md](../specifications/l1-ecs-lifecycle-patterns.md)
- **Status:** Done [Bootstrap]
- **Assignment:** Agent — strongest contributor (C26 cascade risk).
- **Handoff:** Unblocks T-1C03 (archetype migration), T-1D01 (bitmask matching), T-1I02 (pooling).
- **Notes:** Chunk size = 16 KB; document layout in ADR-001 (T-1T05 references). Sparse-set fallback only when `StorageSparseSet` tag is registered.
- **Changes:** Added `column.go` (ColumnSpec + alignment-desc sort utility), `sparseset.go` (per-type dense+sparse, swap-and-pop, zero-size tag support), `table.go` (chunked 16 KB SOA layout, columns sorted by Align desc + Size desc + ID asc for determinism, AddRow/SetCell/CellPtr/RemoveRow swap-and-pop, trailing-empty-chunk release on Remove). Layout math packs columns sequentially with `alignUp` padding; row stride exceeding chunk size panics at construction. Tag-only archetypes skip physical chunk allocation. ADR-001 not yet authored — to be drafted before T-1T05.

### [T-1C03] Archetype graph

- **Spec:** [l1-world-system.md](../specifications/l1-world-system.md) §3, [l2-world-system-go.md](../specifications/l2-world-system-go.md)
- **Status:** Todo
- **Handoff:** Required by T-1D01 (queries iterate archetypes), T-1G02 (observer graph).
- **Notes:** Use `unique.Handle[ComponentSet]` for archetype identity (per C24 stdlib priority).

### [T-1D01] Bitmask matching

- **Spec:** [l2-query-system-go.md](../specifications/l2-query-system-go.md)
- **Status:** Done [Bootstrap]
- **Handoff:** Unblocks T-1D02, T-1D03; required by Phase 2 change-detection filters.
- **Notes:** 128-bit mask covers ~128 component types in POC. Document upgrade path to dynamic-width masks for Phase 2+.
- **Changes:** Added `internal/ecs/query/{mask,access,query}.go`. `Mask` (lo/hi uint64 pair) with Set/Clear/Has/Equal/Contains/IsDisjoint/Intersects/Or/And/AndNot/Count/ForEach/IDs/String — IDs ≥128 panic on Set/Clear, return false on Has. `Access{Read, Write, Exclusive}` (each a Mask) with Conflicts (Read-Read OK; Write vs anything and Exclusive vs anything conflict), Merge, Validate (rejects Exclusive ∩ Read/Write — Read+Write overlap is allowed). `QueryState` holds required/excluded masks + access; `NewQueryState` auto-promotes required IDs to Read unless already Write/Exclusive. 100% coverage.

### [T-1E02] Sequential executor

- **Spec:** [l2-system-scheduling-go.md](../specifications/l2-system-scheduling-go.md)
- **Status:** Todo
- **Handoff:** Required by T-1T01 (benchmarks need a runnable schedule). Parallel executor is intentionally **out of scope** for Phase 1.

### [T-1F01] CommandBuffer + sync.Pool

- **Spec:** [l2-command-system-go.md](../specifications/l2-command-system-go.md), C27
- **Status:** Todo
- **Notes:** Validation: `BenchmarkCommandFlush` MUST show ≤1 alloc/op after warm-up.

### [T-1T05] examples/ecs/poc — C29 unblock

- **Spec:** [l1-examples-framework.md](../specifications/l1-examples-framework.md), all P1 specs (referenced from `Document History`)
- **Status:** Todo
- **Goal:** End-to-end validation that exercises every P1 spec. Output of this task is the gate that lets `magic.spec` promote the P1 cohort `Draft → Stable`.
- **Method:** `go run ./examples/ecs/poc` produces deterministic output; `go test -race ./examples/ecs/poc` passes.
- **Acceptance:**
  1. 10k entities spawned across 3 archetypes.
  2. 3 systems registered in a DAG, executed for 100 ticks.
  3. Commands and events round-trip through one tick boundary.
  4. Bench delta vs T-1T01 baseline ≤ +0%.
  5. `Document History` updated in every P1 spec referencing the POC path.

## Validation Strategy

- **Per-track local tests** (table-driven, `_test.go`) land alongside each implementation task.
- **Cross-track integration** is gated by Track T (`T-1T*`).
- **CI Gates** (mandatory before phase Done):
  - `go vet ./...`
  - `golangci-lint run`
  - `go test -race ./...`
  - `go test -bench=. -benchmem ./pkg/ecs/...` with regression check vs. baseline.

## Exit Criteria

Phase 1 is `Done` when **all** of:

1. Every atomic task above is `[x]`.
2. CI gates green on `master`.
3. `examples/ecs/poc/` runs deterministically (T-1T05).
4. `magic.spec` promotes the 17 P1 specs `Draft → Stable` (C29 unblocked).
5. STATE.md `Phase` advances to `2 — Framework Primitives` and `Status: Active`.
