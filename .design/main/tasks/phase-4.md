---
phase: 4
name: "Render Pipeline"
status: Hold
subsystem: "pkg/render, internal/render"
requires:
  - "Phase 1 ECS Core"
  - "Phase 2 App Framework Stable"
  - "C29 unblocked (POC validated)"
provides:
  - "Render graph + extract pattern + render world"
  - "Mesh / image / texture atlas primitives"
  - "Material system + PBR + lighting + shadows"
  - "Camera + visibility + frustum culling"
  - "Post-processing chain (AA, tonemapping, bloom)"
key_files:
  created: []
  modified: []
patterns_established: []
duration_minutes: ~
bootstrap: true
hold_reason: "STOP FACTOR — unfreezes after Phase 1 POC validated (C29) and Phase 2 App Framework Stable."
---

# Stage 4 Tasks — Render Pipeline

**Phase:** 4
**Status:** Hold
**Strategic Goal:** Backend-agnostic render graph with Bevy-style extract pattern + render world isolation.

## High-Level Checklist

- [ ] [T-4A] Render graph + extract pattern + render world. ([l1-render-core.md](../specifications/l1-render-core.md))
- [ ] [T-4B] Mesh assets, vertex layout, image/texture, atlases. ([l1-mesh-and-image.md](../specifications/l1-mesh-and-image.md))
- [ ] [T-4C] Materials, PBR, lights, shadows, environment maps. ([l1-materials-and-lighting.md](../specifications/l1-materials-and-lighting.md))
- [ ] [T-4D] Camera, projections, visibility hierarchy, frustum culling. ([l1-camera-and-visibility.md](../specifications/l1-camera-and-visibility.md))
- [ ] [T-4E] Post-processing: AA, tonemapping, bloom. ([l1-post-processing.md](../specifications/l1-post-processing.md))
- [ ] [T-4T] Validation: golden-image diff harness, render-world isolation race tests.

## Hold Release Conditions

1. `examples/ecs/poc/` deterministic + benchmarks within baseline.
2. Phase 2 `l1/l2-app-framework` promoted to `Stable`.
3. L2 Go specs for render core authored (currently absent).
