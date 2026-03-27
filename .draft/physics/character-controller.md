# Character Controller

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

The character controller provides kinematic movement for player characters and NPCs that need precise, game-feel-driven motion rather than fully simulated rigid body dynamics. It wraps a capsule collider in a high-level API: move by velocity, detect floor and ceiling, handle step-up, slide along walls, and snap to slopes. The controller is not a rigid body — it does not receive forces, does not have mass, and does not participate in joint constraints. It moves where you tell it to move, stops where geometry blocks it, and reports what it touched.

## Related Specifications

- [physics-server.md](physics-server.md) — PhysicsServer.MoveCharacter executes the move in the backend
- [rigid-body.md](rigid-body.md) — Contrast: RigidBody::Kinematic for objects driven by animation
- [collider.md](collider.md) — Capsule shape used internally; CollisionGroups for filtering
- [physics-query.md](physics-query.md) — GroundProbe uses ShapeCast internally
- [collision-events.md](collision-events.md) — CharacterCollision events delivered per move
- [input-system.md](../main/specifications/input-system.md) — Input drives the velocity passed to the controller
- [time-system.md](../main/specifications/time-system.md) — Movement runs in FixedUpdate using FixedTime.Delta

## 1. Motivation

A fully simulated dynamic body makes a poor player character. Dynamic bodies can be pushed by tiny obstacles, tumble when hit, jitter on stairs, and slide off ledges. Game characters need different rules: instant velocity changes, predictable floor snapping, step-up over kerbs, and wall sliding that feels responsive. Implementing this by hand every project leads to subtle bugs and inconsistent feel. A first-class character controller provides these behaviours as a tested, configurable building block.

## 2. Constraints & Assumptions

- The character controller always uses a **capsule** shape. Non-capsule controllers are not supported — the capsule is the only shape that makes step-up and slope detection reliable.
- Movement runs in `FixedUpdate` to match the physics step rate. Rendering interpolates via `PhysicsTransform` (physics-server §4.5).
- The controller does not receive external impulses or forces. To simulate being pushed, game code writes velocity directly to `CharacterController.velocity`.
- Multiple simultaneous `CharacterController` components on one entity are not permitted.
- The controller interacts with `RigidBody` entities normally — it pushes dynamic bodies based on `push_strength`. It is not pushed back by them (kinematic rule).
- `CharacterController` requires `Transform` and `PhysicsTransform` as required components (same as `RigidBody`).

## 3. Core Invariants

- **INV-1**: `CharacterController.velocity` is the authoritative input each step. The system does not integrate acceleration — that is the caller's responsibility.
- **INV-2**: The controller never penetrates solid geometry. If a move would result in penetration deeper than `skin_width`, the move is clipped.
- **INV-3**: `is_grounded` reflects the state **after** the current step's movement, not before. It is always consistent with the final position.
- **INV-4**: `step_height` applies only when moving horizontally toward an obstacle. Vertical movement (jumping, falling) is never step-assisted.
- **INV-5**: `CharacterCollision` events are emitted for every surface touched during a move, in the order they were encountered.

## 4. Detailed Design

### 4.1 CharacterController Component

```plaintext
CharacterController
  // Shape
  radius:           float32         // capsule radius, default 0.4 m
  height:           float32         // total height (including hemispheres), default 1.8 m

  // Movement
  velocity:         Vec3            // desired velocity this step (m/s), set by game code
  up:               Vec3            // world-space up direction, default (0, 1, 0)
                                    // change for gravity games, planet surfaces, etc.

  // Floor detection
  max_slope_angle:  float32         // steepest walkable slope in radians, default 46°
  step_height:      float32         // max obstacle height the controller steps over, default 0.35 m
  snap_to_floor:    float32         // max distance to snap down to ground, default 0.5 m
                                    // 0 = no snapping (disable for platformers with ledges)

  // Collision response
  skin_width:       float32         // margin between capsule and geometry, default 0.01 m
  push_strength:    float32         // force applied to dynamic bodies on contact, default 50 N
  slide_on_walls:   bool            // project velocity along wall normal, default true

  // Filters
  collision_groups: CollisionGroups // which layers the controller collides with

  // Output (read-only, written by controller system)
  is_grounded:      bool
  ground_normal:    Vec3            // surface normal of the floor, Vec3::UP if not grounded
  ground_entity:    Option[Entity]  // entity the controller is standing on
  velocity_actual:  Vec3            // actual velocity after collision resolution
```

### 4.2 Movement Algorithm

The controller move is a **depenetration-based iterative sweep**. Each call to `MoveCharacter` in the backend:

```plaintext
FUNCTION MoveCharacter(capsule, desired_velocity, dt, settings):
  remaining = desired_velocity * dt    // total displacement to attempt
  position  = capsule.current_position
  MAX_ITERATIONS = 4

  FOR i IN 0..MAX_ITERATIONS:
    IF length(remaining) < skin_width: BREAK

    hit = ShapeCast(capsule_shape, position, normalize(remaining), length(remaining))

    IF hit IS NONE:
      position += remaining
      BREAK

    // Move to just before the hit (leave skin_width gap)
    safe_distance = max(0, hit.distance - skin_width)
    position += normalize(remaining) * safe_distance

    // Emit CharacterCollision event
    emit CharacterCollision{ entity: hit.entity, normal: hit.normal, point: hit.point }

    // Step-up check
    IF is_horizontal_hit(hit.normal, settings.up):
      step_pos = try_step_up(position, remaining, hit, settings)
      IF step_pos IS SOME:
        position = step_pos
        remaining = project_onto_plane(remaining, settings.up)  // continue horizontally
        CONTINUE

    // Slope check
    slope_angle = angle_between(hit.normal, settings.up)
    IF slope_angle <= settings.max_slope_angle:
      // Walkable slope — project remaining onto slope plane
      remaining = project_onto_plane(remaining - normalize(remaining)*safe_distance, hit.normal)
    ELSE IF settings.slide_on_walls:
      // Steep wall — slide along it
      remaining = project_onto_plane(remaining - normalize(remaining)*safe_distance, hit.normal)
    ELSE:
      // Stop on contact
      remaining = Vec3::ZERO
      BREAK

  // Floor snap
  IF settings.snap_to_floor > 0:
    snap_hit = ShapeCast(capsule_shape, position, -settings.up, settings.snap_to_floor)
    IF snap_hit IS SOME AND is_walkable(snap_hit.normal, settings.max_slope_angle):
      position -= settings.up * snap_hit.distance

  // Ground detection probe
  ground_hit = ShapeCast(capsule_shape, position, -settings.up, skin_width * 2)
  is_grounded = ground_hit IS SOME AND is_walkable(ground_hit.normal, settings.max_slope_angle)

  RETURN MoveResult{ final_position: position, is_grounded, ground_normal, ground_entity }
```

### 4.3 Step-Up

Step-up allows the controller to climb low obstacles (kerbs, stairs) without jumping:

```plaintext
FUNCTION try_step_up(position, remaining, hit, settings):
  // 1. Cast up by step_height from current position
  up_hit = ShapeCast(capsule, position, settings.up, settings.step_height)
  clearance = up_hit ? up_hit.distance : settings.step_height

  // 2. Move up by clearance
  raised_pos = position + settings.up * clearance

  // 3. Try horizontal move from raised position
  horiz_hit = ShapeCast(capsule, raised_pos, normalize(remaining), length(remaining))
  IF horiz_hit IS SOME AND horiz_hit.distance < skin_width:
    RETURN None  // still blocked even when raised — not a step

  horiz_pos = raised_pos + normalize(remaining) * (horiz_hit ? horiz_hit.distance - skin_width : length(remaining))

  // 4. Cast down to find the step surface
  down_hit = ShapeCast(capsule, horiz_pos, -settings.up, clearance + skin_width)
  IF down_hit IS NONE:
    RETURN None  // floating after step — reject

  final_pos = horiz_pos - settings.up * (down_hit.distance - skin_width)
  RETURN Some(final_pos)
```

Step-up fires `CharacterCollision` for the step surface with a `collision_type: StepUp` tag.

### 4.4 CharacterCollision Event

Emitted for every surface the controller touches during a single move:

```plaintext
CharacterCollision
  controller_entity: Entity
  hit_entity:        Entity
  normal:            Vec3         // surface normal at contact point
  point:             Vec3         // world-space contact point
  velocity_before:   Vec3         // controller velocity before this collision
  collision_type:    CollisionType

CollisionType:
  Wall      — steep surface, velocity clipped or slid
  Floor     — walkable surface
  Ceiling   — overhead hit, vertical velocity zeroed
  StepUp    — step-up surface was climbed
```

Multiple `CharacterCollision` events can fire per move (one per iteration hit). They are ordered by encounter sequence.

### 4.5 Gravity and Vertical Movement

The controller does not apply gravity automatically. Game code owns vertical velocity:

```plaintext
// Typical character system:
fn move_character(
    ctrl:   Query[&mut CharacterController],
    input:  Res[ButtonInput[KeyCode]],
    time:   Res[FixedTime],
):
    for c in ctrl.Iter():
        // Horizontal input
        move_dir = input_to_direction(input)
        c.velocity.x = move_dir.x * SPEED
        c.velocity.z = move_dir.z * SPEED

        // Gravity
        IF c.is_grounded:
            c.velocity.y = 0
            IF input.JustPressed(KeySpace):
                c.velocity.y = JUMP_SPEED
        ELSE:
            c.velocity.y -= GRAVITY * time.DeltaSeconds()

        // Clamp fall speed
        c.velocity.y = max(c.velocity.y, -MAX_FALL_SPEED)
```

This pattern gives game code full control over gravity strength, jump curves, coyote time, and other feel parameters without the controller system needing to know about them.

### 4.6 Pushing Dynamic Bodies

When the controller contacts a `Dynamic` rigid body, it applies a force to push it aside:

```plaintext
push_force = -collision.normal * settings.push_strength
physics.ApplyImpulse(collision.hit_entity, push_force, collision.point)
```

`push_strength: 0` disables pushing entirely. A high value makes the character feel heavy; a low value lets light characters be deflected by their own push reaction (physically wrong but sometimes desirable).

### 4.7 Slope Behaviour

```plaintext
max_slope_angle = 46° (default)

slope_angle <= max_slope_angle:
  → walkable floor, is_grounded = true
  → velocity projected onto slope plane (character walks up/down naturally)

slope_angle > max_slope_angle AND slide_on_walls = true:
  → steep wall, is_grounded = false on that contact
  → velocity projected along wall (character slides sideways, not into wall)

slope_angle > max_slope_angle AND slide_on_walls = false:
  → full stop on contact
```

Slopes steeper than `max_slope_angle` are treated as walls — the character cannot walk up them. This is intentional: a 90° wall and a 75° steep rock face should both stop the character.

### 4.8 CharacterControllerPlugin

```plaintext
CharacterControllerPlugin

Build(app):
  app.RegisterComponent[CharacterController]()
  app.AddEvent[CharacterCollision]()
  app.AddSystems(FixedUpdate, move_characters)
  app.AddSystems(PostUpdate, interpolate_character_transforms)
  // move_characters runs AFTER physics step, BEFORE WriteBack
  // so the controller's final position feeds into PhysicsTransform this frame
```

### 4.9 Multiple Controllers and NPC Movement

`CharacterController` is not limited to the player. Each entity with the component moves independently each fixed step. NPCs use the same component with their own velocity driven by AI systems:

```plaintext
// NPC patrol system:
fn npc_patrol(
    npcs: Query[(&mut CharacterController, &PatrolState)],
    time: Res[FixedTime],
):
    for (ctrl, patrol) in npcs.Iter():
        dir = direction_to(patrol.current_waypoint, get_position(ctrl))
        ctrl.velocity.x = dir.x * patrol.speed
        ctrl.velocity.z = dir.z * patrol.speed
        // vertical handled by gravity system above
```

Performance: each controller executes up to 4 shape-cast iterations per step. For 100 NPCs at 60 Hz fixed step this is 400 shape casts per second — well within budget (physics-query §4.8 pegs individual shape casts at 2–10 µs).

### 4.10 Relationship to RigidBody::Kinematic

`CharacterController` and `RigidBody { body_type: Kinematic }` are complementary, not competing:

| | CharacterController | RigidBody::Kinematic |
| :--- | :--- | :--- |
| **Use for** | Characters, NPCs | Animated doors, platforms, vehicles |
| **Driven by** | Velocity each FixedUpdate | Transform / animation |
| **Floor detection** | Built-in | Not provided |
| **Step-up** | Built-in | Not provided |
| **Slope handling** | Built-in | Not provided |
| **Pushes dynamics** | Yes, configurable | Yes, implicit from solver |
| **Receives forces** | No | No |
| **Joints** | Not supported | Supported |

Do not use `CharacterController` for doors, elevators, or vehicles. Do not use `RigidBody::Kinematic` for player characters unless you are implementing your own movement solver on top.

## 5. Open Questions

- Should the controller support non-capsule shapes in a future version? A box controller is useful for top-down games where capsule edge-rounding is undesirable.
- `snap_to_floor` default of 0.5 m may be too aggressive for platformers where ledge-dropping is intentional. Should the default be `0` (off) and opt-in?
- How should the controller interact with moving platforms (a `RigidBody::Kinematic` elevator)? The controller needs to inherit the platform's velocity when standing on it. This requires reading `ground_entity`'s velocity from `PhysicsServer.GetLinearVelocity` each step.
- Should `CharacterCollision` events include the `PhysicsMaterial` of the hit surface for audio/VFX dispatch, or is that the responsibility of a separate lookup?
- Multiple capsule shapes per controller (e.g., crouching vs. standing) — swap `radius`/`height` on the component, or maintain two colliders and toggle between them?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-27 | Initial draft — capsule controller, depenetration sweep, step-up, slope handling, gravity ownership, push, NPC usage |
