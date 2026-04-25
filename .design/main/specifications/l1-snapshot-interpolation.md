# Snapshot Interpolation

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

Snapshot Interpolation is a synchronization model where the server periodically sends authoritative World state snapshots to clients. Clients render a slightly delayed version of the game, smoothly interpolating between the two most recent snapshots. This model trades input responsiveness for visual smoothness and simplicity — the client never predicts, never rolls back, and never simulates. It is ideal for spectator views, slow-paced games, and as the fallback rendering mode for other netcode models.

## Related Specifications

- [networking-system.md](l1-networking-system.md) — SnapshotManager (§4.2) provides snapshot capture and delta compression
- [replication.md](l1-replication.md) — Replication pipeline delivers per-entity component updates to clients
- [transport.md](l1-transport.md) — Snapshots sent via ChannelID 0/3 (Unreliable) for state, ChannelID 1 (Reliable) for spawns/despawns
- [time-system.md](l1-time-system.md) — TimeVirtual and TimeFixed drive interpolation timing
- [change-detection.md](l1-change-detection.md) — Server uses Changed[T] to build delta snapshots
- [math-system.md](l1-math-system.md) — Lerp/Slerp operations for interpolation

## 1. Motivation

Client-side prediction (see [client-prediction.md](l1-client-prediction.md)) is complex and error-prone. Many game types do not need it:

- **Spectator mode**: Viewers watch a match with no input — pure state consumption.
- **Slow-paced games**: Turn-based, strategy, simulation — 100ms of visual delay is imperceptible.
- **Non-player entities**: Even in fast games, entities the local player does not control (other players, NPCs, projectiles) are best displayed via interpolation, not prediction.

Snapshot interpolation provides a simple, robust rendering pipeline: receive state → buffer → interpolate → display. No rollback, no misprediction artifacts, no client-side simulation of remote entities.

## 2. Constraints & Assumptions

- The client does **not** simulate the game. It receives state and displays it.
- Interpolation introduces deliberate visual delay (typically 2× the snapshot interval, ~66ms at 30 Hz snapshots).
- The server is authoritative. Clients have no means to alter the simulation.
- Snapshots arrive over Unreliable channels — some may be lost or arrive out of order. The interpolation buffer handles gaps gracefully.
- Extrapolation (predicting forward past the latest snapshot) is opt-in and limited to prevent visual artifacts.

## 3. Core Invariants

- **INV-1**: The client display always shows a state interpolated between two confirmed server snapshots. It never displays a predicted or locally simulated state.
- **INV-2**: Snapshot buffer holds at least 2 snapshots before interpolation begins. Until then, the client displays the latest received state without interpolation.
- **INV-3**: Out-of-order or duplicate snapshots are silently discarded based on tick number.
- **INV-4**: Interpolation factor `t` is always clamped to [0.0, 1.0]. Extrapolation uses a separate, bounded factor.

## 4. Detailed Design

### 4.1 Snapshot Buffer

Each client maintains a time-sorted buffer of received snapshots:

```plaintext
SnapshotBuffer (resource)
  buffer:       SortedRing[SnapshotEntry]   // sorted by tick, ring of configurable size
  capacity:     int                         // default: 32 entries
  render_delay: Duration                    // deliberate delay, default: 100ms

SnapshotEntry
  tick:         uint64
  timestamp:    Duration     // server-side time of this tick
  entities:     map[EntityID] -> ComponentSnapshot
  received_at:  Duration     // local time when this snapshot arrived

ComponentSnapshot
  component_id: ComponentID
  data:         []byte       // deserialized component data (via type registry)
```

When a new snapshot arrives:
```plaintext
1. If tick <= latest buffered tick → discard (out-of-order or duplicate).
2. Insert into sorted position.
3. If buffer is full → evict oldest entry.
```

### 4.2 Interpolation Timing

The client renders at a point in time deliberately behind the server:

```plaintext
render_time = latest_server_time - render_delay

For each frame:
  1. Find snapshot_a and snapshot_b in buffer such that:
     snapshot_a.timestamp <= render_time < snapshot_b.timestamp
  2. Compute interpolation factor:
     t = (render_time - snapshot_a.timestamp) / (snapshot_b.timestamp - snapshot_a.timestamp)
  3. For each entity present in both snapshots:
     interpolated_state = Lerp(snapshot_a[entity], snapshot_b[entity], t)
```

**Adaptive render delay**: If the buffer is consistently underfull (network jitter causing gaps), the system increases `render_delay` automatically. If the buffer is consistently overfull, it decreases. This keeps the buffer at its target fill level (default: 3 snapshots ahead of render_time).

```plaintext
InterpolationConfig (resource)
  render_delay:        Duration    // current delay, default 100ms
  min_render_delay:    Duration    // floor, default 33ms (1 snapshot at 30 Hz)
  max_render_delay:    Duration    // ceiling, default 500ms
  target_buffer_size:  int         // ideal snapshots ahead of render_time, default 3
  adjustment_speed:    float32     // how fast delay adapts, default 0.1
```

### 4.3 Per-Component Interpolation

Not all components interpolate the same way. The system dispatches to per-type interpolation functions:

```plaintext
InterpolationFn: fn(a: []byte, b: []byte, t: float32) -> []byte

Registered per ComponentID in InterpolationRegistry (resource):
  map[ComponentID] -> InterpolationFn

Built-in interpolation functions:
  Transform     → Lerp position, Slerp rotation, Lerp scale
  Velocity      → Lerp
  Color         → Lerp in linear space
  Health        → Snap (no interpolation — discrete value)
  Name          → Snap (use latest)
```

**Snap vs Lerp**: Components registered with `Snap` interpolation simply use the value from `snapshot_b` when `t >= 0.5`, or `snapshot_a` otherwise. This is the default for non-numeric components.

If no InterpolationFn is registered for a component → default to Snap.

### 4.4 Entity Spawn and Despawn

Entities that appear or disappear between snapshots need special handling:

```plaintext
Entity in snapshot_b but NOT in snapshot_a (spawn):
  → Display entity at snapshot_b state immediately (no lerp source).
  → Optionally play a spawn effect/animation.
  → fire_event(InterpolatedEntitySpawned { entity, tick: snapshot_b.tick })

Entity in snapshot_a but NOT in snapshot_b (despawn):
  → Continue displaying at snapshot_a state until render_time passes snapshot_b.timestamp.
  → Then remove the entity.
  → Optionally play a despawn effect.
  → fire_event(InterpolatedEntityDespawned { entity, tick: snapshot_b.tick })
```

### 4.5 Extrapolation (Opt-In)

When the buffer runs dry (no snapshot_b ahead of render_time), the system can extrapolate:

```plaintext
ExtrapolationConfig (resource)
  enabled:      bool        // default: false
  max_duration: Duration    // maximum extrapolation time, default 200ms
  method:       ExtrapolationMethod

ExtrapolationMethod:
  LastValue     — hold last known state (freeze)
  LinearPredict — extend last velocity: pos += velocity * dt
  DeadReckoning — use last velocity + acceleration

When extrapolating:
  Display render_time = latest_snapshot.timestamp + extrapolation_dt
  extrapolation_dt clamped to max_duration.
  If max_duration exceeded → freeze at last state, emit BufferStarved event.
```

Extrapolation is inherently speculative and can produce visual artifacts (entities passing through walls, overshooting turns). The default is off — prefer increasing `render_delay` instead.

### 4.6 Jitter Buffer

Network jitter causes snapshots to arrive at irregular intervals. The jitter buffer smooths this:

```plaintext
Jitter buffer behavior is implicit in the SnapshotBuffer + adaptive render_delay:
  - render_delay acts as the jitter absorption window.
  - If a snapshot arrives late but within render_delay → no visual disruption.
  - If a snapshot is lost entirely → interpolation stretches between adjacent snapshots
    (visual stutter but no discontinuity).
  - If two snapshots arrive in the same frame → both are buffered normally.
```

### 4.7 SnapshotInterpolationPlugin

```plaintext
SnapshotInterpolationPlugin
  config: InterpolationConfig

Build(app):
  app.InsertResource(InterpolationConfig{...})
  app.InsertResource(ExtrapolationConfig{ enabled: false })
  app.InsertResource(SnapshotBuffer{ capacity: 32 })
  app.InsertResource(InterpolationRegistry{})
  app.AddEvent[InterpolatedEntitySpawned]()
  app.AddEvent[InterpolatedEntityDespawned]()
  app.AddEvent[BufferStarved]()
  app.AddSystem(PreUpdate, SnapshotReceiveSystem)   // buffer incoming snapshots
  app.AddSystem(Update, InterpolationSystem)         // compute interpolated state
  app.AddSystem(Update, SpawnDespawnSystem)           // handle entity transitions

  // Register default interpolation functions
  registry.Register[Transform](LerpTransform)
  registry.Register[Velocity](LerpVec3)
```

**Dependency**: Requires `ReplicationPlugin` (for receiving entity state) and `NetworkPlugin` (for transport).

## 5. Open Questions

- Should the server send snapshots at a configurable rate independent of simulation tick rate (e.g., simulate at 60 Hz but snapshot at 20 Hz)?
- Should interpolation support cubic splines (Catmull-Rom) for smoother curves with fewer snapshots, or is linear sufficient?
- How should hierarchical entities (parent-child) be interpolated — independently per entity, or propagate parent interpolation to children?
- Should the jitter buffer report statistics (average jitter, buffer fill level) via the diagnostic system?

## 6. Implementation Notes

1. `SnapshotBuffer` and `InterpolationConfig` first — core data structures, testable without networking.
2. `InterpolationSystem` with hardcoded Transform lerp — proof of concept with `LoopbackBackend`.
3. `InterpolationRegistry` and per-component dispatch — generalize to arbitrary components.
4. Adaptive render delay — requires observing buffer fill level over multiple frames.
5. Extrapolation — opt-in layer, add after basic interpolation is stable.

## 7. Drawbacks & Alternatives

**Deliberate delay**: The fundamental trade-off. Snapshot interpolation adds ~100ms of visual latency by design. For fast-paced competitive games, this delay is unacceptable for the local player's character — client prediction (see [client-prediction.md](l1-client-prediction.md)) is needed for the controlled entity. However, even in fast games, remote entities are typically displayed via interpolation.

**No client authority**: The client cannot affect the simulation. This is a feature for spectator/viewer modes but a limitation for interactive gameplay. The model must be combined with another (client prediction, lockstep) for the local player's input to feel responsive.

**Bandwidth**: Sending full/delta snapshots N times per second consumes more bandwidth than input-only synchronization (lockstep). Mitigated by delta compression (replication.md §4.5) and priority accumulator (replication.md §4.6).

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
| 0.1.0 | 2026-03-28 | Initial draft — snapshot buffer, interpolation timing, per-component lerp, extrapolation, jitter handling |
| — | — | Planned examples: `examples/networking/` |
