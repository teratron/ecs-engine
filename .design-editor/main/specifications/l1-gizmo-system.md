# Gizmo System Specification

| Metadata | Value |
| :--- | :--- |
| **Layer** | 1 (concept) |
| **Status** | Draft |
| **Version** | 0.1.0 |
| **Related Specifications** | [l1-viewport-system.md](l1-viewport-system.md), [ecs-engine: l1-math-system.md], [ecs-engine: l1-input-system.md] |

## Overview

The Gizmo System provides interactive visual handles for manipulating entity transforms (position, rotation, scale) directly in the Viewport. Gizmos are editor-only overlays rendered on top of the game scene.

## 1. Motivation

Editing transform values through numeric inputs in the Property Inspector is slow and unintuitive for spatial operations. Visual gizmos allow designers to drag, rotate, and scale objects directly in 3D/2D space with immediate visual feedback — the standard workflow in all modern game editors.

## 2. Constraints & Assumptions

- **Editor-Only**: Gizmos are never part of the game build. They exist exclusively in the Editor's render layer.
- **Camera-Aware**: Gizmo size must remain constant on screen regardless of distance from the camera (scale-compensated).
- **Multi-Selection**: Gizmos must support operating on multiple selected entities simultaneously, with a shared pivot point.
- **Coordinate Spaces**: Support Local, World, and Parent coordinate spaces for transformations.

## 3. Core Invariants

- **INV-1**: Every gizmo interaction MUST produce a Command (for Undo/Redo integration).
- **INV-2**: Gizmo rendering MUST occur after the game scene render pass and before the final UI composite.
- **INV-3**: Gizmo hit-testing MUST use ray-casting from the viewport camera, not screen-space bounding boxes.
- **INV-4**: Snapping values MUST be configurable per-axis and per-mode (translate, rotate, scale).

## 4. Detailed Design

### 4.1 Gizmo Modes

The system supports three primary manipulation modes, toggled by hotkeys or toolbar buttons:

- **Translate (Move)**: Three axis arrows (X, Y, Z) + three plane handles (XY, XZ, YZ) + center handle (free move).
- **Rotate**: Three circular arcs around each axis + trackball rotation.
- **Scale**: Three axis handles + uniform scale center handle.

### 4.2 Visual Feedback

- **Axis Coloring**: X = Red, Y = Green, Z = Blue (industry standard).
- **Hover Highlight**: Active axis highlights on mouse hover (brighter color + thicker line).
- **Active Drag**: During manipulation, show a ghost outline of the original position and delta values as floating text.
- **Grid Snapping Indicator**: When snapping is active, display snap increments as faint grid lines around the active axis.

### 4.3 Pivot and Space

- **Pivot Point**: Configurable between "Center" (median of selection), "Active" (last selected entity), "Individual Origins" (each entity uses its own origin).
- **Coordinate Space**: Toggle between Local (entity's own axes), World (global axes), and Parent (parent entity's axes).

### 4.4 Custom Gizmos

Plugins can register custom gizmos for specific component types:

- **Volume Gizmos**: Wireframe spheres for audio sources, light radii, trigger zones.
- **Path Gizmos**: Editable spline handles for path-following components.
- **Registration**: `EditorApp.RegisterGizmo[T](draw_func, interact_func)` — where `T` is the component type.

### 4.5 Snapping

- **Grid Snap**: Configurable grid size (e.g., 0.25, 0.5, 1.0 units).
- **Angle Snap**: Configurable angle increments (e.g., 5°, 15°, 45°, 90°).
- **Scale Snap**: Configurable scale increments (e.g., 0.1, 0.25, 1.0).
- **Surface Snap**: Snap to the surface of nearby meshes (advanced feature).

## 5. Open Questions

- Should gizmos support multi-viewport interaction (e.g., dragging in one viewport while previewing in another)?
- What is the mechanism for gizmo occlusion — should gizmos be visible through scene geometry or use X-ray rendering?
- How to handle gizmo interaction during Play mode (read-only gizmos vs. no gizmos)?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft |
