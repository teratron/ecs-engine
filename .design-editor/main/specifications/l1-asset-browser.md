# Asset Browser Specification

| Metadata | Value |
| :--- | :--- |
| **Layer** | 1 (concept) |
| **Status** | Draft |
| **Version** | 0.1.0 |
| **Related Specifications** | [ecs-engine: l1-asset-system.md], [ecs-engine: l1-asset-formats.md] |

## Overview

The Asset Browser is a visual interface for managing the project's disk resources (textures, meshes, sounds, scenes). It acts as a bridge between the filesystem and the engine's internal `Asset Server`.

## 1. Motivation

Working with raw paths is inefficient. The Asset Browser provides a visual overview, searchability, and drag-and-drop workflows for scene building.

## 2. Constraints & Assumptions

- **Filesystem Mirroring**: Must stay in sync with the actual files in the `assets/` directory.
- **Async Loading**: Thumbnails and metadata should be loaded in the background.

## 3. Core Invariants

- **INV-1**: All operations (rename, delete, move) in the Browser MUST physically affect the filesystem.
- **INV-2**: The Browser MUST handle "Stale Handles": if a file is deleted, all engine handles to it must be invalidated.
- **INV-3**: Dragging an asset into a Viewport or Inspector MUST pass the `AssetHandle` to the target.

## 4. Detailed Design

### 4.1 Folder Navigation

- Breadcrumb navigation and folder sidebar.
- Favorites and recent assets.

### 4.2 Asset Thumbnails

The Editor generates and caches thumbnails for supported types:
- **Images**: Downscaled preview.
- **Meshes**: Small render of the model with a neutral background.
- **Scenes**: Icon based on scene content.

### 4.3 Context Actions

- **Import**: Drag files from the OS into the Browser to import them.
- **Export/Package**: Tools for bundling assets for distribution.
- **Re-import**: Force reload of an asset if the source file changed externally.

## 5. Open Questions

- Should the Browser support tagging or external metadata files?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft |
