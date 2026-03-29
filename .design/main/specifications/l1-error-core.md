# Error Core Specification

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

A game engine requires more than simple string-based errors. This specification defines a structured taxonomy for errors, categorizing them by severity, origin, and recoverability. It also establishes the foundation for error localization and developer-centric troubleshooting.

## 1. Error Taxonomy

All engine errors are categorized into the following dimensions:

### 1.1 Severity Levels

- **Fatal**: Unrecoverable engine state. Requires immediate process termination (Panic or os.Exit).
- **Recoverable**: Error that can be handled by the caller (e.g., entity not found).
- **Warning**: Potential issue that doesn't stop execution (e.g., duplicate system registration).
- **Debug**: Detailed trace info for development.

### 1.2 Target Audience

- **Developer Error**: Misuse of the API (e.g., querying for a component not added to the world).
- **User Error**: Runtime issues caused by end-user input or malformed assets (e.g., logic error in a script).
- **System Error**: OS-level or hardware issues (e.g., out of memory, file not found).

## 2. Structured Error Format

All errors MUST implement the `EngineError` interface:

```go
type EngineError interface {
    error
    Code() string        // E0001
    Severity() Severity  // Fatal, Recoverable, etc.
    Module() string      // "ecs", "render", "physics"
    Solution() string    // Actionable advice for the developer
}
```

### 2.1 Error Codes (E-Series)

Codes follow the format `E[Category][ID]`:
- `E0000-E0999`: Core ECS (Entity/Component/World)
- `E1000-E1999`: Scheduling & Systems
- `E2000-E2999`: Render & Assets
- `E3000-E3999`: Physics & Collision

## 3. Localization & UX

Error messages are stored externally to the code to allow for localization:
- **Registry**: A `locales/errors.en.json` file contains mapping from Error Codes to templates.
- **Templates**: Support placeholders (e.g., "System {name} has cyclic dependency").

## 4. Directives for Implementation

- **No Silencing**: Never use `_ = function()` for calls that return `EngineError`.
- **Trace Context**: Errors should automatically capture the stack trace when generated in `Debug` builds.
- **Panic Policy**: Panics are ONLY permitted for `Fatal` developer errors where continued execution would result in data corruption.

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-27 | Initial draft |
| — | — | Planned examples: `examples/diagnostic/` |
