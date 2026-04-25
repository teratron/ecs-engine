---
phase: 6
name: "UI, Tooling & Quality"
status: Hold
subsystem: "pkg/ui, pkg/window, pkg/diag, cmd/cli, pkg/build, pkg/codegen, pkg/platform"
requires:
  - "Phase 1–3 Stable"
provides:
  - "Declarative definition layer (JSON / templates)"
  - "Window + multi-window + platform abstraction"
  - "Diagnostics + profiling overlay + gizmos + error codes"
  - "UI layout, interaction, text, widgets, styling"
  - "Build pipeline + CI + golden testing"
  - "CLI scaffolding + asset management commands"
  - "Platform tier matrix + capabilities + build tags"
  - "AI assistant plugin protocol"
  - "Examples framework + lifecycle"
  - "Compatibility policy + Go toolchain matrix"
  - "Structured error taxonomy (E-series)"
  - "Standardized benchmark suite + regression CI gates"
  - "Codegen: query wrappers, boilerplate"
key_files:
  created: []
  modified: []
patterns_established: []
duration_minutes: ~
bootstrap: true
hold_reason: "Unfreezes after Phase 1–3 Stable."
---

# Stage 6 Tasks — UI, Tooling & Quality

**Phase:** 6
**Status:** Hold
**Strategic Goal:** Developer-experience surface — UI, CLI, build, diagnostics, AI assistant boundary, codegen, errors.

## High-Level Checklist

- [ ] [T-6A] Definition System (declarative JSON layer). ([l1-definition-system.md](../specifications/l1-definition-system.md))
- [ ] [T-6B] Window System + multi-window + platform abstraction. ([l1-window-system.md](../specifications/l1-window-system.md))
- [ ] [T-6C] Diagnostics + profiling overlay + gizmos. ([l1-diagnostic-system.md](../specifications/l1-diagnostic-system.md))
- [ ] [T-6D] UI layout, interaction, widgets, styling. ([l1-ui-system.md](../specifications/l1-ui-system.md))
- [ ] [T-6E] Build tooling: CI pipeline, golden tests, benchmarks, release docs. ([l1-build-tooling.md](../specifications/l1-build-tooling.md))
- [ ] [T-6F] CLI tooling: scaffolding, asset commands, engine routines. ([l1-cli-tooling.md](../specifications/l1-cli-tooling.md))
- [ ] [T-6G] Platform tier matrix + capabilities + build tags. ([l1-platform-system.md](../specifications/l1-platform-system.md))
- [ ] [T-6H] AI assistant plugin protocol. ([l1-ai-assistant-system.md](../specifications/l1-ai-assistant-system.md))
- [ ] [T-6I] Examples framework + lifecycle conventions. ([l1-examples-framework.md](../specifications/l1-examples-framework.md))
- [ ] [T-6J] Compatibility policy: engine versioning + Go toolchain matrix. ([l1-compatibility-policy.md](../specifications/l1-compatibility-policy.md))
- [ ] [T-6K] Error core: E-series codes + localization + severity. ([l1-error-core.md](../specifications/l1-error-core.md))
- [ ] [T-6L] Benchmark spec: standardized perf tests + comparisons. ([l2-benchmark-spec.md](../specifications/l2-benchmark-spec.md))
- [ ] [T-6M] Codegen tools: query wrappers + boilerplate. ([l2-codegen-tools.md](../specifications/l2-codegen-tools.md))
- [ ] [T-6T] Validation: CLI integration tests, codegen golden output, benchmark regression CI gate.
