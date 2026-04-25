# Project State

<!-- STATE.md — live project memory. Read FIRST in every workflow session. -->
<!-- Maximum 100 lines. Agent updates AFTER each completed action. -->

**Workspace:** main
**Updated:** 2026-04-25 13:27
**Phase:** 1 — ECS Core POC
**Status:** Active

## Current Position

- **Task:** [T-1B01] ComponentRegistry — type→ComponentID allocation
- **Spec:** l2-component-system-go.md §3
- **Next Action:** Run `/magic.run main` to start Phase 1 — Track B (Component) is the critical path; begin with T-1B01.

## Progress

```
Phase 1: [0/27] ░░░░░░░░ 0%
Overall: [0/27] ░░░░░░░░ 0%
```

## Recent Decisions

<!-- Last 3-5 locked decisions. Older entries → archived to PLAN.md -->

- 2026-04-25 **Decision:** Force-Bootstrap regeneration of PLAN.md across 76 specs into 8 phases (user override of C6 Bootstrap precondition).
- 2026-04-25 **Pattern:** Phase workbooks use YAML frontmatter (`phase`, `status`, `requires`, `provides`, `bootstrap`) — read by `/magic.task` for dependency-aware planning.
- 2026-04-25 **Decision:** Track B (Component) is critical path; T-1B02 storage strategy gates 10 dependent tasks across 5 tracks.

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
