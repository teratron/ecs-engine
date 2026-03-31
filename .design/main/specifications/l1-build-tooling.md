# Build Tooling

**Version:** 0.4.0
**Status:** Draft
**Layer:** concept

## Overview

Defines the engine's build automation, CI pipeline, testing infrastructure, and release documentation conventions. The project is still in a spec-first phase, so this document separates current bootstrap checks from the later `cmd/ci/` and showcase tooling that will be introduced once the codebase justifies them.

## Related Specifications

- [l1-examples-framework.md](l1-examples-framework.md) — Examples validated by CI pipeline
- [l1-diagnostic-system.md](l1-diagnostic-system.md) — Profiling integration used by benchmark tooling
- [l1-app-framework.md](l1-app-framework.md) — Plugin architecture tested by compile-check tools
- [l1-platform-system.md](l1-platform-system.md) — CI matrix targets and cross-compilation strategy

## 1. Motivation

A game engine needs repeatable validation long before it has a large implementation footprint. Build tooling exists to:

1. keep local and CI checks aligned
2. surface regressions early
3. make performance tracking intentional
4. define how release documentation is produced once public versions exist

The tooling strategy must match the current phase. A repository with skeletal packages should not document itself as if a full CI command suite and showcase runner already exist.

## 2. Constraints & Assumptions

- All tooling is written in Go and uses only the standard library when practical (C24).
- CI checks must be runnable locally with a deterministic command path.
- Golden file testing uses a `BLESS` environment variable to regenerate expected output.
- Migration guides use a structured markdown format with YAML frontmatter.
- The long-term CI tool is a single Go binary with subcommands, not a collection of shell scripts.
- Phase 0 may rely on direct `go` commands and `node .magic/scripts/executor.js check-prerequisites --json` before the dedicated Go tooling exists.

## 3. Core Invariants

- **INV-1**: Every CI check that runs in the pipeline must be reproducible locally with the same tool and flags.
- **INV-2**: Bootstrap checks must only claim support for artifacts that actually exist in the repository.
- **INV-3**: Golden file mismatches produce a clear diff, not a generic `test failed` message.
- **INV-4**: Migration guides are required for every breaking API change before a release is tagged.
- **INV-5**: Benchmark regressions beyond a configurable threshold are flagged only after benchmark baselines exist.
- **INV-6**: The CI tool exits with non-zero status on any failure; `--keep-going` mode aggregates all failures before exiting.

## 4. Detailed Design

### 4.1 Phase 0 Bootstrap Checks

Before `cmd/ci/` exists, the project uses a minimal honest validation surface:

```plaintext
go build ./cmd/cli
go test ./...                  # only when tests exist
go test -race ./...            # only when concurrent code and tests exist
node .magic/scripts/executor.js check-prerequisites --json
```

These checks verify that:

- the module builds at its current bootstrap level
- specification metadata stays healthy
- future tooling has a stable baseline to replace

### 4.2 Planned `cmd/ci` Surface

The intended end state is a single Go binary with modular subcommands:

```plaintext
ci format
ci vet
ci lint
ci test
ci test-doc
ci bench
ci compile-check
ci example-check
ci golden-test
ci integration
ci all
```

Planned flags:

- `--keep-going` — Continue on failure, report all errors at the end.
- `--jobs N` — Parallelism level for independent checks.
- `--verbose` — Detailed output for debugging CI failures.

The dedicated CI binary is introduced only after there are enough commands to justify a stable command router.

### 4.3 Golden File Testing

Golden file tests compare program output against stored expected files. They are used for:

- CLI output validation
- error message wording verification
- serialization format stability

Convention:

- Golden files live next to their tests under `testdata/`
- `BLESS=1` regenerates expected files locally
- CI never sets `BLESS=1`

Golden testing activates only when there are stable serializers, CLI outputs, or diagnostics worth freezing.

### 4.4 Benchmark Infrastructure

Benchmarks track performance across commits to detect regressions once the core runtime exists.

Minimum benchmark domains:

- entity spawn/despawn throughput
- query iteration throughput
- command buffer flush time
- scheduler overhead
- transform propagation
- event send/receive throughput

During the current specification phase, benchmark policy exists as design intent. It does not imply that the repository already has benchmark baselines or automated gates.

### 4.5 Example Validation and Showcase Runner

A future `cmd/showcase/` tool bulk-runs examples for validation and documentation:

- filtered execution by category or pattern
- headless runs for CI
- frame-limited execution
- screenshot capture for visual regression
- batched reporting

This runner is explicitly deferred until the examples framework has enough runnable entries to justify dedicated orchestration.

### 4.6 Release Documentation

When public releases begin, the build tooling also governs:

- migration guides for breaking API changes
- release notes for major features
- changelog assembly from per-feature release documents

These artifacts are required only once tagged releases become part of the project lifecycle.

## 5. Open Questions

- Should benchmark baselines be committed to the repo or stored externally?
- Visual regression testing: compare screenshots pixel-by-pixel or use perceptual hash?
- Should migration guides be auto-generated from git commit conventions?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-26 | Initial draft from reference engine tooling analysis |
| 0.2.0 | 2026-03-26 | Added co-located tests, named test commands, error suppression, event testing, serialization roundtrip tests |
| 0.3.0 | 2026-03-26 | Added parallel test dispatching with work stealing and cooperative main thread |
| 0.4.0 | 2026-03-31 | Split current bootstrap checks from planned CI/showcase tooling to match the repository's actual phase. |
| — | — | Planned examples: [examples/stress_test/](../../../examples/stress_test/) |
