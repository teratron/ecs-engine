# CLI Tooling

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

Defines the structure, commands, and functionality of the internal command-line interface (`cmd/cli/`) used for scaffolding projects, managing assets, and executing engine-specific routines.

## Related Specifications

- [l1-build-tooling.md](l1-build-tooling.md) - CI pipeline commands vs. user-facing CLI.
- [l2-codegen-tools.md](l2-codegen-tools.md) - Code generation commands executed via this CLI.

## 1. Motivation

A dedicated CLI provides a unified entry point for developers using the ECS engine. It replaces fragmented scripts by standardizing common project tasks (e.g., scaffolding boundaries, managing components).

## 2. Constraints & Assumptions

- The CLI uses standard Go `flag` or a lightweight router rather than heavy third-party frameworks unless required.
- Commands must be deterministic and support scripting (e.g., `--json` output flags).

## 3. Core Invariants

- **INV-1**: All scaffolding commands must safely skip or prompt before overwriting existing files.
- **INV-2**: The CLI must print a structured help menu when invoked without arguments.

## 4. Detailed Design

### 4.1 Scaffold Command

Used to generate boilerplate structures like missing `pkg/` interfaces.

**Usage:**

```bash
ecs-engine scaffold boundaries
```

### 4.2 Build and Run

Wrapper around `go build` to handle engine-specific tags (e.g., SIMD, debug features).

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-30 | Initial Draft |
| — | — | Planned examples: `examples/cli_stub/` |
