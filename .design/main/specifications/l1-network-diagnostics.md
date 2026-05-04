# Network Diagnostics

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

Network Diagnostics extends the engine's diagnostic system with networking-specific metrics, visualizations, and alerts. It collects per-connection statistics (RTT, packet loss, bandwidth), per-system replication costs, and synchronization health indicators (desync events, rollback frequency, buffer fill levels). All metrics feed into the standard `DiagnosticsStore` and are renderable via the debug overlay. Network diagnostics are zero-cost when disabled and compile-time removable in release builds.

## Related Specifications

- [diagnostic-system.md](l1-diagnostic-system.md) — DiagnosticsStore, Diagnostic type, debug overlay infrastructure
- [transport.md](l1-transport.md) — ConnectionStats (§4.7) provides per-connection RTT, packet loss, bandwidth
- [replication.md](l1-replication.md) — ReplicationStats (§4.9) tracks replicated entity count, bytes, deferred updates
- [networking-system.md](l1-networking-system.md) — DesyncDetector (§4.7) fires DesyncDetected events
- [client-prediction.md](l1-client-prediction.md) — Misprediction frequency and rollback depth
- [snapshot-interpolation.md](l1-snapshot-interpolation.md) — Buffer fill level, interpolation quality
- [lockstep.md](l1-lockstep.md) — Stall frequency, input delay stats

## 1. Motivation

Network bugs are the hardest to reproduce and diagnose. A player reports "the game felt laggy" — but was it:

- High RTT (network distance)?
- Packet loss (unreliable connection)?
- Bandwidth saturation (too many entities replicated)?
- Frequent mispredictions (simulation divergence)?
- Snapshot buffer starvation (jitter spikes)?
- A desync that went undetected?

Without first-party network diagnostics, developers add ad-hoc `fmt.Println` statements, miss the real cause, and ship broken netcode. A structured diagnostic layer makes every networking issue visible, measurable, and traceable.

## 2. Constraints & Assumptions

- All network diagnostics are registered in the standard `DiagnosticsStore` (diagnostic-system.md §4.1). No separate metric store.
- Diagnostic collection is opt-in per category. Disabled categories have zero overhead.
- Debug overlay rendering depends on the render pipeline. Headless servers expose metrics via a queryable resource (no rendering).
- All diagnostic paths use `DiagnosticPath` strings prefixed with `"net/"` for namespace isolation.
- Network diagnostics never modify game state. They read transport stats, replication stats, and event counts — pure observation.

## 3. Core Invariants

- **INV-1**: Network diagnostics have zero overhead when no readers are registered (inherits from diagnostic-system.md INV-1).
- **INV-2**: Diagnostic collection never modifies game state, transport state, or replication state.
- **INV-3**: All network metrics use the standard `DiagnosticsStore` API. No custom metric infrastructure.
- **INV-4**: Alert thresholds are configurable. No hardcoded magic numbers for warning/critical states.

## 4. Detailed Design

### 4.1 Metric Categories

```plaintext
Network diagnostics are organized into categories, each with a dedicated collection system:

Category: Connection
  Path prefix: "net/connection/"
  Source: transport.md ConnectionStats
  Metrics:
    net/connection/rtt              Duration   // round-trip time (EWMA)
    net/connection/rtt_variance     Duration   // jitter indicator
    net/connection/packet_loss      float32    // fraction 0..1, last 128 packets
    net/connection/send_rate        float32    // bytes/sec outbound
    net/connection/receive_rate     float32    // bytes/sec inbound
    net/connection/send_queue       int        // outbound packets queued
    net/connection/peer_count       int        // active connections

Category: Replication
  Path prefix: "net/replication/"
  Source: replication.md ReplicationStats
  Metrics:
    net/replication/entities        int        // entities currently replicated
    net/replication/bytes_sent      uint64     // replication bytes this frame
    net/replication/bytes_received  uint64     // replication bytes received
    net/replication/updates_deferred int       // updates that didn't fit in bandwidth
    net/replication/entity_map_size int        // entity mappings active

Category: Prediction
  Path prefix: "net/prediction/"
  Source: client-prediction.md PredictionHistory
  Metrics:
    net/prediction/mispredictions   int        // mispredictions this second
    net/prediction/rollback_depth   float32    // average ticks rolled back per misprediction
    net/prediction/corrections      int        // active CorrectionState blends
    net/prediction/input_rtt        Duration   // time from input send to server confirmation

Category: Interpolation
  Path prefix: "net/interpolation/"
  Source: snapshot-interpolation.md SnapshotBuffer
  Metrics:
    net/interpolation/buffer_fill   int        // snapshots buffered ahead of render_time
    net/interpolation/render_delay  Duration   // current adaptive render delay
    net/interpolation/extrapolating bool       // true if extrapolation is active
    net/interpolation/buffer_starved int       // BufferStarved events this second

Category: Lockstep
  Path prefix: "net/lockstep/"
  Source: lockstep.md LockstepScheduler
  Metrics:
    net/lockstep/input_delay        uint8      // current input delay in ticks
    net/lockstep/stall_count        int        // stalls waiting for peer input this second
    net/lockstep/stall_duration     Duration   // total stall time this second
    net/lockstep/speculative_ticks  int        // ticks run speculatively (if enabled)
    net/lockstep/desync_events      int        // desync detections this session

Category: RPC
  Path prefix: "net/rpc/"
  Source: rpc.md RpcRegistry
  Metrics:
    net/rpc/sent                    int        // RPCs sent this second
    net/rpc/received                int        // RPCs received this second
    net/rpc/dropped                 int        // RPCs dropped (unknown type or rate limited)
    net/rpc/bytes                   uint64     // total RPC bytes this second
```

### 4.2 Collection Systems

Each category has a lightweight collection system that reads source stats and pushes to `DiagnosticsStore`:

```plaintext
NetworkDiagnosticsSystem (runs in Last schedule, after all gameplay)
  For each enabled category:
    Read source stats (ConnectionStats, ReplicationStats, etc.)
    Push values to DiagnosticsStore via Diagnostic.PushValue()

  Collection frequency: every frame (60 Hz default).
  History depth: inherits from DiagnosticsStore (default 120 entries = 2 seconds).
```

**Server-side aggregation**: On the server, connection metrics are per-client. The collection system reports:

- Per-connection metrics tagged with ConnectionID (for per-client debugging)
- Aggregate metrics (average RTT across all clients, total bandwidth, worst-case packet loss)

```plaintext
Server-specific metrics:
  net/server/total_bandwidth      float32    // sum of all client send_rate
  net/server/worst_rtt            Duration   // highest RTT across all clients
  net/server/worst_packet_loss    float32    // highest loss across all clients
  net/server/client_count         int        // active connections
```

### 4.3 Alert System

Configurable thresholds trigger alerts when metrics exceed acceptable ranges:

```plaintext
NetworkAlertConfig (resource)
  alerts: []AlertRule

AlertRule
  metric_path:   DiagnosticPath    // e.g., "net/connection/packet_loss"
  warning:       float64           // threshold for Warning level
  critical:      float64           // threshold for Critical level
  window:        int               // number of samples to average over (default: 60)
  cooldown:      Duration          // min time between repeated alerts (default: 5s)

AlertLevel:
  Normal
  Warning
  Critical

Default alert rules:
  net/connection/rtt            warning: 100ms   critical: 200ms
  net/connection/packet_loss    warning: 0.05    critical: 0.15
  net/prediction/mispredictions warning: 5/sec   critical: 15/sec
  net/interpolation/buffer_fill warning: < 2     critical: < 1
  net/lockstep/stall_duration   warning: 100ms   critical: 500ms
```

When a threshold is crossed:

```plaintext
fire_event(NetworkAlert {
  metric:  DiagnosticPath,
  level:   AlertLevel,
  value:   float64,
  message: string,    // e.g., "Packet loss 12% exceeds warning threshold (5%)"
})
```

Game code can react to alerts (show UI indicator, log to analytics, trigger quality adaptation).

### 4.4 Debug Overlay

Network diagnostics render in the standard debug overlay (diagnostic-system.md) when enabled:

```plaintext
NetworkOverlayConfig (resource)
  enabled:      bool        // default: false (toggle via keybind or console)
  position:     OverlayPosition  // TopRight (default), TopLeft, BottomRight, etc.
  categories:   []string    // which categories to show, default: all enabled
  graph_height: int         // pixels for history graph, default: 60
  show_graphs:  bool        // show rolling history graphs, default: true

Overlay layout:
  ┌─────────────────────────────┐
  │ NET DIAGNOSTICS             │
  │ RTT:    42ms ▸ [graph]      │
  │ Loss:   1.2% ▸ [graph]      │
  │ BW Out: 128 KB/s            │
  │ BW In:  64 KB/s             │
  │ Entities: 247               │
  │ Mispredictions: 0/s         │
  │ Buffer: 3/32 snapshots      │
  │ ⚠ Packet loss > 5%         │
  └─────────────────────────────┘
```

The overlay uses the gizmo/immediate-mode drawing infrastructure from diagnostic-system.md. Alerts are highlighted with color coding (yellow = warning, red = critical).

### 4.5 Network Profiling Spans

For deep performance analysis, the system emits tracing spans around networking operations:

```plaintext
Spans (compile-time removable via build tag "netprofile"):
  "net.transport.send"       — time spent serializing and enqueuing outbound packets
  "net.transport.recv"       — time spent processing inbound packets
  "net.replication.send"     — time spent in ReplicationSendSystem
  "net.replication.recv"     — time spent in ReplicationReceiveSystem
  "net.prediction.rollback"  — time spent in rollback-resimulate cycle
  "net.prediction.reconcile" — time spent comparing prediction vs server state
  "net.interpolation.lerp"   — time spent interpolating snapshot buffer
  "net.rpc.dispatch"         — time spent deserializing and dispatching RPCs
```

These integrate with the profiling protocol (profiling-protocol.md) and are visible in external profilers (Tracy, pprof).

### 4.6 Desync Report

When a `DesyncDetected` event fires, the diagnostic system captures a detailed report:

```plaintext
DesyncReport (generated on DesyncDetected event)
  tick:             uint64
  local_checksum:   uint32
  remote_checksum:  uint32
  peer:             ConnectionID
  local_snapshot:   []byte    // optional: full state dump for offline comparison
  timestamp:        Duration

Stored in:
  DesyncHistory (resource)
    reports: RingBuffer[DesyncReport]
    capacity: 16
    dump_to_disk: bool   // default: true in debug builds
    dump_path: string    // default: "logs/desync/"
```

Disk dumps are written as binary files named `desync_tick_{N}.bin` and can be loaded by a companion diff tool to identify exactly which entities/components diverged.

### 4.7 NetworkDiagnosticsPlugin

```plaintext
NetworkDiagnosticsPlugin
  config: NetworkAlertConfig

Build(app):
  app.InsertResource(NetworkAlertConfig{ /* default rules */ })
  app.InsertResource(NetworkOverlayConfig{ enabled: false })
  app.InsertResource(DesyncHistory{ capacity: 16 })
  app.AddEvent[NetworkAlert]()
  app.AddSystem(Last, NetworkDiagnosticsSystem)
  app.AddSystem(Last, NetworkAlertSystem)
  app.AddSystem(Last, DesyncReportSystem)
  // Overlay rendering added conditionally if render pipeline is available

  // Register diagnostic paths in DiagnosticsStore
  diagnostics.Register("net/connection/rtt", "ms")
  diagnostics.Register("net/connection/packet_loss", "%")
  diagnostics.Register("net/replication/entities", "count")
  // ... etc
```

**Dependency**: Requires `DiagnosticsPlugin` (diagnostic-system.md). Optionally reads from any networking plugin that is active (transport, replication, prediction, interpolation, lockstep, rpc). Missing sources are silently skipped — the diagnostic system gracefully degrades.

## 5. Open Questions

- Should network diagnostics be exportable to external monitoring systems (Prometheus, Grafana) for dedicated servers?
- Should the debug overlay support per-entity replication cost visualization (color-code entities by bandwidth consumed)?
- Should the desync diff tool be a first-party engine utility, or is the binary dump format sufficient for third-party tooling?
- How should headless servers expose diagnostics — HTTP endpoint, console commands, or structured log output?

## 6. Implementation Notes

1. `NetworkDiagnosticsSystem` reading from `ConnectionStats` — simplest end-to-end: transport → DiagnosticsStore → overlay.
2. Alert system with default thresholds — configurable, fires events.
3. Prediction and interpolation metric collection — add as those plugins are implemented.
4. Desync report capture and disk dump — essential debugging tool for lockstep.
5. Profiling spans — add last, behind build tag.

## 7. Drawbacks & Alternatives

**Dependency on DiagnosticsStore**: All network metrics go through the generic diagnostic system. The alternative — a dedicated network metric store with specialized queries — would be more powerful but fragments the diagnostic surface. Using the standard store keeps one API for all diagnostics (frame time, entity count, network) and one overlay renderer.

**Per-frame collection overhead**: Collecting stats every frame (even when the overlay is hidden) has a small cost. Mitigation: categories are individually toggleable, and the collection system short-circuits when no readers are registered (INV-1).

**Debug-only overlay**: The visual overlay requires the render pipeline. Headless servers (dedicated game servers) cannot render it. For servers, metrics are exposed as resources queryable by admin tools or exported via a future monitoring endpoint.

## Canonical References

<!-- MANDATORY for Stable status. List authoritative source files that downstream agents
     MUST read before implementing this spec. Use relative paths from project root.
     Stub state — fill with concrete files when implementation begins (Phase 1+). -->

| Alias | Path | Purpose |
| :--- | :--- | :--- |

<!-- Empty table = no canonical sources yet. Populate one row per authoritative file
     when implementation lands (Phase 1+). Stable promotion requires ≥1 row. -->

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-28 | Initial draft — metric categories, alerts, debug overlay, profiling spans, desync reports |
| — | — | Planned examples: `examples/networking/` |
