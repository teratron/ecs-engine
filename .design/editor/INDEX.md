# Editor Workspace Registry

**Version:** 1.0.0
**Status:** Active

## Overview

Local registry of specifications for the GUI Editor application. The editor is a standalone tool built on top of the generic engine framework, providing a high-level UI for scene composition, asset management, and live debugging.

## P1 — Editor Core

| File | Description | Status | Layer | Version |
| :--- | :--- | :--- | :--- | :--- |
| [editor-framework.md](specifications/editor-framework.md) | Main application shell, docking layout, multi-window support | Draft | concept | 0.1.0 |
| [viewport-system.md](specifications/viewport-system.md) | Integrated rendering viewport within the UI hierarchy | Draft | concept | 0.1.0 |
| [scene-hierarchy.md](specifications/scene-hierarchy.md) | Entity tree visualization and management (reparenting, selection) | Draft | concept | 0.1.0 |
| [property-inspector.md](specifications/property-inspector.md) | Reflection-based component field editing and widget mapping | Draft | concept | 0.1.0 |
| [asset-browser.md](specifications/asset-browser.md) | Visual asset management: thumbnails, folder navigation, drag-and-drop | Draft | concept | 0.1.0 |

## P2 — Tooling & UX

| File | Description | Status | Layer | Version |
| :--- | :--- | :--- | :--- | :--- |
| [gizmo-system.md](specifications/gizmo-system.md) | Interactive transform handles (move, rotate, scale) and snapping | Draft | concept | 0.1.0 |
| [undo-redo-system.md](specifications/undo-redo-system.md) | Command-based undo/redo for all editor actions | Draft | concept | 0.1.0 |
| [console-log.md](specifications/console-log.md) | Integrated engine log viewer with filtering and source mapping | Draft | concept | 0.1.0 |

## Meta Information

- **Maintainer**: Editor Team
- **Last Updated**: 2026-03-25
