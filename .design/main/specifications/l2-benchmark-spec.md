# Benchmark Suite Specification

**Version:** 0.1.0
**Status:** Draft
**Layer:** test

## Overview

High performance is a core design goal of the engine. This specification defines the standardized benchmark suite used to measure performance, identify bottlenecks, and compare against existing ECS implementations in the Go ecosystem.

## 1. Core Benchmarks

| Test | Name | Goal |
| :--- | :--- | :--- |
| **Simple Iteration** | `BenchmarkIter1` | Iterate 100k entities with 1 component. Measure cycles per entity. |
| **Complex Iteration** | `BenchmarkIter3` | Iterate 100k entities with 3 components (Position, Velocity, Acceleration). |
| **Structural Changes** | `BenchmarkSpawn` | Spawn 100k entities with random components. |
| **Despawn** | `BenchmarkDespawn` | Despawn all entities. |
| **Query Re-evaluation** | `BenchmarkQueryBuild` | Measure query creation and archetype matching overhead. |
| **System Scheduling** | `BenchmarkSchedule` | Measure DAG building and parallel execution overhead with 100 systems. |
| **Archetype Fragmentation** | `BenchmarkFragTest` | Iterate with 1000 different archetypes (diverse component combinations). |

## 2. Competitive Comparison

Measure against:
- [Arche](https://github.com/mlange-42/arche) (benchmark for archetype-based ECS)
- [go-ecs](https://github.com/bytearena/ecs)
- [engo/ecs](https://github.com/engoengine/ecs)

## 3. Tooling

- Use Go's `testing.B` for all measurements.
- Use `benchstat` for comparing results between versions.
- Profile memory allocations with `go test -benchmem`.
- Use pprof flame graphs to visualize hotspots.

## 4. Reporting

Benchmarks should be part of the CI pipeline. A performance report is generated after each major architectural change.

## 5. Implementation Strategy

1. **Suite Setup**: Define the benchmark runner and standard test scenarios.
2. **Execution**: Run benchmarks on a controlled environment (disabled turbo boost, high-performance power plan).
3. **Tracking**: Store history of benchmark results in `benchmarks/history/`.

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-27 | Initial draft |
