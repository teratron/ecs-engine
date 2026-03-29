# Viewport System Specification

| Metadata | Value |
| :--- | :--- |
| **Layer** | 1 (concept) |
| **Status** | Draft |
| **Version** | 0.1.0 |
| **Related Specifications** | [ecs-engine: l1-render-core.md], [ecs-engine: l1-camera-and-visibility.md] |

## Overview

The Viewport System integrates the engine's 3D/2D rendering output into the Editor's UI. It allows the user to see and interact with the game world directly within a panel.

## 1. Motivation

An editor is useless without a visual representation of the game world. The viewport must be a high-performance bridge that allows rendering a "Sub-App" (the game) into a texture that the main "App" (the Editor UI) can display.

## 2. Constraints & Assumptions

- **Performance**: Viewport rendering should not significantly lag the Editor UI.
- **Resolution**: The viewport must handle resizing dynamically to match the UI panel size.
- **Multiple Viewports**: The system should support multiple viewports (e.g., Top, Side, Perspective).

## 3. Core Invariants

- **INV-1**: The Viewport MUST render the game world to an intermediate texture (Render-to-Texture).
- **INV-2**: Mouse/Keyboard events over the viewport MUST be translated from "Panel Space" to "World Space" using the viewport's camera.
- **INV-3**: The Viewport must support debug overlays (gizmos, grids, diagnostic info) that are NOT visible in the final game.

## 4. Detailed Design

### 4.1 Rendering Bridge

The Viewport uses the **Extract Pattern** from the render pipeline:
1. **Game Scene** renders to an internal `RenderTexture`.
2. **Editor UI** uses that `RenderTexture` as an image source for a `ViewportWidget`.

### 4.2 Interaction Logic

- **Picking**: When the user clicks in the viewport, a ray is cast through the viewport camera to identify the clicked entity.
- **Input Capturing**: The viewport can "capture" the mouse (e.g., for FPS-style orbiting).

### 4.3 Viewport Modes

- **Perspective/Orthographic** toggles.
- **Shading Modes**: Wireframe, Lit, Unlit, Physics Debug.

## 5. Open Questions

- How to handle high-DPI scaling across different monitors for the UI-integrated viewport?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft |
