# Property Inspector Specification

| Metadata | Value |
| :--- | :--- |
| **Layer** | 1 (concept) |
| **Status** | Draft |
| **Version** | 0.1.0 |
| **Related Specifications** | [ecs-engine: l1-type-registry.md], [ecs-engine: l1-component-system.md] |

## Overview

The Property Inspector (or "Details Panel") allows the user to view and modify the components of the currently selected entity. It uses the `Type Registry` to automatically generate UI widgets for component fields.

## 1. Motivation

Manually writing UI code for every component property is a waste of time and error-prone. The Inspector must be dynamic: if a developer adds a field to a Go struct and registers it, it should automatically appear in the Editor.

## 2. Constraints & Assumptions

- **Reflection-driven**: Heavy reliance on `Type Registry` metadata.
- **Undo/Redo**: Every change must be recordable.
- **Live Updating**: If a system changes a component value (e.g., physics), the Inspector should reflect the change in real-time.

## 3. Core Invariants

- **INV-1**: Changes in the Inspector MUST be committed to the World via a Command.
- **INV-2**: The Inspector MUST support "Batch Editing" (editing the same component across multiple selected entities).
- **INV-3**: Types not registered in the `Type Registry` will not be visible in the Inspector.

## 4. Detailed Design

### 4.1 Widget Mapping

The Inspector maps Go types to UI widgets:
- `bool` → Checkbox.
- `float32/64` → Input box or Slider (if range attribute is present).
- `string` → Text field.
- `Vec3` → Three-column float input.
- `Color` → Color picker.
- `Handle[T]` → Asset drop target.

### 4.2 Component Control

- **Add Component**: Searchable list of all registered component types.
- **Remove Component**: Delete button with confirmation.
- **Enable/Disable**: Toggle for components that support a "disabled" state.

## 5. Open Questions

- How to handle custom UI widgets for specific components (e.g., a curve editor for animation)?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft |
