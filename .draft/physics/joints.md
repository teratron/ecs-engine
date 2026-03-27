# Joints

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

Joints constrain the relative motion between two rigid bodies. Each joint type restricts a specific subset of the six degrees of freedom (three translational, three rotational) while allowing the rest to move freely. Joints are declared as ECS components on a dedicated joint entity that references the two bodies being connected. The Physics Server reads joint components during the Sync phase and maintains corresponding constraint objects inside the backend. Breaking forces, motor drives, and soft limits are all configured on the component.

## Related Specifications

- [physics-server.md](physics-server.md) — Server Sync phase that creates and destroys joint constraints
- [rigid-body.md](rigid-body.md) — The two bodies connected by a joint
- [component-system.md](component-system.md) — OnAdd/OnRemove hooks, required components
- [event-system.md](event-system.md) — JointBroken event delivered through the event bus
- [math-system.md](math-system.md) — Vec3, Quat, coordinate frames

## 1. Motivation

Many game objects require constrained motion between two bodies. A door swings on a hinge. A vehicle's wheels rotate around an axle. A rope segment connects to the next. A crane lifts a crate on a cable. Without joints, these behaviours require manual velocity manipulation every frame — fragile, non-deterministic, and divergent under high loads. A constraint solver handles all of this correctly as part of the normal physics step.

## 2. Constraints & Assumptions

- Every joint connects exactly two bodies: `body_a` and `body_b`. One-body joints (anchored to the world) are expressed by setting `body_b = Entity::PLACEHOLDER`; the backend treats this as a connection to a static world anchor.
- Joint anchors are expressed in the local space of each body. The backend transforms them to world space each step.
- Joints are broken (removed) when the constraint force exceeds `break_force` or `break_torque`. A `JointBroken` event is emitted. The joint entity is not despawned automatically — game code decides whether to despawn it.
- Changing `body_a` or `body_b` at runtime is not supported. Despawn the joint entity and spawn a new one.
- A joint between two `Static` bodies is a configuration error — logged and ignored.

## 3. Core Invariants

- **INV-1**: A joint entity must reference two distinct live entities via `body_a` and `body_b`. Self-referential joints are rejected at Sync time.
- **INV-2**: When either referenced body is despawned, the joint is automatically removed from the backend and a `JointBroken` event is emitted with `reason: BodyDespawned`.
- **INV-3**: Joint limits are always expressed such that `lower_limit <= upper_limit`. Inverted limits are clamped at Sync time with a warning.
- **INV-4**: A motor and a position target cannot both be active simultaneously on the same axis. The motor takes precedence; position target is ignored with a warning.
- **INV-5**: Joint constraint forces are solved as part of the normal solver iteration. Joints do not bypass the contact solver.

## 4. Detailed Design

### 4.1 Joint Entity Pattern

Joints live on their own dedicated entities — they are not components on either body:

```plaintext
// A door hinge:
hinge_entity = commands.Spawn(
    Transform {},                         // world-space anchor position
    RevoluteJoint {
        body_a:       door_frame_entity,
        body_b:       door_entity,
        anchor_a:     Vec3{1.0, 1.0, 0},  // hinge point in frame's local space
        anchor_b:     Vec3{-0.5, 1.0, 0}, // hinge point in door's local space
        axis:         Vec3{0, 1, 0},       // rotation axis (Y = vertical hinge)
        lower_limit:  -Deg(90),
        upper_limit:  Deg(0),
    },
)
```

This separation means joints are queryable, independently despawnable, and can carry additional components (e.g., a `Name`, a `BreakEffect`) without polluting either body entity.

### 4.2 Joint Types

#### FixedJoint

Locks all six degrees of freedom. The two bodies move as a single rigid unit.

```plaintext
FixedJoint
  body_a:        Entity
  body_b:        Entity
  anchor_a:      Vec3          // local-space anchor on body_a
  anchor_b:      Vec3          // local-space anchor on body_b
  break_force:   float32       // Newtons, default infinity (never breaks)
  break_torque:  float32       // N·m, default infinity
```

Use for: welding a prop to a vehicle at runtime, constructing compound structures from separate bodies that need to separate later.

#### RevoluteJoint (Hinge)

Allows rotation around one axis. Locks all translation and the other two rotational axes.

```plaintext
RevoluteJoint
  body_a, body_b:  Entity
  anchor_a:        Vec3          // pivot point in body_a local space
  anchor_b:        Vec3          // pivot point in body_b local space
  axis:            Vec3          // rotation axis in body_a local space
  lower_limit:     Option[float32]  // radians, default None (free)
  upper_limit:     Option[float32]  // radians, default None (free)
  stiffness:       float32       // spring stiffness when at limit, default 0 (hard)
  damping:         float32       // damping at limit, default 0
  motor:           Option[JointMotor]
  break_force:     float32
  break_torque:    float32
```

Use for: doors, wheels, levers, pendulums, cranks.

#### PrismaticJoint (Slider)

Allows translation along one axis. Locks all rotation and the other two translational axes.

```plaintext
PrismaticJoint
  body_a, body_b:  Entity
  anchor_a:        Vec3
  anchor_b:        Vec3
  axis:            Vec3          // slide axis in body_a local space
  lower_limit:     Option[float32]  // metres
  upper_limit:     Option[float32]  // metres
  stiffness:       float32
  damping:         float32
  motor:           Option[JointMotor]
  break_force:     float32
  break_torque:    float32
```

Use for: pistons, elevators, sliding doors, recoil mechanisms.

#### SphericalJoint (Ball-and-Socket)

Allows rotation around all three axes. Locks all translation. An optional cone limit restricts the total swing angle.

```plaintext
SphericalJoint
  body_a, body_b:  Entity
  anchor_a:        Vec3
  anchor_b:        Vec3
  cone_limit:      Option[float32]  // half-angle in radians, default None (free)
  twist_limit:     Option[TwistLimit]
  break_force:     float32
  break_torque:    float32

TwistLimit
  lower: float32   // radians
  upper: float32   // radians
```

Use for: ragdoll shoulder/hip joints, rope links, character neck.

#### DistanceJoint (Spring / Rope)

Maintains a target distance between two anchor points. Can be rigid (rope) or soft (spring).

```plaintext
DistanceJoint
  body_a, body_b:  Entity
  anchor_a:        Vec3
  anchor_b:        Vec3
  min_distance:    float32       // lower bound, default 0
  max_distance:    float32       // upper bound; rope behaviour when min == max
  stiffness:       float32       // spring constant k (N/m), 0 = rigid rope
  damping:         float32       // damping coefficient
  break_force:     float32
  break_torque:    float32
```

Use for: elastic bands, ropes, chains (chain = series of DistanceJoints), bungee cords, tethers.

#### GenericJoint (6-DOF)

Full control over all six degrees of freedom. Each axis can be freely configured as locked, limited, or motorised. All other joint types are specialisations of this.

```plaintext
GenericJoint
  body_a, body_b:  Entity
  anchor_a:        Vec3
  anchor_b:        Vec3
  frame_a:         Quat          // local orientation of constraint frame on body_a
  frame_b:         Quat          // local orientation of constraint frame on body_b
  linear_axes:     [3]AxisConfig // X, Y, Z translation axes
  angular_axes:    [3]AxisConfig // X, Y, Z rotation axes
  break_force:     float32
  break_torque:    float32

AxisConfig
  motion:   AxisMotion           // Locked | Limited | Free
  limits:   AxisLimits           // lower, upper bounds (used when Limited)
  stiffness: float32
  damping:   float32
  motor:     Option[JointMotor]
```

Use for: vehicle suspension (specific DOF combinations), complex mechanical linkages, editor-configurable joints.

### 4.3 JointMotor

A motor drives a joint axis toward a target velocity or position:

```plaintext
JointMotor
  mode:             MotorMode       // VelocityDrive | PositionDrive
  target_velocity:  float32         // rad/s or m/s depending on joint type
  target_position:  float32         // radians or metres (PositionDrive only)
  max_force:        float32         // maximum force/torque the motor can apply, default infinity
  stiffness:        float32         // PD controller proportional gain (PositionDrive)
  damping:          float32         // PD controller derivative gain

MotorMode:
  VelocityDrive   — maintains target_velocity; used for wheels, conveyor belts, fans
  PositionDrive   — moves toward target_position using PD controller; used for robotic arms
```

Motor example — powered door:

```plaintext
// Open door motor:
hinge.motor = Some(JointMotor {
    mode:            MotorMode::PositionDrive,
    target_position: Deg(-90),    // open angle
    stiffness:       500.0,
    damping:         50.0,
    max_force:       200.0,       // limit so player can hold it shut
})

// Close door:
hinge.motor.target_position = Deg(0)
```

### 4.4 Soft Limits

When `stiffness > 0` on a limit, the joint becomes a spring rather than a hard stop. The body can exceed the limit; a restoring force proportional to `stiffness * penetration_depth` plus `damping * velocity` pulls it back.

This avoids jitter at hard limits for heavy bodies and produces more natural-looking motion for ragdolls and suspension:

```plaintext
// Soft suspension spring:
PrismaticJoint {
    lower_limit: -0.3,    // metres
    upper_limit:  0.0,
    stiffness:    8000.0, // N/m — typical car spring rate
    damping:       400.0, // N·s/m — shock absorber
}
```

### 4.5 Break Forces and JointBroken Event

When the constraint force or torque during a solver iteration exceeds the threshold, the joint is broken:

```plaintext
JointBroken
  joint_entity: Entity
  body_a:       Entity
  body_b:       Entity
  break_force:  float32     // actual force at break
  reason:       BreakReason

BreakReason:
  ForceExceeded     — constraint force exceeded break_force
  TorqueExceeded    — constraint torque exceeded break_torque
  BodyDespawned     — one of the referenced bodies was removed
```

The joint entity is **not** automatically despawned. Game code listens for `JointBroken` and decides:

```plaintext
// Snap joint on break, spawn debris:
for event in joint_broken_events.Read():
    commands.Despawn(event.joint_entity)
    spawn_break_particles(event.body_a, event.body_b)
    play_sound("crack.wav")
```

### 4.6 World Anchor

To anchor a body to a fixed point in the world, set `body_b = Entity::PLACEHOLDER`:

```plaintext
// Hang a lantern from the ceiling (world anchor):
commands.Spawn(
    DistanceJoint {
        body_a:       lantern_entity,
        body_b:       Entity::PLACEHOLDER,  // world anchor
        anchor_a:     Vec3{0, 0.2, 0},      // top of lantern
        anchor_b:     Vec3{3, 4, 2},        // fixed ceiling point in world space
        min_distance: 1.5,
        max_distance: 1.5,                  // rigid rope length
        stiffness:    0.0,
    },
)
```

The backend treats `PLACEHOLDER` as a static body with infinite mass at the world origin. `anchor_b` is interpreted in world space when `body_b = PLACEHOLDER`.

### 4.7 Lifecycle and Change Detection

`OnAdd` hook on a joint component: enqueue `CreateJoint(descriptor)` in `PhysicsCommandQueue`.
`OnRemove` hook: enqueue `DestroyJoint(handle)`.

Field changes picked up in Sync phase:

| Field changed | Command issued |
| :--- | :--- |
| `lower_limit` / `upper_limit` | UpdateLimits |
| `stiffness` / `damping` | UpdateCompliance |
| `motor` | UpdateMotor |
| `break_force` / `break_torque` | UpdateBreakThresholds |
| `body_a` or `body_b` | Rejected — log error, no-op |

### 4.8 Ragdoll Helper

A ragdoll is a hierarchy of bodies connected by SphericalJoints and RevoluteJoints mirroring a skeleton. A `RagdollBuilder` utility (not a component) constructs the entity graph from a skeleton description:

```plaintext
RagdollBuilder
  fn from_skeleton(skeleton: SkeletonDesc) -> RagdollDesc
  fn build(world: &mut Commands, root_transform: Transform) -> RagdollInstance

RagdollInstance
  root_entity:  Entity
  body_map:     map[BoneName]Entity
  joint_map:    map[BoneName]Entity
```

Once built, the ragdoll is a normal collection of `RigidBody` + `Collider` + joint entities. No special runtime system manages it — physics handles everything.

## 5. Open Questions

- Should `RevoluteJoint` support a second motor for the perpendicular axis (for universal joints / CV joints in vehicles)?
- `GenericJoint` is powerful but complex. Should it be exposed in v1 or deferred until a concrete use case demands it?
- Constraint stabilisation: should joints use Baumgarte stabilisation (fast, adds energy) or pseudo-velocity correction (slower, more stable)? This is a backend implementation detail but affects the feel of all joint types.
- Should there be a `WeldJoint` alias for `FixedJoint` for discoverability, or is `FixedJoint` name clear enough?
- Motor `max_force: infinity` by default means an uncapped motor can generate unrealistic forces. Should the default be a finite value (e.g., 1000 N) to fail safe?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-27 | Initial draft — joint types, motors, soft limits, break forces, world anchor, ragdoll helper |
