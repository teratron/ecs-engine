# Master Task Index (Registry)

**Version:** 1.0.0
**Generated:** 2026-04-25
**Based on:** .design/main/PLAN.md v1.0.0
**Based on RULES:** .design/RULES.md v1.7.1
**Execution Mode:** Parallel (per C3)
**Status:** Active
**Mode:** `[Bootstrap]` — full Draft cohort, C29 unblock pending

## Overview

Tactical registry of all phases and their statuses. Atomic checklists live in per-phase workbooks under `tasks/phase-{N}.md`.

## Active Phases

| Phase | Description | Status |
| :--- | :--- | :--- |
| [Phase 1](tasks/phase-1.md) | ECS Core POC — world, entities, components, queries, scheduler, validating `examples/ecs/poc/` | In Progress |
| [Phase 2](tasks/phase-2.md) | Framework Primitives — hierarchy, time, input, state, change-detection, app/plugin | Todo |
| [Phase 3](tasks/phase-3.md) | Assets, Math & Concurrency — task pool, asset server, scene, math | Todo |
| [Phase 4](tasks/phase-4.md) | Render Pipeline — render graph, mesh, materials, camera, post-processing | Hold |
| [Phase 5](tasks/phase-5.md) | Content Systems — audio, asset codecs, 2D, animation, tweening | Hold |
| [Phase 6](tasks/phase-6.md) | UI, Tooling & Quality — definition, window, UI, build, CLI, platform, AI, examples, errors, benchmark, codegen | Hold |
| [Phase 7](tasks/phase-7.md) | Networking & Hot-Reload — profiling, transport, replication, sync, RPC, hot-reload | Hold |
| [Phase 8](tasks/phase-8.md) | Physics & Scripting — physics server, bodies, colliders, queries, joints, character, scripting | Hold |

## Archived Phases

| Phase | Description | Archive |
| :--- | :--- | :--- |
| Phase 0 (legacy) | POC implementation (pre-Bootstrap layout) | [archives/tasks/01-poc-implementation.md](archives/tasks/01-poc-implementation.md) |

## Cross-Phase Status Counters

- **Total atomic tasks (Phase 1 only)**: 27 (Tracks A–I + T)
- **Phases 2–8**: structural workbooks, atomic decomposition deferred to per-phase `/magic.task {workspace} "decompose phase-N"` invocations.

## Meta Information

- **Last Updated**: 2026-04-25
- **Maintainer**: Core Team
