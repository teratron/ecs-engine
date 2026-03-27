# Physics Queries

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

Physics queries allow game systems to interrogate the simulation world without waiting for a physics step. A ray cast finds the first (or all) bodies intersecting a ray. A shape cast sweeps a collision shape through space and reports the first contact. An overlap test returns all bodies currently touching a given shape. All queries run synchronously on the calling thread during `Update` or `FixedUpdate` — they are read-only operations against the backend's spatial acceleration structure and never mutate simulation state.

## Related Specifications

- [physics-server.md](physics-server.md) — PhysicsServer service exposes the query API
- [collider.md](collider.md) — CollisionGroups used as query filter
- [rigid-body.md](rigid-body.md) — BodyType affects query results
- [math-system.md](math-system.md) — Ray3D, Vec3, Quat primitives
- [event-system.md](event-system.md) — Queries are pull-based; contrast with push-based collision events

## 1. Motivation

Collision events (push-based) cover the case where the simulation tells you something happened. Queries (pull-based) cover the case where you want to ask the simulation a question right now:

- "Is there ground beneath the player's feet?" — character controller floor detection.
- "What does this bullet hit along its path?" — hitscan weapons.
- "Which enemies are within 10 metres of the player?" — AI aggro radius.
- "Can I place a building here without overlapping anything?" — RTS construction.
- "Where would this grenade land if thrown?" — trajectory prediction via shape cast.

Without a query API, game code would have to maintain its own spatial data structures in parallel with the physics backend — duplicated work that drifts out of sync.

## 2. Constraints & Assumptions

- Queries are **read-only**. They never modify body positions, velocities, or contact state.
- Queries execute against the state of the physics world at the **end of the last completed step**. They do not see mid-step intermediate state.
- Queries run on the calling goroutine. They block until results are ready. For expensive batch queries, parallel dispatch is the caller's responsibility via the task system.
- Queries respect `CollisionGroups` filtering — a collider excluded by its filter mask is invisible to queries using that mask.
- `TriMesh` and `HeightField` colliders are queryable. `sensor: true` colliders are optionally included via `QueryFilter.include_sensors`.
- Sleeping bodies are included in query results — sleep affects integration, not spatial presence.

## 3. Core Invariants

- **INV-1**: Queries never mutate simulation state. Calling any query method during a physics step is safe and returns consistent results from the last completed step.
- **INV-2**: A query that hits a sensor collider only returns a result if `QueryFilter.include_sensors = true`. Default is false.
- **INV-3**: All query results return `Entity` identifiers, not `PhysicsBodyHandle`. The mapping from handle to entity is the backend's responsibility via `user_data`.
- **INV-4**: `RayCastAll` and `OverlapAll` results are unordered unless the caller requests sorting. The engine does not guarantee hit order without explicit sort.
- **INV-5**: A query with a `QueryFilter` that excludes all groups returns an empty result immediately — no traversal performed.

## 4. Detailed Design

### 4.1 QueryFilter

Every query accepts a `QueryFilter` that controls which colliders are considered:

```plaintext
QueryFilter
  collision_groups:  CollisionGroups    // membership + filter mask (collider.md §5.5)
                                        // default: all groups
  exclude_entities:  []Entity           // ignore these specific entities
  exclude_rigid_body_types: BodyTypeMask // bitmask: Skip Static | Kinematic | Dynamic
  include_sensors:   bool               // include sensor colliders, default false
  predicate:         Option[func(Entity) bool]
                                        // custom per-entity rejection function
                                        // called only after group filter passes
```

`predicate` is an escape hatch for cases the bitfield cannot express — for example "ignore all entities that have a `TeamComponent` matching my own team". It is called at most once per candidate hit.

`BodyTypeMask` combinations:

```plaintext
BodyTypeMask:
  SkipStatic     = 0b001
  SkipKinematic  = 0b010
  SkipDynamic    = 0b100
  OnlyDynamic    = SkipStatic | SkipKinematic
  OnlyStatic     = SkipKinematic | SkipDynamic
```

### 4.2 Ray Cast

Projects a ray from an origin in a direction and returns the first collider hit:

```plaintext
RayCastHit
  entity:   Entity        // the hit entity
  point:    Vec3          // world-space hit point on the collider surface
  normal:   Vec3          // world-space surface normal at hit point, pointing outward
  distance: float32       // distance from ray origin to hit point
  collider: ColliderIndex // which shape on a compound body was hit (0 for simple bodies)
```

API on `PhysicsServer`:

```plaintext
RayCast(
  ray:         Ray3D,
  max_distance: float32,       // default math.MaxFloat32
  filter:      QueryFilter,
) -> Option[RayCastHit]
// Returns the closest hit along the ray, or None.

RayCastAll(
  ray:          Ray3D,
  max_distance: float32,
  filter:       QueryFilter,
) -> []RayCastHit
// Returns all hits along the ray, unsorted by default.
// Caller sorts by hit.distance if order matters.
```

Common use cases:

```plaintext
// Hitscan weapon — find first solid hit, ignore the shooter:
hit = physics.RayCast(
    Ray3D{ origin: gun_pos, direction: aim_dir },
    max_distance: 500.0,
    filter: QueryFilter{
        exclude_entities: [shooter_entity],
        exclude_rigid_body_types: BodyTypeMask::SkipStatic,  // only dynamic targets
    },
)

// Floor detection for character controller:
hit = physics.RayCast(
    Ray3D{ origin: feet_pos + Vec3{0, 0.1, 0}, direction: Vec3{0, -1, 0} },
    max_distance: 0.3,
    filter: QueryFilter{ exclude_entities: [player_entity] },
)
is_grounded = hit.IsSome()

// Piercing shot — hit all enemies in a line:
hits = physics.RayCastAll(ray, 200.0, filter)
sort hits by distance ascending
for hit in hits: apply_damage(hit.entity)
```

### 4.3 Shape Cast (Sweep)

Sweeps a collision shape through space along a direction vector and returns the first contact:

```plaintext
ShapeCastHit
  entity:       Entity
  point:        Vec3          // deepest contact point on the target collider
  normal:       Vec3          // surface normal at contact, pointing away from target
  distance:     float32       // distance from shape origin to first contact
  time_of_impact: float32     // in [0, 1]: fraction of direction vector at first contact
```

API:

```plaintext
ShapeCast(
  shape:     ColliderShape,   // the sweeping shape (any shape from collider.md §5.2)
  origin:    Vec3,            // world-space starting position of shape centre
  rotation:  Quat,            // world-space starting rotation of shape
  direction: Vec3,            // sweep direction and maximum distance (length = max dist)
  filter:    QueryFilter,
) -> Option[ShapeCastHit]

ShapeCastAll(
  shape:     ColliderShape,
  origin:    Vec3,
  rotation:  Quat,
  direction: Vec3,
  filter:    QueryFilter,
) -> []ShapeCastHit
```

Common use cases:

```plaintext
// Grenade arc preview — sweep sphere along predicted path:
for each segment in arc_segments:
    hit = physics.ShapeCast(
        shape: Sphere{ radius: 0.15 },
        origin: segment.start,
        rotation: Quat::IDENTITY,
        direction: segment.end - segment.start,
        filter: QueryFilter{},
    )
    if hit.IsSome(): draw_impact_marker(hit.point); break

// Character step-up detection — can the player walk up this ledge?
hit = physics.ShapeCast(
    shape: Capsule{ radius: 0.4, half_height: 0.9 },
    origin: player_pos + Vec3{0, step_height, 0},   // raised start
    rotation: Quat::IDENTITY,
    direction: move_direction * move_speed * dt,
    filter: QueryFilter{ exclude_entities: [player_entity] },
)
can_step = hit.IsNone()
```

### 4.4 Overlap Test

Returns all colliders whose shapes currently intersect a given test shape — no movement, purely an intersection query:

```plaintext
OverlapResult
  entity:   Entity
  collider: ColliderIndex
```

API:

```plaintext
OverlapSphere(
  center:  Vec3,
  radius:  float32,
  filter:  QueryFilter,
) -> []OverlapResult
// Convenience wrapper — equivalent to OverlapShape with a Sphere.

OverlapBox(
  center:      Vec3,
  half_extents: Vec3,
  rotation:    Quat,
  filter:      QueryFilter,
) -> []OverlapResult

OverlapShape(
  shape:    ColliderShape,
  position: Vec3,
  rotation: Quat,
  filter:   QueryFilter,
) -> []OverlapResult
// General form — accepts any ColliderShape.

IntersectsShape(
  shape:    ColliderShape,
  position: Vec3,
  rotation: Quat,
  filter:   QueryFilter,
) -> bool
// Fast boolean version — stops at first hit. No allocation.
```

Common use cases:

```plaintext
// AI aggro radius — find all enemies within 12m:
enemies = physics.OverlapSphere(
    center: ai_pos,
    radius: 12.0,
    filter: QueryFilter{
        collision_groups: CollisionGroups{
            membership: GROUP_AI,
            filter: GROUP_PLAYER,
        },
    },
)
for result in enemies: add_to_threat_list(result.entity)

// RTS building placement — check if footprint is clear:
is_clear = !physics.IntersectsShape(
    shape: Box{ half_extents: building_half_size },
    position: cursor_world_pos,
    rotation: Quat::IDENTITY,
    filter: QueryFilter{ exclude_rigid_body_types: BodyTypeMask::SkipStatic },
)
show_placement_indicator(is_clear)

// Explosion radius — apply damage falloff to all nearby bodies:
hits = physics.OverlapSphere(explosion_pos, blast_radius, QueryFilter{})
for hit in hits:
    dist = distance(explosion_pos, get_position(hit.entity))
    falloff = 1.0 - (dist / blast_radius)
    apply_damage(hit.entity, base_damage * falloff)
```

### 4.5 Point Test

Tests whether a world-space point is inside any collider:

```plaintext
PointResult
  entity:   Entity
  collider: ColliderIndex
  distance: float32   // distance from point to nearest surface (0 if inside)

PointQuery(
  point:  Vec3,
  filter: QueryFilter,
) -> Option[PointResult]
// Returns the collider containing the point, or closest if none contain it.

ProjectPoint(
  point:  Vec3,
  filter: QueryFilter,
) -> PointResult
// Projects the point onto the nearest collider surface.
// Always returns a result (distance = 0 if inside).
```

Common use case:

```plaintext
// Is the cursor inside a UI-world object?
result = physics.PointQuery(world_cursor_pos, QueryFilter{ include_sensors: true })
if result.IsSome(): highlight_entity(result.entity)

// Snap object to surface:
proj = physics.ProjectPoint(held_object_pos, terrain_filter)
held_object_pos = proj.point + proj.normal * object_radius
```

### 4.6 Contact Pair Query

Reads the current contact manifold between two specific entities (if they are touching):

```plaintext
ContactPair
  entity_a:    Entity
  entity_b:    Entity
  manifold:    ContactManifold    // same type as in collision events
  is_touching: bool

ContactsBetween(
  entity_a: Entity,
  entity_b: Entity,
) -> Option[ContactPair]
// Returns None if the pair is not in the contact graph.

ContactsWithEntity(
  entity: Entity,
) -> []ContactPair
// Returns all current contact pairs involving this entity.
```

Common use cases:

```plaintext
// Is the player currently touching the ground platform?
pair = physics.ContactsBetween(player, ground_platform)
if pair.IsSome() && pair.is_touching: do_footstep_fx()

// Inspect all contacts on a physics prop for damage:
for pair in physics.ContactsWithEntity(prop):
    max_impulse = max over pair.manifold.contact_points of point.impulse
    if max_impulse > damage_threshold: apply_damage(prop, max_impulse)
```

### 4.7 Batch Queries

For AI systems or large-scale simulations that need many queries per frame, a batch API avoids per-query overhead:

```plaintext
QueryBatch
  ray_casts:    []RayCastRequest
  shape_casts:  []ShapeCastRequest
  overlaps:     []OverlapRequest

QueryBatchResults
  ray_hits:     []Option[RayCastHit]    // parallel to ray_casts
  shape_hits:   []Option[ShapeCastHit]  // parallel to shape_casts
  overlap_hits: [][]OverlapResult       // parallel to overlaps

ExecuteQueryBatch(batch: QueryBatch) -> QueryBatchResults
```

The backend is free to parallelise batch execution across the BVH. Individual queries traverse the acceleration structure independently, making them embarrassingly parallel. The task system (task-system.md §4.8) dispatches batches across worker goroutines.

For very large batches (AI flocks, particle raycasts), the caller should use `ParIter` from the query system and issue individual `physics.RayCast` calls from parallel system iteration — the backend is thread-safe for concurrent reads.

### 4.8 Query Performance Guidelines

| Query type | Approximate cost | Notes |
| :--- | :--- | :--- |
| `RayCast` | ~0.5–2 µs | BVH traversal + shape test |
| `ShapeCast` | ~2–10 µs | More expensive narrow-phase |
| `OverlapSphere` | ~1–5 µs per result | Scales with hit count |
| `OverlapShape` | ~3–15 µs per result | Shape-dependent |
| `IntersectsShape` | ~1–3 µs | Exits at first hit |
| `ContactsBetween` | O(1) | Direct hash lookup in contact graph |
| `ContactsWithEntity` | O(contacts) | Linear in contact count |

Practical limits for `Update`-rate queries (assuming 60 Hz, 1 ms physics budget):

- Up to ~500 individual `RayCast` calls per frame on modern hardware.
- For >1000 queries per frame, use `ExecuteQueryBatch` or parallel iteration.
- `OverlapAll` over a large radius can return thousands of results — always apply tight `CollisionGroups` filtering first.

### 4.9 Debug Visualisation

All query types integrate with the gizmo system (diagnostic-system §4.6) when debug drawing is enabled:

```plaintext
PhysicsDebugSettings (Resource)
  draw_queries:       bool    // visualise all queries issued this frame
  query_ray_color:    Color   // default cyan
  query_shape_color:  Color   // default yellow
  query_hit_color:    Color   // default red
  query_miss_color:   Color   // default grey
```

When `draw_queries = true`, every `RayCast` draws the ray segment (hit in red, miss in grey), every `ShapeCast` draws the swept shape outline, and every `OverlapShape` draws the test shape wire frame with hit colliders highlighted.

## 5. Open Questions

- Should `RayCastAll` and `OverlapAll` accept a result capacity limit to prevent accidentally querying an entire dense scene? For example `RayCastAll(..., max_results: 16)`.
- `QueryFilter.predicate` is a closure — does this prevent the batch query path from being sent to worker goroutines safely? May need to be restricted to non-capturing functions or replaced with a component tag.
- Should there be a `RayCastSorted` convenience that returns `[]RayCastHit` already sorted by distance, or is caller-side sorting always sufficient?
- `ContactsWithEntity` could be expensive for bodies with many contacts (e.g., a large platform with 50 objects resting on it). Should it have a built-in limit or return an iterator?
- Should physics queries be available during the `FixedUpdate` schedule against the in-progress step state, or only against the last completed step? Mid-step queries introduce consistency concerns.

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-27 | Initial draft — ray cast, shape cast, overlap, point test, contact query, batch API, debug visualisation |
