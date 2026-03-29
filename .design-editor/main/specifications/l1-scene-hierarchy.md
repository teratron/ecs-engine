# Scene Hierarchy Specification

| Metadata | Value |
| :--- | :--- |
| **Layer** | 1 (concept) |
| **Status** | Draft |
| **Version** | 0.1.0 |
| **Related Specifications** | [ecs-engine: l1-hierarchy-system.md], [ecs-engine: l1-entity-system.md] |

## Overview

The Scene Hierarchy panel provides a tree-based visualization of all entities in the current Scene. It is the primary tool for organizational management of the game world.

## 1. Motivation

Game worlds often consist of thousands of entities. A flat list is unmanageable. The Hierarchy panel leverages the engine's parent-child relationship system to show the logical structure of the world.

## 2. Constraints & Assumptions

- **Reactivity**: The panel must update immediately when entities are spawned or despawned in the world.
- **Performance**: Must remain responsive even with deep hierarchies (e.g., thousands of entities).

## 3. Core Invariants

- **INV-1**: The Hierarchy tree MUST exactly represent the `Parent/Child` components in the World.
- **INV-2**: Moving an entity in the tree MUST trigger a Command to change its `Parent` component.
- **INV-3**: Selection in the Hierarchy MUST be synchronized with selection in the Viewport.

## 4. Detailed Design

### 4.1 Tree Representation

The panel uses a lazy-loading tree widget:
- **Filtering**: Search bar to filter entities by name or component type.
- **Visibility Toggles**: Buttons to hide/show entities in the viewport.
- **Locking**: Prevent accidental selection/modification.

### 4.2 Drag and Drop

- **Reparenting**: Drag an entity onto another to make it a child.
- **Reordering**: Move siblings to change iteration/render order if applicable.
- **Prefab Creation**: Drag an entity into the Asset Browser to save it as a scene/prefab.

## 5. Open Questions

- Should we show "system-internal" entities (hidden from the user) or only "user-land" entities?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft |
