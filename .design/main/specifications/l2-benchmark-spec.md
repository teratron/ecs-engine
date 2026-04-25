# Benchmark Suite Specification

**Version:** 0.2.0
**Status:** Draft
**Layer:** test

## Overview

High performance is a core design goal of the engine. This specification defines the standardized benchmark suite used to measure performance, identify bottlenecks, and compare against existing ECS implementations in the Go ecosystem.

## Related Specifications

- [l1-ecs-lifecycle-patterns.md](l1-ecs-lifecycle-patterns.md) — Optimization patterns validated by the benchmark suite
- [l1-query-system.md](l1-query-system.md) — Target of iteration and query-build benchmarks
- [l1-system-scheduling.md](l1-system-scheduling.md) — Target of scheduler benchmarks
- [l1-build-tooling.md](l1-build-tooling.md) — CI pipeline orchestrates benchmark runs and regression gates

## 1. Motivation

Performance is not a feature that can be asserted — it must be measured, tracked, and regression-gated. Without a standardized benchmark suite:

- **Optimizations become anecdotal.** A PR claims "20% faster iteration" but the measurement method, hardware, and baseline are different from the last claim.
- **Regressions slip through.** An innocuous refactor adds an allocation per entity; the cumulative effect is only visible months later when a release fails its frame budget.
- **Positioning versus alternatives is impossible.** Users choosing between this engine, Arche, or bytearena/ecs have no common yardstick.

This specification fixes all three by codifying a fixed set of benchmarks (`BenchmarkIter1`, `BenchmarkSpawn`, `BenchmarkSchedule`, ...), a comparison methodology against named competitor libraries, and a CI integration so that every merge carries a signed performance delta. The suite is the ground truth for `l1-ecs-lifecycle-patterns` optimization claims and the gate that protects hot-path specifications from silent regressions.

## 2. Core Benchmarks

| Test | Name | Goal |
| :--- | :--- | :--- |
| **Simple Iteration** | `BenchmarkIter1` | Iterate 100k entities with 1 component. Measure cycles per entity. |
| **Complex Iteration** | `BenchmarkIter3` | Iterate 100k entities with 3 components (Position, Velocity, Acceleration). |
| **Structural Changes** | `BenchmarkSpawn` | Spawn 100k entities with random components. |
| **Despawn** | `BenchmarkDespawn` | Despawn all entities. |
| **Query Re-evaluation** | `BenchmarkQueryBuild` | Measure query creation and archetype matching overhead. |
| **System Scheduling** | `BenchmarkSchedule` | Measure DAG building and parallel execution overhead with 100 systems. |
| **Archetype Fragmentation** | `BenchmarkFragTest` | Iterate with 1000 different archetypes (diverse component combinations). |

## 3. Competitive Comparison

Measure against:

- [Arche](https://github.com/mlange-42/arche) (benchmark for archetype-based ECS)
- [go-ecs](https://github.com/bytearena/ecs)
- [engo/ecs](https://github.com/engoengine/ecs)

## 4. Tooling

- Use Go's `testing.B` for all measurements.
- Use `benchstat` for comparing results between versions.
- Profile memory allocations with `go test -benchmem`.
- Use pprof flame graphs to visualize hotspots.

## 5. Reporting

Benchmarks should be part of the CI pipeline. A performance report is generated after each major architectural change.

## 6. Implementation Strategy

1. **Suite Setup**: Define the benchmark runner and standard test scenarios.
2. **Execution**: Run benchmarks on a controlled environment (disabled turbo boost, high-performance power plan).
3. **Tracking**: Store history of benchmark results in `benchmarks/history/`.

## Canonical References

<!-- MANDATORY for Stable status. List authoritative source files that downstream agents
     MUST read before implementing this spec. Use relative paths from project root.
     Stub state — fill with concrete files when implementation begins (Phase 1+). -->

| Alias | Path | Purpose |
| :--- | :--- | :--- |

<!-- Empty table = no canonical sources yet. Populate one row per authoritative file
     when implementation lands (Phase 1+). Stable promotion requires ≥1 row. -->

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-27 | Initial draft |
| 0.2.0 | 2026-04-19 | Added `## Related Specifications` and `## 1. Motivation` sections; renumbered §1–§5 → §2–§6 (RULES §5/§6 compliance) |
