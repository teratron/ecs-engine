# Editor Framework Specification

| Metadata | Value |
| :--- | :--- |
| **Layer** | 1 (concept) |
| **Status** | Draft |
| **Version** | 0.1.0 |
| **Related Specifications** | [ecs-engine: l1-ui-system.md], [ecs-engine: l1-scene-system.md] |

## Overview

The Editor Framework defines the architectural shell for the standalone GUI Editor. Unlike the core engine which is data-driven and non-visual, the Editor is a heavy-UI application that orchestrates multiple "sub-apps" (the Editor UI itself and the "Game Scene" being edited).

## 1. Motivation

A modern engine requires an integrated environment for scene composition. Relying solely on code-based configuration (as in early versions of basic ECS frameworks) slows down iteration for designers and artists. The Editor must bridge the gap between engine data (entities, components) and visual manipulation.

## 2. Constraints & Assumptions

- **Shared State**: The Editor and the Scene-under-edit can run in the same process but should be isolated (using Sub-Apps).
- **Extensibility**: Third-party plugins must be able to add new panels and tools to the Editor.
- **Independence**: The Editor must remain a shell: all game data is stored in standard Scenes/Assets formats.

## 3. Core Invariants

- **INV-1**: All Editor interactions that modify game data MUST be wrapped in a Command (for Undo/Redo).
- **INV-2**: The "Game World" being edited MUST NOT be aware it is running inside the Editor, except through standard engine interfaces.
- **INV-3**: The Editor UI is built using the engine's OWN `ui-system`, fulfilling the "dogfooding" principle.


## 4. Detailed Design

### 4.1 Shell Architecture

The Editor is composed of:
1. **Docking Host**: The root window container that manages the spatial arrangement of panels.
2. **Global Editor State**: Tracks current selection, application mode (Edit/Play/Pause), and active project path.
3. **Workspace Layouts**: Saved configurations of panels (e.g., "Default", "Animation", "2D Layout").


### 4.2 Panel System

Panels are independent plugins that register with the Editor Framework:
- **Registry**: `EditorApp.RegisterPanel("id", widget_factory)`
- **Lifecycle**: Panels handle `Open`, `Close`, `Focus`, `SaveState`.


### 4.3 Input Dispatch

The Editor must distinguish between inputs targeted at:
1. **Editor UI** (buttons, text fields).
2. **Viewport Manipulation** (orbiting camera).
3. **In-game Interaction** (when the game is "running" in-editor).

## 5. Open Questions

- Should the Editor run as an `Engine Sub-App` or as a separate process communicating via IPC?
- How to handle graphics backend sharing if both the Editor and Viewport use the same GPU?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft |
