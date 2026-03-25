# Input System

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

The input system provides a device-agnostic abstraction for keyboard, mouse, gamepad, and touch input. It operates on two layers: **Resources** for polling current state (`ButtonInput[T]` with pressed/just_pressed/just_released semantics) and **Events** for frame-granular input events. A pointer abstraction unifies mouse, touch, and pen into a single interaction model. A picking backend enables pointer-based hit testing against entities with observer-driven event bubbling through the entity hierarchy.

## Related Specifications

- [event-system.md](event-system.md) — Input events use the engine's event system
- [hierarchy-system.md](hierarchy-system.md) — Pointer events bubble through parent-child hierarchy
- [app-framework.md](app-framework.md) — Input systems run in PreUpdate schedule

## 1. Motivation

Games must react to player input from diverse devices. Without a structured input system:

- Device-specific code scatters across gameplay logic.
- "Just pressed" semantics (fire-once per press) are reimplemented in every system.
- Hit testing (which entity did the user click?) has no standard solution.
- Supporting multiple input devices simultaneously (keyboard + gamepad) requires ad-hoc code.

The input system centralizes all of this into a consistent, extensible abstraction.

## 2. Constraints & Assumptions

- Input state is updated once per frame in the PreUpdate schedule, before gameplay systems run.
- "Just pressed" and "just released" flags are valid for exactly one frame, then cleared.
- All input types are represented as Resources and Events — no special system parameter types.
- The input system does not perform input mapping (binding actions to keys). That is a higher-level concern for a separate action-mapping specification.
- Zero external dependencies (C24).

## 3. Core Invariants

- **INV-1**: `JustPressed(key)` returns true for exactly one frame after the key transitions from released to pressed.
- **INV-2**: `JustReleased(key)` returns true for exactly one frame after the key transitions from pressed to released.
- **INV-3**: Input events are ordered within a frame — multiple presses/releases in a single frame are all recorded.
- **INV-4**: Pointer events propagate through the entity hierarchy in a well-defined order (deepest child first, then ancestors).
- **INV-5**: Input resources are cleared/updated by engine systems, not by user systems.

## 4. Detailed Design

### 4.1 ButtonInput[T] Resource

A generic resource parameterized by button type. Tracks three sets per frame:

```
ButtonInput[T]
  - pressed       Set[T]   // currently held down
  - just_pressed  Set[T]   // transitioned to pressed this frame
  - just_released Set[T]   // transitioned to released this frame

Methods:
  Press(button T)
  Release(button T)
  Pressed(button T) bool
  JustPressed(button T) bool
  JustReleased(button T) bool
  AnyPressed() bool
  AnyJustPressed() bool
  GetPressed() []T
  GetJustPressed() []T
  GetJustReleased() []T
  Clear()                   // called by engine at end of frame
```

Concrete instantiations: `ButtonInput[KeyCode]`, `ButtonInput[MouseButton]`, `ButtonInput[GamepadButton]`.

### 4.2 Device Types

**Keyboard:**

```
KeyCode — enumeration of physical key codes (KeyA..KeyZ, Digit0..Digit9,
          ArrowUp, Space, Enter, Escape, ShiftLeft, ControlLeft, etc.)

Events:
  KeyboardInput { KeyCode, Pressed bool }
```

**Mouse:**

```
MouseButton — Left, Right, Middle, Back, Forward, Other(uint16)

Resources:
  ButtonInput[MouseButton]
  CursorPosition { X, Y float64 } // window-relative pixel coordinates

Events:
  MouseButtonInput { Button MouseButton, Pressed bool }
  MouseMotion { DeltaX, DeltaY float64 }
  MouseWheel { X, Y float64 }
  CursorMoved { Position Vec2 }
  CursorEntered { WindowEntity Entity }
  CursorLeft { WindowEntity Entity }
```

**Gamepad:**

```
GamepadButton — South, East, North, West, LeftTrigger, RightTrigger,
                LeftBumper, RightBumper, DPadUp, DPadDown, DPadLeft,
                DPadRight, Start, Select, LeftStick, RightStick

GamepadAxis — LeftStickX, LeftStickY, RightStickX, RightStickY,
              LeftTrigger, RightTrigger

Resources:
  ButtonInput[GamepadButton]
  Axis[GamepadAxis]          // normalized -1.0 to 1.0 (or 0.0 to 1.0 for triggers)

Events:
  GamepadConnected { ID GamepadID }
  GamepadDisconnected { ID GamepadID }
  GamepadButtonInput { Gamepad GamepadID, Button GamepadButton, Pressed bool }
  GamepadAxisChanged { Gamepad GamepadID, Axis GamepadAxis, Value float64 }
```

**Touch:**

```
Events:
  TouchInput { Phase TouchPhase, ID uint64, Position Vec2, Force float64 }
  TouchPhase — Started, Moved, Ended, Cancelled
```

### 4.3 Input Update Flow

```
PreUpdate schedule:
  1. Read platform events from window backend
  2. Translate platform events into engine event types
  3. Clear just_pressed and just_released sets from previous frame
  4. For each input event:
     a. Update ButtonInput[T] resources (press/release)
     b. Update position resources (CursorPosition, Axis)
     c. Send corresponding engine events
```

### 4.4 Pointer Abstraction

Mouse, touch, and pen are unified into a single Pointer model for interaction:

```
PointerID — identifies a pointer (mouse = single pointer, touch = one per finger)

Pointer Events (entity-targeted):
  PointerOver   { PointerID, Entity, Position Vec2 }
  PointerOut    { PointerID, Entity }
  PointerDown   { PointerID, Entity, Button PointerButton }
  PointerUp     { PointerID, Entity, Button PointerButton }
  PointerClick  { PointerID, Entity, Button PointerButton }
  PointerMove   { PointerID, Entity, Delta Vec2 }
  DragStart     { PointerID, Entity }
  Drag          { PointerID, Entity, Delta Vec2 }
  DragEnd       { PointerID, Entity }
  DragEnter     { PointerID, Entity }
  DragLeave     { PointerID, Entity }
  DragDrop      { PointerID, Entity }
```

### 4.5 Picking Backend

Picking determines which entity a pointer is interacting with:

```
1. Each frame, the picking system projects pointer positions into the scene
2. Hit tests run against entities with a Pickable component
3. Results are sorted by depth (front to back)
4. The topmost hit entity receives the pointer event
5. Events bubble up through the entity hierarchy (see hierarchy-system.md)
6. Any entity in the chain can mark the event as handled to stop bubbling
```

Picking backends are pluggable — different backends for 2D sprites, 3D meshes, and UI rectangles. Each backend implements a hit-test function and registers it with the picking system.

### 4.6 Event Bubbling

When a pointer event targets an entity, it propagates upward:

```mermaid
graph TB
    A[Button Entity - receives PointerClick] --> B[Panel Entity - observes PointerClick]
    B --> C[Window Entity - observes PointerClick]
```

At each level, observers registered on that entity can handle the event and optionally stop propagation. This matches the DOM event model: capture phase (top-down) is not supported; only bubble phase (bottom-up).

## 5. Open Questions

- Should there be a built-in action mapping layer (e.g., "Jump" maps to Space + GamepadSouth), or is that a separate specification?
- How should input be handled during state transitions (e.g., should input be consumed when transitioning from Menu to Playing)?
- Should gamepads support per-player assignment natively, or is that a game-level concern?
- Dead zone configuration for gamepad axes — global resource or per-gamepad?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft |
