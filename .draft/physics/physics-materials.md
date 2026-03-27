# Physics Materials

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

A physics material defines the surface response properties of a collider — how much it resists sliding (friction) and how much kinetic energy it returns after impact (restitution). Materials are either declared inline on the `Collider` component or referenced as shared `PhysicsMaterial` assets that multiple colliders can reuse. When two colliders meet, the backend combines their material values using configurable blend rules to produce the final contact response for that pair.

## Related Specifications

- [collider.md](collider.md) — Collider carries inline material fields or a Handle to PhysicsMaterial
- [physics-server.md](physics-server.md) — Backend applies combined material values during contact solving
- [asset-system.md](../main/specifications/asset-system.md) — PhysicsMaterial loaded as an asset

## 1. Motivation

Surface feel is one of the most noticeable aspects of physics in games. Ice feels different from rubber. Steel bouncing on concrete sounds and behaves differently from a basketball on wood. Without a material system, every collider carries hardcoded friction and restitution values that are tedious to tune and impossible to reuse across objects of the same surface type. A shared material asset means changing "ice friction" in one place updates every icy surface in the game instantly, including via hot-reload.

## 2. Constraints & Assumptions

- Material values are applied per contact pair, not per body. Two colliders with different materials blend their values each contact.
- `friction` and `restitution` are in the range `[0.0, ∞)` and `[0.0, 1.0]` respectively. Values outside range are clamped at Sync time with a warning.
- A `PhysicsMaterial` asset is immutable once loaded. Runtime mutation goes through a new asset version, triggering hot-reload.
- Inline material fields on `Collider` (friction, restitution, combine rules) take priority over a referenced `PhysicsMaterial` asset. If both are present, the asset is ignored and a debug warning is logged.
- The default material (when no material is set) is `friction: 0.5, restitution: 0.0, combine: Average/Maximum`.

## 3. Core Invariants

- **INV-1**: Every contact pair has exactly one effective friction value and one effective restitution value, computed by blending the two colliders' materials.
- **INV-2**: Blend rule precedence: `Maximum > Multiply > Minimum > Average`. When two colliders use different rules for the same property, the higher-precedence rule wins.
- **INV-3**: `restitution` is always clamped to `[0.0, 1.0]` before being passed to the solver. Values above 1.0 would add energy to the system.
- **INV-4**: Changing a `PhysicsMaterial` asset value (hot-reload) propagates to all colliders referencing it within the same Sync phase.

## 4. Detailed Design

### 4.1 Inline Material Fields on Collider

The simplest approach — material values live directly on the `Collider` component (already listed in collider.md §5.1). No asset required:

```plaintext
Collider
  ...
  friction:            float32       // kinetic friction coefficient, default 0.5
  restitution:         float32       // bounciness [0..1], default 0.0
  friction_combine:    CombineRule   // default Average
  restitution_combine: CombineRule   // default Maximum
```

Use inline fields for one-off surfaces or prototyping. For anything used on more than one object, prefer a shared asset (§4.2).

### 4.2 PhysicsMaterial Asset

A reusable surface definition loaded through the asset system:

```plaintext
PhysicsMaterial
  friction:            float32
  restitution:         float32
  friction_combine:    CombineRule
  restitution_combine: CombineRule
  density:             float32       // kg/m³, overrides Collider.density if set
                                     // used for Auto mass computation
```

Referencing a material from a collider:

```plaintext
Collider
  shape:    Box{ half_extents: Vec3{1,1,1} }
  material: Handle<PhysicsMaterial>   // optional, mutually exclusive with inline fields
```

File format (`.phymat.json`):

```plaintext
{
  "friction": 0.05,
  "restitution": 0.1,
  "friction_combine": "Minimum",
  "restitution_combine": "Maximum",
  "density": 900.0
}
```

### 4.3 CombineRule

Controls how the values of two colliders are blended at contact:

```plaintext
CombineRule:
  Average    — (a + b) / 2
  Minimum    — min(a, b)
  Maximum    — max(a, b)
  Multiply   — a * b

Precedence (highest wins when rules differ):
  Maximum > Multiply > Minimum > Average
```

Worked example — ball on ice:

```plaintext
ball:  friction=0.8, friction_combine=Average
ice:   friction=0.02, friction_combine=Minimum

// ice.Minimum has higher precedence than ball.Average
effective_friction = min(0.8, 0.02) = 0.02
// Result: the ball slides as if both surfaces are slippery — correct behaviour.
```

Worked example — rubber ball bouncing on concrete:

```plaintext
ball:      restitution=0.9, restitution_combine=Maximum
concrete:  restitution=0.1, restitution_combine=Average

// ball.Maximum wins
effective_restitution = max(0.9, 0.1) = 0.9
// Result: high bounce because the rubber ball dominates — correct.
```

### 4.4 Predefined Material Presets

A set of built-in presets shipped with the physics plugin. They are normal `PhysicsMaterial` assets loaded from the engine's embedded asset path (`vfs:///engine/physics/materials/`):

| Preset | Friction | Restitution | Density | Notes |
| :--- | :--- | :--- | :--- | :--- |
| `Default` | 0.50 | 0.00 | 1000 | General purpose |
| `Ice` | 0.02 | 0.05 | 900 | Near-frictionless |
| `Rubber` | 0.90 | 0.80 | 1200 | High grip, bouncy |
| `Metal` | 0.40 | 0.20 | 7800 | Moderate friction, low bounce |
| `Wood` | 0.60 | 0.10 | 600 | Typical floor material |
| `Concrete` | 0.70 | 0.05 | 2000 | High friction, barely bouncy |
| `Glass` | 0.20 | 0.30 | 2500 | Low friction, moderate bounce |
| `Mud` | 0.95 | 0.00 | 1500 | Maximum grip, no bounce |

Presets are accessed via constants:

```plaintext
PhysicsMaterials::ICE      -> Handle<PhysicsMaterial>
PhysicsMaterials::RUBBER   -> Handle<PhysicsMaterial>
PhysicsMaterials::METAL    -> Handle<PhysicsMaterial>
// etc.
```

Presets are not special — they are regular assets that can be overridden per-project by mounting a replacement at the same VFS path.

### 4.5 Per-Pair Material Override

For cases where two specific surfaces should behave differently than their individual materials would imply (e.g., player feet on ice should be stickier than a ball on ice), a `MaterialOverrideTable` resource maps entity pairs to override values:

```plaintext
MaterialOverrideTable (Resource)
  overrides: map[(EntityCategory, EntityCategory)]MaterialOverride

MaterialOverride
  friction:   Option[float32]    // None = use computed blend
  restitution: Option[float32]

EntityCategory (user-defined tag component)
  // e.g.: struct PlayerFeet {}; struct IceSurface {}
```

The Sync phase checks this table after blending. If a matching override exists, the overridden values replace the blended ones. This is intentionally limited to category pairs (not individual entity pairs) to keep the table small.

### 4.6 Sound and VFX Hints

`PhysicsMaterial` carries optional tags that game systems can use to select appropriate audio and visual effects on contact — without hard-coding material logic in the physics system itself:

```plaintext
PhysicsMaterial (additions)
  surface_tag:   StringName   // e.g., "ice", "metal", "wood", "grass"
                              // game code maps this to sound/particle assets
```

`surface_tag` is read from `ContactManifold` data exposed in collision events:

```plaintext
// In collision event handler — pick footstep sound by surface:
hit_entity_material = get_material(event.entity_b)
sound = footstep_sounds[hit_entity_material.surface_tag]
audio.play_one_shot(sound, event.manifold.contact_points[0].position)
```

The engine does not define what tags mean — that is game code's responsibility. The tag is a free-form string, not an enum.

### 4.7 Hot-Reload

Because `PhysicsMaterial` is a standard asset, modifying the `.phymat.json` file during development triggers hot-reload:

1. Asset system detects file change, reloads `PhysicsMaterial`.
2. `AssetEvent<PhysicsMaterial>::Modified` is emitted.
3. Physics Sync system listens for this event and marks all colliders referencing the changed material as dirty.
4. On next Sync phase, dirty colliders update their backend shape descriptors with new material values.

This allows tuning surface feel in real time without restarting the game.

## 5. Open Questions

- Should `density` live on `PhysicsMaterial` (§4.2) or stay only on `Collider`? Having it on both is a source of confusion. Keeping density exclusively on `Collider` keeps materials purely about surface response, not mass.
- `MaterialOverrideTable` uses `EntityCategory` tag components — is this expressive enough, or should overrides be keyed on `Handle<PhysicsMaterial>` pairs instead?
- Should `surface_tag` be a `StringName` (interned) or a `TypeID` pointing to a registered tag component? TypeID approach integrates with the type registry; StringName is simpler.
- Should the engine ship more presets (sand, snow, carpet, water surface) or keep the built-in list minimal and let games define their own?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-27 | Initial draft — inline fields, shared asset, combine rules, presets, per-pair override, surface tags, hot-reload |
