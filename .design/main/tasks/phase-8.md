---
phase: 8
name: "Physics & Scripting"
status: Hold
subsystem: "pkg/physics, pkg/scripting"
requires:
  - "Phase 4 Render Core Stable"
  - "Phase 3 Math System Stable"
provides:
  - "Physics server (deterministic solver, SubApp integration)"
  - "Rigid body + axis locks + sleep"
  - "Collision shapes + compound + mesh/convex + filters"
  - "Ray/Shape/Point/Overlap queries + batching"
  - "Joint constraints (Hinge, Piston, Ball, Distance, Fixed, Motorized)"
  - "Contact / Trigger events + manifolds + deferred despawn"
  - "Friction / Restitution assets + combine rules"
  - "Kinematic capsule character controller"
  - "Scripting bridge (Lua / Tengo, deferred)"
key_files:
  created: []
  modified: []
patterns_established: []
duration_minutes: ~
bootstrap: true
hold_reason: "Unfreezes after Phase 4 Render Core + Phase 3 Math Stable."
---

# Stage 8 Tasks — Physics & Scripting

**Phase:** 8
**Status:** Hold

## High-Level Checklist

- [ ] [T-8A] Physics server: solver, SubApp integration, interpolation. ([l1-physics-system.md](../specifications/l1-physics-system.md))
- [ ] [T-8B] Rigid body: mass, damping, axis locks, body types, sleep. ([l1-rigid-body.md](../specifications/l1-rigid-body.md))
- [ ] [T-8C] Collider: primitives, compound, mesh/convex, filters. ([l1-collider.md](../specifications/l1-collider.md))
- [ ] [T-8D] Physics query: Ray/Shape/Point/Overlap, batching, predicates. ([l1-physics-query.md](../specifications/l1-physics-query.md))
- [ ] [T-8E] Joints: Hinge, Piston, Ball, Distance, Fixed, Motorized. ([l1-joints.md](../specifications/l1-joints.md))
- [ ] [T-8F] Collision events: contact/trigger, manifolds, deferred despawn. ([l1-collision-events.md](../specifications/l1-collision-events.md))
- [ ] [T-8G] Physics materials: friction/restitution assets, combine rules, hot-reload. ([l1-physics-materials.md](../specifications/l1-physics-materials.md))
- [ ] [T-8H] Character controller: kinematic capsule, sweep, step-up, slope snap. ([l1-character-controller.md](../specifications/l1-character-controller.md))
- [ ] [T-8I] Scripting system: Lua/Tengo bridge, ECS API reference design (deferred). ([l1-scripting-system.md](../specifications/l1-scripting-system.md))
- [ ] [T-8T] Validation: deterministic physics step (CI hash gate), character-controller climb fixture.
