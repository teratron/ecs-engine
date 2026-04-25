---
phase: 7
name: "Networking & Hot-Reload"
status: Hold
subsystem: "pkg/net, pkg/profiling, pkg/hotreload"
requires:
  - "Phase 2 App Framework Stable"
  - "Phase 1 Scheduler Stable"
provides:
  - "Profiling protocol (Tracy + pprof bridges)"
  - "Multiplayer boundaries: snapshot/rollback primitives"
  - "UDP transport: channels, reliability, MTU"
  - "State replication: markers, mapping, visibility, deltas"
  - "Sync models: snapshot interpolation, client prediction, lockstep"
  - "Typed RPC + rate limiting"
  - "Network diagnostics: metrics, alerts, debug overlay, desync reports"
  - "Hot-reload orchestrator: code restart + shader hot-swap"
key_files:
  created: []
  modified: []
patterns_established: []
duration_minutes: ~
bootstrap: true
hold_reason: "Unfreezes after Phase 2 App Framework + Phase 1 Scheduler Stable."
---

# Stage 7 Tasks — Networking & Hot-Reload

**Phase:** 7
**Status:** Hold

## High-Level Checklist

- [ ] [T-7A] Profiling protocol: Tracy integration, custom spans, pprof mapping. ([l1-profiling-protocol.md](../specifications/l1-profiling-protocol.md))
- [ ] [T-7B] Networking system: snapshot/rollback primitives, fixed-step sync. ([l1-networking-system.md](../specifications/l1-networking-system.md))
- [ ] [T-7C] UDP transport: channels, reliability, lifecycle, MTU. ([l1-transport.md](../specifications/l1-transport.md))
- [ ] [T-7D] Replication: markers, entity mapping, visibility, deltas, priority. ([l1-replication.md](../specifications/l1-replication.md))
- [ ] [T-7E] Snapshot interpolation: server snapshots + client buffer + adaptive delay. ([l1-snapshot-interpolation.md](../specifications/l1-snapshot-interpolation.md))
- [ ] [T-7F] Client prediction: input prediction, reconciliation, rollback smoothing. ([l1-client-prediction.md](../specifications/l1-client-prediction.md))
- [ ] [T-7G] Lockstep: deterministic, input delay, speculative exec, desync detect. ([l1-lockstep.md](../specifications/l1-lockstep.md))
- [ ] [T-7H] RPC: typed send/receive, event integration, rate limiting. ([l1-rpc.md](../specifications/l1-rpc.md))
- [ ] [T-7I] Network diagnostics: metrics, alerts, overlay, desync reports. ([l1-network-diagnostics.md](../specifications/l1-network-diagnostics.md))
- [ ] [T-7J] Hot reload: code restart with state snapshot, shader hot-swap, orchestrator. ([l1-hot-reload.md](../specifications/l1-hot-reload.md))
- [ ] [T-7T] Validation: lossy-network sim suite, deterministic lockstep checksum, rollback fuzz.
