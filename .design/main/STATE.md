# Project State

<!-- STATE.md — live project memory. Read FIRST in every workflow session. -->
<!-- Maximum 100 lines. Agent updates AFTER each completed action. -->

**Workspace:** main
**Updated:** 2026-04-29 21:05
**Phase:** 1 — ECS Core POC
**Status:** Active

## Current Position

- **Task:** Track D (Query) — T-1D03 filters (With/Without/Added/Changed) + ParIter scaffold
- **Spec:** l2-query-system-go.md, l1-query-system.md
- **Next Action:** T-1D02 done. Next: T-1D03 — concrete QueryFilter implementations and a `ParIter` scaffold (work-stealing deferred to Phase 3).

## Progress

```
Phase 1: [11/27] ████░░░░ 41%
Overall: [11/27] ████░░░░ 41%
```

## Recent Decisions

<!-- Last 3-5 locked decisions. Older entries → archived to PLAN.md -->

- 2026-04-26 **Done:** T-1B02 — `internal/ecs/component/{column,sparseset,table}.go`. Table uses 16 KB physical chunks, columns sorted by Align desc → Size desc → ID asc (deterministic), SOA layout with `alignUp` padding, swap-and-pop on RemoveRow, releases trailing empty chunk. SparseSet covers tag-stored (zero-size) components and fallback storage. 97.2% coverage. ADR-001 (chunk layout) deferred to pre-T-1T05.
- 2026-04-26 **Done:** T-1B03 — `internal/ecs/component/{hooks,bundle,required}.go`. Hooks use a forward-declared `HookContext` (empty interface) to avoid circular import on world; concrete `*world.DeferredWorld` will satisfy it post-T-1C02. Required-component graph resolved at registration with three-state cycle detection (visiting set fires before typeToID short-circuit so already-registered nodes still trip on cycles). Bundle.Components flatten via reflect, supporting both value- and pointer-receiver implementations. 95.7% coverage. Track B complete.
- 2026-04-26 **Pattern:** Lifecycle hook signature is `func(HookContext, entity.Entity)` with `HookContext` as an opaque interface. Future world packages will type-assert to the concrete deferred world; this avoids import cycles between component and world.
- 2026-04-26 **Done:** T-1C01 — `internal/ecs/world/{world,resource}.go`. World struct (EntityAllocator + Registry + ResourceMap + Tick), Tick type with IsNewerThan, ResourceMap (sync.RWMutex, stores *T as any for stable mutable pointers), SpawnEmpty/Contains/Despawn, IncrementChangeTick/ClearTrackers, package-level generics SetResource/Resource/RemoveResource/ContainsResource. 100% coverage.
- 2026-04-26 **Pattern:** ResourceMap stores `*T` (wrapped as `any`) so Resource[T] returns a stable mutable pointer; mutation through the pointer is immediately visible on re-fetch.
- 2026-04-26 **Done:** T-1C02 — `internal/ecs/world/deferred.go`. DeferredWorld wraps *World, satisfies component.HookContext (compile-time assertion), exposes resource ops (DeferredResource/SetDeferredResource/RemoveDeferredResource/ContainsDeferredResource), IsAlive, and ApplyDeferred stub (wired by T-1F02). World.NewDeferred() convenience constructor. 100% coverage.
- 2026-04-26 **Done:** T-1C03 — `internal/ecs/world/{archetype,entity_ops}.go` + Table/Registry extensions. ArchetypeStore (sorted-IDs string key, generation counter), Archetype (componentIDs/table/entities/edges), entityRecord (archetypeID + row). Spawn/Insert/Remove/Get implemented with full archetype migration: extract RowValues from old table, swap-and-pop, AddRow into new archetype's table; SparseSet components survive migration without reload. Required components transitively auto-injected at Spawn. Track C complete. component 96.6%, world 94.7% coverage.
- 2026-04-26 **Pattern:** Empty archetype (ID 0) is the home of SpawnEmpty entities; Insert from empty archetype to the spawn target follows the same migration path as later transitions. Archetype keys = LE-encoded sorted component IDs as a string (`componentSetKey`).
- 2026-04-29 **Done:** T-1D01 — `internal/ecs/query/{mask,access,query}.go`. 128-bit `Mask` (lo/hi uint64) with full bitwise ops + ascending ForEach/IDs; mutators panic on id≥128, queries return false (asymmetric to catch programmer errors). `Access{Read, Write, Exclusive}` with Conflicts (Read-Read OK; Write conflicts with Read/Write; Exclusive conflicts with anything), Merge, Validate (Exclusive ∩ Read/Write rejected; Read+Write overlap allowed — Write supersedes). `QueryState` carries required/excluded masks + Access; NewQueryState auto-promotes required IDs to Read unless caller explicitly declared Write/Exclusive. 100% coverage.
- 2026-04-29 **Pattern:** Query primitives live in `internal/ecs/query/` and depend only on `internal/ecs/component` (for `ID`). Archetype-side caching of a per-archetype Mask is deferred to T-1D02 to avoid reaching into `world` from `query`; for now `MaskFromIDs(arch.ComponentIDs())` is the bridge.
- 2026-04-29 **Done:** T-1D02 — `internal/ecs/query/{tuple,resolver,query1,query2,query3}.go`. Query1[T]/Query2[A,B]/Query3[A,B,C] auto-register T via reflect, hold a per-query `nextScan` watermark over `world.ArchetypeStore` (append-only ⇒ watermark is sufficient — generation counter unused for now). Iteration via `iter.Seq2[Entity, *T]` / `Tuple2` / `Tuple3`; per-row `fetchComponent` dispatches Table.CellPtrByID vs SparseSet.Get based on `component.Info.Storage`. NewQuery2/3 reject duplicate type parameters. 97.5% pkg coverage. Added world accessors `ArchetypeStore.At/Each/EachFrom` and `World.SparseSet(id)` to enable cross-package iteration without exporting the underlying maps.
- 2026-04-29 **Pattern:** Query iteration uses Go 1.23 range-over-func. `Tuple2[A,B]{A,B *...}` carries pointer fields (not values) so callers can mutate in place. `iter.Seq2` is the only ergonomic shape for (Entity, payload) — multi-component queries pack the components into a Tuple struct because the language caps Seq2 at two values per step.

## Blockers

<!-- Empty if none. Format: [severity] description -->

<!-- (none) -->

## Blocking Constraints

<!-- Anti-patterns discovered through real failures. MANDATORY reading. -->
<!-- Agent MUST explicitly acknowledge each constraint before working. -->

- [C-001] **C29 Promotion Gate**: No P1 spec may be promoted Draft → Stable until `examples/ecs/poc/` validates the runtime end-to-end (T-1T05).
- [C-002] **STOP FACTOR (Phase ≥ 4 Hold)**: Phases 4–8 stay in `Hold` until Phase 1 POC is validated AND Phase 2 App Framework reaches `Stable`. No premature implementation work in those subsystems.
- [C-003] **C24 Stdlib Priority**: Engine core MUST have zero external Go deps. Any third-party package requires explicit justification recorded in an ADR.
- [C-004] **C27 GC Compensation**: Hot-path allocations (commands, events, transient views) MUST flow through `sync.Pool`. Validation Track verifies ≤1 alloc/op for `BenchmarkCommandFlush`.
- [C-005] **C28 Race Gate**: All concurrent tests MUST pass with `-race`; CI blocks otherwise.

## Session Continuity

**Last Session Ended:** 2026-04-25 13:27
**Handoff File:** none
**Bootstrap Mode:** true
