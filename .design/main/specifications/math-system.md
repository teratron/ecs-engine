# Math System

**Version:** 0.2.0
**Status:** Draft
**Layer:** concept

## Overview

The math system is a pure math library with no ECS dependencies. It provides vector, matrix, quaternion, transform, color, and geometric primitive types used throughout the engine. All operations follow value semantics — methods return new values rather than mutating the receiver.

## Related Specifications

- [hierarchy-system.md](hierarchy-system.md) — Transform components use Vec3, Quat, and Affine3A

## 1. Motivation

Every engine subsystem depends on math primitives. A dedicated math library must:
- Provide correct, well-tested implementations of common 3D math operations.
- Use value semantics to avoid aliasing bugs in concurrent systems.
- Offer both convenience (euler angles) and performance (quaternions, affine transforms).
- Remain decoupled from ECS so it can be used in standalone tools and tests.

## 2. Constraints & Assumptions

- All floating-point types use float32 unless explicitly suffixed (e.g., Vec3D for float64).
- Column-major matrix layout matches GPU conventions.
- Quaternions are always normalized after construction; operations that could break normalization re-normalize.
- Direction types enforce unit-length at construction time via fallible constructors.
- Color conversions are non-lossy within the same bit depth.

## 3. Core Invariants

1. **Immutability.** No method mutates its receiver; all return new values.
2. **Quaternion normalization.** Quat values are always unit quaternions.
3. **Direction normalization.** Dir2 and Dir3 always have length 1.0 (within float epsilon).
4. **Column-major layout.** Mat3 and Mat4 store columns contiguously for GPU compatibility.

## 4. Detailed Design

### 4.1 Vector Types

```plaintext
Vec2 { x: f32, y: f32 }
Vec3 { x: f32, y: f32, z: f32 }
Vec4 { x: f32, y: f32, z: f32, w: f32 }

Common operations (all return new values):
  add, sub, mul, div           -- component-wise arithmetic
  dot(other) -> f32            -- dot product
  cross(other) -> Vec3         -- cross product (Vec3 only)
  length() -> f32              -- magnitude
  normalize() -> Self          -- unit vector
  lerp(other, t) -> Self       -- linear interpolation
  distance_to(other) -> f32
  min(other), max(other)       -- component-wise min/max
  clamp(min, max) -> Self
```

### 4.2 Matrix Types

```plaintext
Mat3 { cols: [Vec3; 3] }      -- 3x3, used for 2D transforms and normals
Mat4 { cols: [Vec4; 4] }      -- 4x4, used for projection and view matrices

Operations:
  mul_mat(other) -> Self       -- matrix multiplication
  mul_vec(v) -> Vec            -- transform a vector
  transpose() -> Self
  inverse() -> Option[Self]    -- None if singular
  determinant() -> f32
  from_scale(Vec3) -> Mat4
  from_translation(Vec3) -> Mat4
  perspective(fov, aspect, near, far) -> Mat4
  orthographic(left, right, bottom, top, near, far) -> Mat4
```

### 4.3 Quaternion

```plaintext
Quat { x: f32, y: f32, z: f32, w: f32 }   -- always normalized

Construction:
  from_axis_angle(axis: Vec3, angle: f32) -> Quat
  from_euler(order: EulerOrder, a: f32, b: f32, c: f32) -> Quat
  from_rotation_arc(from: Vec3, to: Vec3) -> Quat
  IDENTITY -> Quat

Operations:
  mul(other: Quat) -> Quat             -- compose rotations
  mul_vec3(v: Vec3) -> Vec3            -- rotate a vector
  slerp(other: Quat, t: f32) -> Quat  -- spherical interpolation
  to_euler(order: EulerOrder) -> (f32, f32, f32)
  inverse() -> Quat
  angle_between(other: Quat) -> f32
```

### 4.4 Affine Transforms

```plaintext
Affine3A
  translation: Vec3
  rotation:    Quat
  scale:       Vec3

Operations:
  transform_point(Vec3) -> Vec3
  transform_vector(Vec3) -> Vec3       -- ignores translation
  inverse() -> Affine3A
  mul(other: Affine3A) -> Affine3A     -- compose transforms
  to_mat4() -> Mat4
  from_mat4(Mat4) -> Affine3A
```

Affine3A avoids a full 4x4 multiply for the common case of rotation + translation + scale. The 'A' suffix indicates aligned storage for SIMD friendliness.

**Split inversion methods**: Two distinct inverse operations are exposed:

- `Inverse() -> Affine3A` — fast path for orthonormal transforms (rotation + translation only, uniform scale). Assumes no shear or non-uniform scale.
- `AffineInverse() -> Affine3A` — general case that handles arbitrary affine transforms including non-uniform scale and shear. Slower but always correct.

User code should prefer `Inverse()` for camera and entity transforms (which are typically orthonormal) and use `AffineInverse()` only when non-uniform scale is involved.

### 4.5 Direction and Rotation Types

```plaintext
Dir2 { vec: Vec2 }     -- guaranteed unit length
Dir3 { vec: Vec3 }     -- guaranteed unit length

Construction (fallible):
  new(vec) -> Option[Self]              -- None if zero-length
  new_unchecked(vec) -> Self            -- caller guarantees normalization

Rot2 { angle: f32 }   -- 2D rotation
  from_degrees(f32) -> Rot2
  from_radians(f32) -> Rot2
  rotate(Vec2) -> Vec2
```

### 4.6 Isometry Types

Rotation plus translation, without scale. Useful for physics bodies and cameras.

```plaintext
Isometry2D { rotation: Rot2, translation: Vec2 }
Isometry3D { rotation: Quat, translation: Vec3 }

Operations:
  transform_point(point) -> point
  inverse() -> Self
  mul(other) -> Self
```

### 4.7 Geometric Primitives

```plaintext
Ray2D   { origin: Vec2, direction: Dir2 }
Ray3D   { origin: Vec3, direction: Dir3 }
AABB    { min: Vec3, max: Vec3 }
Sphere  { center: Vec3, radius: f32 }
Plane   { normal: Dir3, distance: f32 }
Frustum { planes: [Plane; 6] }

Intersection tests:
  ray_aabb(Ray3D, AABB) -> Option[f32]         -- hit distance
  ray_sphere(Ray3D, Sphere) -> Option[f32]
  aabb_aabb(AABB, AABB) -> bool
  sphere_sphere(Sphere, Sphere) -> bool
  frustum_aabb(Frustum, AABB) -> bool           -- visibility culling
  aabb_contains(AABB, Vec3) -> bool
```

### 4.8 Color Types

```plaintext
Color
├── Srgba    { r: u8, g: u8, b: u8, a: u8 }         -- 0-255, display space
├── LinearRgba { r: f32, g: f32, b: f32, a: f32 }    -- 0.0-1.0, linear space
├── Hsla     { h: f32, s: f32, l: f32, a: f32 }
└── Hsva     { h: f32, s: f32, v: f32, a: f32 }

Conversions:
  Srgba ↔ LinearRgba       -- gamma encode/decode
  LinearRgba ↔ Hsla
  LinearRgba ↔ Hsva
  lerp(other, t) -> Self    -- interpolation in linear space
```

All rendering math operates in linear space. Conversion to sRGB happens at the final output stage.

### 4.9 Curves

```plaintext
CubicSegment[P]
  fn position(t: f32) -> P           -- evaluate at t in [0, 1]
  fn velocity(t: f32) -> P           -- first derivative
  fn acceleration(t: f32) -> P       -- second derivative

RationalSegment[P]                    -- weighted control points
  fn position(t: f32) -> P

CubicCurve[P]
  segments: []CubicSegment[P]
  fn position(t: f32) -> P           -- t spans full curve length
  fn iter_positions(steps: uint) -> Iterator[P]
```

Curves support any point type (Vec2, Vec3, f32) and are used for animation paths, camera splines, and easing functions.

### 4.10 Transform Interpolation

For engines with decoupled physics and render rates (e.g., physics at 60 Hz, rendering at 144 Hz), transforms must be interpolated between physics steps to produce smooth visuals.

The `TransformInterpolator` provides a two-phase design:

```plaintext
Phase 1 — Determine method (once per physics tick):
  TransformInterpolator.Prepare(from: Affine3A, to: Affine3A) -> InterpolationParams
    // Analyzes the two keyframes and selects the optimal method:
    //   LERP         — translation-only difference
    //   SLERP        — rotation difference, uniform scale
    //   SCALED_SLERP — rotation + non-uniform scale difference

Phase 2 — Apply (every render frame):
  InterpolationParams.Interpolate(t: f32) -> Affine3A
    // Applies the pre-determined method at fraction t in [0, 1]
```

This separation avoids re-analyzing the transform pair every render frame. The method selection (Phase 1) runs at physics frequency; the actual interpolation (Phase 2) runs at render frequency with minimal cost.

### 4.11 Named Vector Constants

Vector types provide named directional constants for both world-space and model-space conventions:

```plaintext
Vec3 constants:
  ZERO, ONE                              // origin, unit
  UP, DOWN, LEFT, RIGHT, FORWARD, BACK   // world-space directions
```

This avoids magic numbers like `Vec3{0, 1, 0}` scattered through gameplay code and makes directional intent explicit.

## 5. Open Questions

1. Should the library provide float64 variants of all types for editor-precision use cases?
2. Should SIMD be an explicit design layer or an internal optimization detail?
3. Are there additional color spaces needed (OKLCH, CIE-Lab)?
4. Should the math library support a build-time precision switch (compile tag) that changes the base float type for all math types between float32 and float64, or should float64 variants be separate types with explicit suffixes?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft from architecture analysis |
| 0.2.0 | 2026-03-26 | Added split inversion, TransformInterpolator, named constants, precision open question |
