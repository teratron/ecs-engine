# Undo/Redo System Specification

| Metadata | Value |
| :--- | :--- |
| **Layer** | 1 (concept) |
| **Status** | Draft |
| **Version** | 0.1.0 |
| **Related Specifications** | [l1-editor-framework.md](l1-editor-framework.md), [ecs-engine: l1-command-system.md] |

## Overview

The Undo/Redo System provides a command-based history mechanism for all editor actions. Every mutation to the game world that originates from the Editor UI is captured as a reversible command, enabling users to step backward and forward through their editing history.

## 1. Motivation

Undo/Redo is a fundamental expectation in any creative tool. Without it, a single misclick can destroy hours of work. The system must be robust enough to handle complex operations (e.g., reparenting a subtree of entities) while remaining transparent to the user.

## 2. Constraints & Assumptions

- **Command-Based**: All mutations are encapsulated in Command objects with `Execute()` and `Undo()` methods.
- **Memory Limits**: The history stack must be bounded (configurable max depth or memory budget).
- **Serializable**: Commands should be serializable for crash recovery (optional, advanced feature).
- **Branch Pruning**: When the user undoes N steps and then performs a new action, the "future" branch is discarded.

## 3. Core Invariants

- **INV-1**: Every undoable action MUST implement both `Execute()` and `Undo()`. If `Undo()` cannot be implemented for an action, it MUST be marked as a "history barrier" that clears the undo stack.
- **INV-2**: `Undo()` MUST restore the World state to exactly what it was before `Execute()`, including component values, entity existence, and hierarchy.
- **INV-3**: Compound operations (e.g., "Duplicate Entity" = spawn + copy components + reparent) MUST be grouped into a single undo step using `BeginGroup()` / `EndGroup()`.
- **INV-4**: The system MUST NOT interfere with runtime game logic — it operates exclusively on the Editor's command layer.

## 4. Detailed Design

### 4.1 Command Interface

Every editor action implements the `EditorCommand` interface:

- `Name() string` — human-readable description for the Edit menu (e.g., "Move Entity", "Delete 3 Entities").
- `Execute()` — apply the change to the World.
- `Undo()` — reverse the change.
- `Redo()` — optional override; defaults to calling `Execute()` again.

### 4.2 History Stack

The history is a linear stack with a cursor:

- **Undo**: Move cursor back, call `Undo()` on the current command.
- **Redo**: Move cursor forward, call `Execute()` on the next command.
- **New Action**: Truncate everything after the cursor, push the new command.
- **Max Depth**: Configurable (default: 100 steps). Oldest commands are evicted when the limit is reached.

### 4.3 Compound Commands

For multi-step operations that should appear as a single action:

- `history.BeginGroup("Duplicate Entities")`
- ... multiple individual commands ...
- `history.EndGroup()`
- Undo/Redo treats the entire group as one step.

### 4.4 Snapshot-Based Commands

For operations where computing the inverse is too complex (e.g., "Apply Physics Simulation"):

- Capture a snapshot of affected components before the operation.
- `Undo()` restores from the snapshot.
- Trade-off: higher memory usage but guaranteed correctness.

### 4.5 Integration Points

- **Property Inspector**: Field changes produce `SetComponentFieldCommand`.
- **Scene Hierarchy**: Reparenting produces `ReparentEntityCommand`.
- **Viewport Gizmos**: Transform drags produce `TransformEntityCommand`.
- **Asset Browser**: Asset operations may produce `ImportAssetCommand`, etc.

## 5. Open Questions

- Should we support "Undo History" panel (visual list of all past commands with preview)?
- How to handle undo across scene transitions (user switches scenes — is the history per-scene or global)?
- Should the system support collaborative undo (multiple users editing the same scene)?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft |
