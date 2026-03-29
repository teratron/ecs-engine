# Console & Log Viewer Specification

| Metadata | Value |
| :--- | :--- |
| **Layer** | 1 (concept) |
| **Status** | Draft |
| **Version** | 0.1.0 |
| **Related Specifications** | [ecs-engine: l1-diagnostic-system.md], [l1-editor-framework.md](l1-editor-framework.md) |

## Overview

The Console & Log Viewer is an integrated panel within the Editor that captures, displays, and filters log output from the engine and game systems in real-time. It provides structured access to diagnostics, warnings, errors, and user-defined messages.

## 1. Motivation

Debugging a game engine without visible log output requires external terminal monitoring. An integrated console accelerates the development loop by showing relevant messages in-context, with clickable source references and filtering by severity or subsystem.

## 2. Constraints & Assumptions

- **Performance**: Must handle high log throughput (thousands of messages per second during stress tests) without freezing the Editor UI.
- **`log/slog` Integration**: The engine uses Go's `log/slog` (stdlib). The Console hooks into the slog handler chain.
- **Non-Blocking**: Log capture MUST NOT slow down engine execution. The console reads from a ring buffer asynchronously.

## 3. Core Invariants

- **INV-1**: The Console MUST display messages with correct timestamps, severity level, and source system attribution.
- **INV-2**: Filtering and search MUST operate on the in-memory buffer without re-reading the log source.
- **INV-3**: Error and Warning messages MUST persist across "Clear" operations until explicitly dismissed, unless the user opts out.
- **INV-4**: The Console panel MUST NOT drop messages — if the display cannot keep up, it must buffer and catch up when the user scrolls down.

## 4. Detailed Design

### 4.1 Log Capture Pipeline

The Console installs a custom `slog.Handler` that:

1. Receives structured log records from the engine.
2. Writes them into a ring buffer (fixed size, configurable, e.g., 10,000 entries).
3. Notifies the Console UI of new entries via an event.

### 4.2 Message Display

Each log entry is rendered as a row with:

- **Severity Icon**: Color-coded icon (Debug=gray, Info=blue, Warn=yellow, Error=red).
- **Timestamp**: Relative to session start or absolute (user toggle).
- **Source**: The slog group/logger name (e.g., `render.pipeline`, `ecs.world`, `game.player`).
- **Message**: The log message text.
- **Attributes**: Expandable section showing key-value pairs from structured logging.

### 4.3 Filtering & Search

- **Severity Filter**: Toggle buttons for Debug, Info, Warn, Error — show/hide each level.
- **Source Filter**: Dropdown or text input to filter by source system (e.g., show only `ecs.*`).
- **Text Search**: Incremental search across message text and attribute values.
- **Regex Support**: Optional toggle for regular expression search.

### 4.4 Actions

- **Clear**: Flush the visible buffer (optionally keep errors/warnings pinned).
- **Copy**: Copy selected messages to clipboard (formatted as plain text or JSON).
- **Pause/Resume**: Freeze the display while the engine continues logging (buffer continues filling).
- **Source Link**: If the log record includes a source file and line number, clicking it opens the corresponding file in the user's configured external editor.

### 4.5 Console Input (Optional)

An optional command input line at the bottom of the panel for:

- Executing engine debug commands (e.g., `spawn entity`, `set timescale 0.5`).
- Evaluating simple expressions against the World state.
- This feature is gated behind an "Advanced" toggle and is a post-MVP addition.

## 5. Open Questions

- Should the console support multiple tabs (e.g., one per subsystem)?
- How to handle log output during Play mode vs Edit mode — separate buffers or merged?
- Should crash logs from previous sessions be loadable in the Console?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft |
