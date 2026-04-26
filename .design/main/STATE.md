# Project State

<!-- STATE.md — live project memory. Read FIRST in every workflow session. -->
<!-- Maximum 100 lines. Agent updates AFTER each completed action. -->

**Workspace:** main
**Updated:** 2026-04-26 09:50
**Phase:** 1 — ECS Core POC
**Status:** Active

## Current Position

- **Task:** T-1C01 World struct (entities, components, ResourceMap, change tick) — next (Track C critical path)
- **Spec:** l2-world-system-go.md, l1-world-system.md §3
- **Next Action:** Run `/magic.run main`. Tracks A+B fully closed; recommended commit point. Track C is the next critical-path step (T-1C01 → T-1C02 → T-1C03), Track E/F/G/H/I parallelizable from here.

## Progress

```
Phase 1: [6/27] ██░░░░░░ 22%
Overall: [6/27] ██░░░░░░ 22%
```

## Recent Decisions

<!-- Last 3-5 locked decisions. Older entries → archived to PLAN.md -->

- 2026-04-25 **Decision:** Force-Bootstrap regeneration of PLAN.md across 76 specs into 8 phases (user override of C6 Bootstrap precondition).
- 2026-04-25 **Pattern:** Phase workbooks use YAML frontmatter (`phase`, `status`, `requires`, `provides`, `bootstrap`) — read by `/magic.task` for dependency-aware planning.
- 2026-04-25 **Decision:** Track B (Component) is critical path; T-1B02 storage strategy gates 10 dependent tasks across 5 tracks.
- 2026-04-26 **Done:** T-1A01 — `internal/ecs/entity/entity.go` (EntityID/Entity, 100% coverage). `-race` deferred to CI (no local CGO).
- 2026-04-26 **Done:** T-1A02 — `internal/ecs/entity/allocator.go` (freelist + generational reuse, AllocateMany batch, 98.6% coverage). Null sentinel preserved via gen-1 floor.
- 2026-04-26 **Done:** T-1A03 — `internal/ecs/entity/{set,tags}.go` (EntitySet swap-and-pop, EntityMap[V] generic, DisabledTag zero-size, 99.2% coverage). Track A complete.
- 2026-04-26 **Done:** T-1B01 — `internal/ecs/component/{component,registry}.go` (ID/Info/Registry, sequential IDs from 1, idempotent reregistration, deterministic ordering verified by fuzz, 97.6% coverage). Storage/hooks/required deferred to T-1B02/T-1B03.
- 2026-04-26 **Pattern:** Package naming follows RULES anti-stutter — `component.ID`/`Info`/`Registry` (not `ComponentID`); same will apply to `entity.*`, `world.*`, `query.*`.
- 2026-04-26 **Done:** T-1B02 — `internal/ecs/component/{column,sparseset,table}.go`. Table uses 16 KB physical chunks, columns sorted by Align desc → Size desc → ID asc (deterministic), SOA layout with `alignUp` padding, swap-and-pop on RemoveRow, releases trailing empty chunk. SparseSet covers tag-stored (zero-size) components and fallback storage. 97.2% coverage. ADR-001 (chunk layout) deferred to pre-T-1T05.
- 2026-04-26 **Done:** T-1B03 — `internal/ecs/component/{hooks,bundle,required}.go`. Hooks use a forward-declared `HookContext` (empty interface) to avoid circular import on world; concrete `*world.DeferredWorld` will satisfy it post-T-1C02. Required-component graph resolved at registration with three-state cycle detection (visiting set fires before typeToID short-circuit so already-registered nodes still trip on cycles). Bundle.Components flatten via reflect, supporting both value- and pointer-receiver implementations. 95.7% coverage. Track B complete.
- 2026-04-26 **Pattern:** Lifecycle hook signature is `func(HookContext, entity.Entity)` with `HookContext` as an opaque interface. Future world packages will type-assert to the concrete deferred world; this avoids import cycles between component and world.

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
