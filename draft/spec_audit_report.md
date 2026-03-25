# Specification Audit Report

**Date:** 2026-03-25
**Scope:** All 30 specs in `.design/main/specifications/`
**Auditor:** Antigravity (this session)

## Summary

| Check | Result | Details |
| :--- | :--- | :--- |
| Files on disk | ✅ **30 files** | All 30 specs exist |
| INDEX.md sync | ✅ **29/30 registered** | All 29 domain specs in INDEX.md. `examples-framework.md` listed in P6. |
| Version sync | ✅ | All files v0.1.0 except `examples-framework.md` (v0.2.0) — matches INDEX |
| Status sync | ✅ | All Draft — matches INDEX |
| Layer | ✅ | All `concept` (L1 only, no L2 yet) — correct |
| **C25 Violation** | ⚠️ **14 files** | Say "from Bevy analysis" in Document History |
| Template compliance | ⚠️ **Partial** | See details below |

## C25 Violation: External Branding (14 files)

These files contain "Bevy" in Document History — violates **C25 §3** (No External Branding):

| File | Line | Content |
| :--- | :--- | :--- |
| `animation-system.md` | 209 | "Initial draft from Bevy analysis" |
| `app-framework.md` | 298 | "Initial draft from Bevy analysis" |
| `command-system.md` | 105 | "Initial draft from Bevy analysis" |
| `component-system.md` | 157 | "Initial draft from Bevy analysis" |
| `diagnostic-system.md` | 225 | "Initial draft from Bevy analysis" |
| `entity-system.md` | 112 | "Initial draft from Bevy analysis" |
| `event-system.md` | 157 | "Initial draft from Bevy analysis" |
| `hierarchy-system.md` | 127 | "Initial draft from Bevy analysis" |
| `post-processing.md` | 170 | "Initial draft from Bevy analysis" |
| `query-system.md` | 124 | "Initial draft from Bevy analysis" |
| `system-scheduling.md` | 174 | "Initial draft from Bevy analysis" |
| `ui-system.md` | 269 | "Initial draft from Bevy analysis" |
| `window-system.md` | 203 | "Initial draft from Bevy analysis" |
| `world-system.md` | 114 | "Initial draft from Bevy analysis" |

**Fix:** Replace "from Bevy analysis" → "Initial draft" in all 14 files.

## Template Compliance

### Spec template vs actual structure

| Template Section | Present In | Missing From |
| :--- | :--- | :--- |
| `## Overview` | All 30 | — |
| `## Related Specifications` | All 30 | — |
| `## 1. Motivation` | All 30 | — |
| `## 2. Constraints & Assumptions` | All 30 | — |
| `## 3. Core Invariants` | All 30 | — |
| `## 4. Invariant Compliance` | 0 | All (expected — L1 doesn't need this) |
| `## 5. Detailed Design` / `## 4. Detailed Design` | 30 | — |
| `## 6. Implementation Notes` | 1 (examples-framework) | 29 specs |
| `## 7. Drawbacks & Alternatives` | 1 (examples-framework) | 29 specs  |
| `## Document History` | All 30 | — |

> [!NOTE]
> The 29 domain specs use `## 4. Detailed Design` (numbered 4) instead of template's `## 5. Detailed Design` (numbered 5).
> This is because they skip `## 4. Invariant Compliance` (L2 only section), shifting numbers.
> Also they replace `## 6+7` with `## 5. Open Questions` — a useful addition not in the template.

### Section numbering deviation

Two patterns observed:

**Pattern A** (29 domain specs):

```
## 1. Motivation
## 2. Constraints & Assumptions
## 3. Core Invariants
## 4. Detailed Design        ← skipped §4 Invariant Compliance (correct for L1)
## 5. Open Questions          ← replaces §6 Implementation Notes + §7 Drawbacks
```

**Pattern B** (examples-framework.md):

```
## 1. Motivation
## 2. Constraints & Assumptions
## 3. Core Invariants
## 5. Detailed Design         ← skipped §4, numbered as §5
## 6. Implementation Notes
## 7. Drawbacks & Alternatives
```

> [!IMPORTANT]
> Both patterns are acceptable for L1 specs since `§4 Invariant Compliance` is explicitly L2-only.
> However, `examples-framework.md` jumps from §3 to §5 (no §4), while domain specs go §3 → §4.
> **Recommendation:** Standardize on Pattern A (§3 → §4) for all L1 specs.

## Content Quality (Sampled)

| Spec | Lines | Quality Assessment |
| :--- | :--- | :--- |
| `world-system.md` | 115 | ✅ Good: clear invariants, pseudo-code API, DeferredWorld concept |
| `render-core.md` | 106 | ✅ Good: render graph DAG, backend abstraction, phase sorting |
| `math-system.md` | 213 | ✅ Excellent: comprehensive types, curves, color, invariants |
| `input-system.md` | 205 | ✅ Excellent: ButtonInput[T] generic, pointer abstraction, picking |
| `state-system.md` | 216 | ✅ Excellent: SubStates, ComputedStates, DespawnOnExit, run conditions |
| `examples-framework.md` | 543 | ✅ Comprehensive: 280+ examples, lifecycle, CI, mapping |

## INDEX.md Structure

INDEX.md v2.0.0 organizes specs into 6 priority batches:

| Batch | Specs | Coverage |
| :--- | :--- | :--- |
| P1 ECS Core | 7 | world, entity, component, query, system-scheduling, command, event |
| P2 Framework | 6 | hierarchy, time, input, state, change-detection, app-framework |
| P3 Assets & Math | 4 | task, asset, scene, math |
| P4 Render Pipeline | 5 | render-core, mesh-and-image, materials-and-lighting, camera-and-visibility, post-processing |
| P5 Content Systems | 4 | audio, asset-formats, 2d-rendering, animation |
| P6 UI & Tools | 4 | window, diagnostic, ui, examples-framework |

> [!NOTE]
> This is a different categorization than the original proposal (26 specs in 4 priorities).
> The 29 domain specs were recut into more focused topics (e.g., `render-pipeline` split into 5 specs).

## Actionable Fixes

1. **[CRITICAL] Fix C25 violations** — Remove "Bevy" from 14 Document History entries
2. **[MINOR] Standardize section numbering** — Decide on one pattern for L1 specs
3. **[MINOR] `examples-framework.md` §5 numbering** — Renumber §5→§4 for consistency
