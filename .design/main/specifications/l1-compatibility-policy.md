# Compatibility and Versioning Policy

**Version:** 0.2.0
**Status:** Draft
**Layer:** concept

## Overview

This specification defines the versioning strategy for the ECS engine and its relationship with the Go toolchain. As a high-performance engine, we often utilize bleeding-edge compiler and runtime features, necessitating a clear policy on Go version support and backwards compatibility.

## 1. Engine Versioning (SemVer)

The engine follows [Semantic Versioning 2.0.0](https://semver.org/).

- **MAJOR**: Breaking API changes or fundamental architectural shifts (e.g., switching from Table to something else).
- **MINOR**: New features, new diagnostic metrics, or significant performance optimizations that maintain backward API compatibility.
- **PATCH**: Bug fixes, minor documentation updates, or internal refactoring.

## 2. Go Version Compatibility Matrix

The engine maintains a strict "Latest + 1" support policy.

| Engine Version | Min Go Version | Primary Features Used |
| :--- | :--- | :--- |
| **v0.1.x** | Go 1.26.1 | Green Tea GC, SIMD, Enhanced `new`, Iterators (Go 1.23+) |

### 2.1 Support Window

- **Active Development**: Targets the absolute latest stable Go release (currently Go 1.26.1).
- **Maintenance**: Supports the current (N) and previous (N-1) major Go releases.
- **Deprecation**: Support for a Go version is dropped when it becomes N-2 relative to the current Go release, unless otherwise specified.

## 3. Toolchain Enforcement

- **go.mod**: The `go` directive in `go.mod` MUST reflect the minimum supported Go version.
- **Build Tags**: Features dependent on specific Go versions (e.g., experimental SIMD) must be guarded by build tags (e.g., `//go:build go1.26`).

## 4. Breaking Changes Policy

### 4.1 API Stability

- During `0.x.y` (alpha/beta), breaking changes are permitted in MINOR versions.
- After `1.0.0`, breaking changes require a MAJOR version bump.

### 4.2 Go Runtime Dependencies

If a new Go version introduces a critical performance feature (like a new GC or memory allocator optimization) that requires a breaking change in our memory layout, it will trigger a MAJOR or MINOR version bump depending on the impact.

## 5. Development Workflow

- All PRs are tested against the current and previous Go version in CI.
- Benchmarks are run on the latest Go version to ensure performance gains from the new runtime are captured.

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-27 | Initial draft |
| 0.2.0 | 2026-03-30 | Added C26 example correlation placeholder for compatibility validation |
| — | — | Planned examples: `examples/app/version_compatibility/` |
