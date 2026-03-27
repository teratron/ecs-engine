# Collision Events

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

Collision events are the push-based complement to physics queries. Where queries ask "what is touching X right now?", collision events tell you "something started touching X" or "something stopped touching X" without polling. Events are emitted by the Physics Server at the end of each fixed step, after the contact solver has run, and delivered through the standard event bus. Game systems subscribe to them in `FixedUpdate` or `Update` and react accordingly. A second family — trigger events — covers sensor colliders that detect overlap without generating contact forces.

## Related Specifications

- [physics-server.md](physics-server.md) — Events phase that diffs contact sets and emits events
- [collider.md](collider.md) — Collider sensor flag, CollisionGroups, contact_force_threshold
- [rigid-body.md](rigid-body.md) — BodyType affects which event types are generated
- [event-system.md](event-system.md) — EventWriter/EventReader infrastructure used for delivery
- [physics-query.md](physics-query.md) — Pull-based alternative; ContactsBetween for frame-synchronous reads

## 1. Motivation

Polling `ContactsWithEntity` every frame to detect when a player lands on the ground, when a bullet hits an enemy, or when a pickup zone is entered works but wastes cycles. Most frames nothing changes. Collision events invert this: the engine pays the diffing cost once per step and notifies only the systems that registered interest. Game code becomes reactive rather than polling, and the physics budget is spent where state actually changes.

## 2. Constraints & Assumptions

- Events are emitted once per physics step, not once per render frame. At 60 Hz fixed step and 120 Hz render, each event fires once every two render frames.
- `CollisionStarted` and `CollisionEnded` are guaranteed to be paired — every start has a corresponding end, even if the body is despawned mid-simulation (despawn forces an end event).
- Events are delivered via the standard double-buffered `EventBus`. They persist for two frames (event-system §4.1). Systems in `FixedUpdate` and `Update` both see them.
- Two sensor colliders produce trigger events but never collision events, even if their shapes overlap.
- Collision events are not emitted between two `Static` bodies — static-static pairs never enter the contact graph.
- `ContactManifold` data inside events is a snapshot. It is valid for the frame it is read; do not store it across frames.

## 3. Core Invariants

- **INV-1**: `CollisionStarted` fires exactly once when a contact pair enters the contact graph. It does not repeat while the pair persists.
- **INV-2**: `CollisionEnded` fires exactly once when a contact pair leaves the contact graph or either body is despawned/disabled.
- **INV-3**: `CollisionPersisting` fires every step while a contact pair remains active. It always fires between the corresponding `CollisionStarted` and `CollisionEnded`.
- **INV-4**: Events are always emitted with both `entity_a` and `entity_b` populated with valid `Entity` values. If a body was despawned this step, the entity ID is still valid until ECS processes the despawn command (which happens in the next `ApplyDeferred`).
- **INV-5**: `ContactForceEvent` fires only when `contact_force_threshold > 0` on the receiving collider AND the resolved impulse this step exceeds it. It is independent of `CollisionStarted`/`CollisionPersisting`.

## 4. Detailed Design

### 4.1 Event Types Overview

```plaintext
Contact events (solid colliders):
  CollisionStarted    — first frame two solid colliders touch
  CollisionPersisting — every subsequent frame they remain in contact
  CollisionEnded      — first frame they separate

Trigger events (sensor colliders):
  TriggerEntered      — first frame a non-sensor body overlaps a sensor
  TriggerExited       — first frame they stop overlapping

Force events (threshold-gated):
  ContactForceEvent   — impulse on a collider exceeds configured threshold
```

### 4.2 Contact Events

```plaintext
CollisionStarted
  entity_a:    Entity
  entity_b:    Entity
  manifold:    ContactManifold

CollisionPersisting
  entity_a:    Entity
  entity_b:    Entity
  manifold:    ContactManifold

CollisionEnded
  entity_a:    Entity
  entity_b:    Entity
  // No manifold — contact geometry no longer exists at separation.
```

The ordering of `entity_a` and `entity_b` within a single event is deterministic (sorted by internal body handle) and consistent across `Started` → `Persisting` → `Ended` for the same pair. Systems that filter by entity must check both fields:

```plaintext
// Correct — order not assumed:
for event in started.Read():
    if event.entity_a == player || event.entity_b == player:
        on_player_contact(event)
```

### 4.3 ContactManifold

The contact geometry snapshot for one pair of touching colliders:

```plaintext
ContactManifold
  contact_points:  []ContactPoint    // 1–4 points per manifold (backend-dependent)
  normal:          Vec3              // world-space collision normal, points from B toward A
  total_impulse:   float32           // sum of all contact point impulses this step (N·s)
  tangent_impulse: Vec2              // friction impulse components (tangential plane)

ContactPoint
  position:        Vec3     // world-space contact point (midpoint between surfaces)
  depth:           float32  // penetration depth in metres (>0 means overlapping)
  impulse:         float32  // normal impulse resolved at this point this step (N·s)
```

Reading manifold data:

```plaintext
// Landing detection — check normal is upward-ish and impulse significant:
for event in started.Read():
    if event.entity_a != player && event.entity_b != player: continue
    up_dot = dot(event.manifold.normal, Vec3{0, 1, 0})
    if up_dot > 0.7 && event.manifold.total_impulse > 5.0:
        player_state.just_landed = true
        play_land_sound(event.manifold.total_impulse)
```

### 4.4 Trigger Events

Sensor colliders (`Collider { sensor: true }`) generate trigger events instead of contact events:

```plaintext
TriggerEntered
  entity_sensor: Entity   // the sensor collider's entity
  entity_other:  Entity   // the non-sensor body that entered

TriggerExited
  entity_sensor: Entity
  entity_other:  Entity
```

Triggers do not generate `ContactManifold` data — sensors have no contact response and the backend does not compute a manifold for them.

Trigger events are emitted when:
- A non-sensor body enters the sensor's shape → `TriggerEntered`.
- A non-sensor body exits the sensor's shape → `TriggerExited`.
- The non-sensor body is despawned while inside → `TriggerExited` with valid (still-alive this frame) entity ID.

Two sensor bodies overlapping each other produce no events.

```plaintext
// Pickup zone:
for event in trigger_entered.Read():
    if event.entity_sensor != pickup_zone: continue
    if world.has_component::<Player>(event.entity_other):
        collect_pickup(event.entity_other, pickup_zone)
        commands.Despawn(pickup_zone)
```

### 4.5 ContactForceEvent

Emitted when the total resolved impulse on a collider exceeds its configured threshold in a single step:

```plaintext
ContactForceEvent
  entity_self:  Entity      // the entity with contact_force_threshold set
  entity_other: Entity      // the other body in the contact
  total_force:  Vec3        // world-space resultant force vector this step
  max_impulse:  float32     // peak impulse across all contact points
```

The threshold is set per-collider in `Collider.contact_force_threshold` (collider.md §5.1). Default is `0.0` which disables the event entirely — no overhead for colliders that do not need it.

```plaintext
// Glass shattering on hard impact:
Collider {
    shape: Box{ half_extents: ... },
    contact_force_threshold: 500.0,  // Newtons
}

for event in force_events.Read():
    if event.entity_self != glass_entity: continue
    shatter_glass(glass_entity)
    commands.Despawn(glass_entity)
```

### 4.6 Event Filtering Patterns

Because every system that reads `EventReader<CollisionStarted>` sees all collision events in the world, efficient filtering is important in dense scenes.

**Pattern 1 — Entity tag component check:**

```plaintext
for event in started.Read():
    if !world.has_component::<Damageable>(event.entity_b): continue
    apply_damage(event.entity_b, event.manifold.total_impulse)
```

**Pattern 2 — Observer on specific entity (preferred for targeted reactions):**

```plaintext
// Fires only when this specific entity is involved in a collision:
world.Entity(boss_entity).Observe(func(trigger: On[CollisionStarted]):
    if trigger.entity != boss_entity: return
    play_boss_react_animation()
)
```

**Pattern 3 — Collision layer pre-filtering:**

Set `collision_groups` on colliders so pairs that should never generate events never enter the contact graph at all. This is the most efficient approach — events that do not happen cost nothing.

**Pattern 4 — ContactForceEvent for damage:**

Instead of reading `CollisionPersisting` and computing impulse manually, configure `contact_force_threshold` on destructible objects. Only hard impacts trigger the event — gentle resting contact is ignored.

### 4.7 Contact Graph State Machine

The Physics Server maintains a contact graph across steps. The state machine per pair:

```plaintext
stateDiagram-v2
  [*] --> NoContact
  NoContact --> Touching : shapes overlap (narrow-phase positive)
  Touching --> NoContact : shapes separate OR either body despawned
  Touching --> Touching  : still overlapping next step

  NoContact --> Touching : emits CollisionStarted
  Touching --> Touching  : emits CollisionPersisting
  Touching --> NoContact : emits CollisionEnded
```

The diffing algorithm runs after the Step phase:

```plaintext
FUNCTION diff_contacts(prev_pairs, curr_pairs):
  started  = curr_pairs - prev_pairs
  ended    = prev_pairs - curr_pairs
  persisting = curr_pairs ∩ prev_pairs

  FOR pair IN started:    emit CollisionStarted(pair)
  FOR pair IN persisting: emit CollisionPersisting(pair)
  FOR pair IN ended:      emit CollisionEnded(pair)

  prev_pairs = curr_pairs
```

Contact pairs are keyed by `(min(handleA, handleB), max(handleA, handleB))` — order-independent, consistent across steps.

### 4.8 Sleeping Bodies

A sleeping body does not participate in integration, but its contact pairs remain in the graph. `CollisionPersisting` is **not** emitted for sleeping contact pairs — the pair is considered dormant.

When a sleeping body wakes (due to a nearby impulse or force), its dormant pairs re-enter the active set and `CollisionPersisting` resumes. No new `CollisionStarted` is emitted for pairs that were active before sleep.

This means `CollisionStarted` + `CollisionEnded` count is always balanced even across sleep/wake cycles.

### 4.9 Event Volume and Performance

In dense scenes, `CollisionPersisting` can generate thousands of events per step. Mitigation strategies:

**Disable `CollisionPersisting` selectively:**

```plaintext
ColliderEventFlags (on Collider entity)
  emit_started:    bool   // default true
  emit_persisting: bool   // default false  ← off by default to reduce volume
  emit_ended:      bool   // default true
```

`CollisionPersisting` is disabled by default. Enable it only on entities that need per-frame contact data (e.g., grinding surfaces, conveyor belts, continuous damage zones).

**Use `ContactForceEvent` instead of polling `CollisionPersisting`:**

For impact damage, configure `contact_force_threshold` and react to the force event. No per-frame overhead.

**Use `ContactsBetween` query for specific pairs:**

When only one specific pair matters (is the player touching the ice surface?), a direct query is cheaper than reading all `CollisionPersisting` events and filtering.

### 4.10 Deferred Despawn Safety

A common pattern is despawning an entity in response to a collision event:

```plaintext
for event in started.Read():
    if is_bullet(event.entity_a) && is_enemy(event.entity_b):
        commands.Despawn(event.entity_a)    // despawn bullet
        apply_damage(event.entity_b)
```

This is safe. `Commands.Despawn` is deferred — the entity remains alive for the rest of this frame's systems. The physics backend will emit `CollisionEnded` for this pair in the next step (the entity is still alive in the physics world until the next Sync phase processes the despawn command).

Do not store `event.entity_a` beyond the current frame — it may be invalid after `ApplyDeferred`.

### 4.11 EventReader Registration

All collision event types must be registered with the App:

```plaintext
// Done automatically by PhysicsPlugin:
app.AddEvent[CollisionStarted]()
app.AddEvent[CollisionPersisting]()
app.AddEvent[CollisionEnded]()
app.AddEvent[TriggerEntered]()
app.AddEvent[TriggerExited]()
app.AddEvent[ContactForceEvent]()
app.AddEvent[JointBroken]()   // from joints.md
```

Systems read events via standard `EventReader`:

```plaintext
fn on_collision(
    started:       EventReader[CollisionStarted],
    trigger_enter: EventReader[TriggerEntered],
    force_events:  EventReader[ContactForceEvent],
):
    for event in started.Read(): ...
    for event in trigger_enter.Read(): ...
    for event in force_events.Read(): ...
```

## 5. Open Questions

- Should `CollisionPersisting` be opt-in per collider (as `ColliderEventFlags` in §4.9) or opt-out globally via `PhysicsSettings`? Per-collider is more precise; global toggle is simpler to configure.
- `ContactManifold` in `CollisionStarted` / `CollisionPersisting` — should `ContactPoint.impulse` be zero on the first frame of contact (impulse not yet resolved) or reflect the resolved value from this step?
- Should there be a `CollisionLayer` event filter at the event bus level (emit only events where one entity matches a given group) to reduce reader-side overhead?
- `TriggerEntered` / `TriggerExited` for sensor-sensor pairs: currently no event. Should sensor-sensor overlaps optionally emit their own event type, or is the use case too niche?
- How should events behave during fast-forward (when `FixedUpdate` runs 10+ times in one render frame due to a large delta)? Each step emits independently — could produce bursts of `CollisionStarted` + `CollisionEnded` in the same render frame if an object bounces rapidly.

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-27 | Initial draft — contact/trigger/force events, contact graph state machine, sleep handling, deferred despawn safety, performance guidelines |
