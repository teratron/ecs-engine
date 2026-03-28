# Client-Side Prediction

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

Client-Side Prediction allows the local player to see the results of their input immediately, without waiting for a server round-trip. The client applies inputs locally as if it were authoritative, then reconciles with the server's confirmed state when it arrives. If the prediction was wrong (misprediction), the client rolls back to the server state and resimulates forward using its buffered inputs. This model is the standard for fast-paced multiplayer games — shooters, platformers, racing — where input responsiveness is paramount.

## Related Specifications

- [networking-system.md](networking-system.md) — SnapshotManager (§4.2) for rollback snapshots, InputBuffer (§4.3) for input history, RollbackCoordinator (§4.4) reference implementation, DeterministicSchedule (§4.1) for resimulation
- [replication.md](replication.md) — Server state delivery to clients, NetworkAuthority (§4.10) for Predicted entities
- [transport.md](transport.md) — Input sent via ChannelID 1 (ReliableUnordered), state via ChannelID 0 (Unreliable)
- [snapshot-interpolation.md](snapshot-interpolation.md) — Remote (non-predicted) entities use interpolation for display
- [time-system.md](time-system.md) — FixedTime drives deterministic simulation ticks
- [input-system.md](input-system.md) — Raw input captured and serialized per tick
- [command-system.md](command-system.md) — Server corrections applied via deferred commands

## 1. Motivation

Without client prediction, every player action takes one full round-trip before the player sees the result. At 80ms RTT, the game feels sluggish. At 150ms, it feels broken. Players expect their character to move the instant they press a key.

Client prediction solves this by running the simulation locally for the controlled entity. The server remains authoritative — if the client predicted wrong (e.g., the server rejected a movement due to a wall the client didn't know about), the client corrects by snapping or smoothly blending to the server state. The key insight: mispredictions are rare in practice (client and server run the same simulation code), so the visual result is almost always correct and responsive.

## 2. Constraints & Assumptions

- Only entities with `NetworkAuthority::Predicted(local_connection_id)` are predicted. All other entities are displayed via snapshot interpolation.
- Prediction requires **deterministic simulation** for the predicted systems. Non-deterministic systems (particles, audio) must not be rolled back.
- The client must buffer its inputs for the last N ticks to replay them during resimulation. Buffer depth = max expected round-trip in ticks.
- The server sends authoritative state tagged with the tick number at which it was produced. The client matches this to its prediction history.
- Misprediction correction must be visually smooth — no teleporting. A blending strategy is required.

## 3. Core Invariants

- **INV-1**: The server is always authoritative. Client prediction is a visual optimization — the server's word is final.
- **INV-2**: Predicted entities are simulated locally using the same systems that run on the server (deterministic code path).
- **INV-3**: When server state arrives, the client compares it against its prediction for the same tick. If they match, no correction is needed. If they differ, the client must reconcile.
- **INV-4**: Rollback-resimulation must reproduce identical results to live simulation given the same inputs (depends on networking-system.md INV-2 and INV-4).
- **INV-5**: Non-predicted entities are never rolled back. They use snapshot interpolation only.

## 4. Detailed Design

### 4.1 Prediction Loop

Each client tick for predicted entities:

```plaintext
PredictionSystem (runs in FixedUpdate)
  1. Capture local input for this tick.
  2. Serialize input → InputBuffer.RecordInput(current_tick, local_player, input).
  3. Send input to server via transport (ChannelID 1, ReliableUnordered).
  4. Apply input to predicted entities locally (run deterministic systems).
  5. Save predicted state: PredictionHistory.Record(current_tick, predicted_state).
  6. Advance current_tick.
```

The client runs ahead of the server by approximately `RTT / 2 / fixed_timestep` ticks. This means the client is always simulating a "future" that the server hasn't confirmed yet.

### 4.2 Prediction History

The client stores its predicted state for each tick to compare against server confirmations:

```plaintext
PredictionHistory (resource)
  ring: RingBuffer[PredictionEntry]
  capacity: int    // default: 64 entries (~1 second at 60 Hz)

PredictionEntry
  tick:       uint64
  entities:   map[EntityID] -> PredictedSnapshot

PredictedSnapshot
  components: map[ComponentID] -> []byte   // serialized predicted component values
  checksum:   uint32                       // for fast mismatch detection
```

### 4.3 Server Reconciliation

When the server's authoritative state arrives for a tick:

```plaintext
ReconciliationSystem (runs in PreUpdate, after network receive)
  1. Receive server state for tick T (via replication pipeline).
  2. Look up PredictionHistory entry for tick T.
  3. Compare:
     a. Compute checksum of server state for predicted entities.
     b. Compare against PredictionHistory[T].checksum.
     c. If match → prediction was correct. Discard history entries <= T. Done.
     d. If mismatch → misprediction detected. Proceed to rollback.

  4. Rollback:
     a. Restore predicted entities to server state at tick T.
        (Apply server component values via commands.)
     b. Retrieve buffered inputs for ticks T+1 through current_tick.
     c. Resimulate: for each tick from T+1 to current_tick:
        - Apply buffered input for that tick.
        - Run deterministic systems.
        - Save new PredictionHistory entry (overwriting old prediction).
     d. The final state after resimulation becomes the new predicted state.

  5. Discard PredictionHistory entries older than T.
  6. Update confirmed_tick = T.
```

### 4.4 Misprediction Smoothing

Raw rollback correction causes visual teleporting. The system smooths corrections over multiple frames:

```plaintext
CorrectionState (component, added to predicted entities during misprediction)
  visual_offset:    Vec3     // difference between corrected position and displayed position
  rotation_offset:  Quat     // difference in rotation
  blend_remaining:  Duration // time left to blend out the offset
  blend_duration:   Duration // total blend time (default: 100ms)

On misprediction:
  1. Before applying server state, record the visual position/rotation.
  2. After rollback+resimulate, compute difference:
     visual_offset = old_display_position - new_predicted_position
     rotation_offset = old_display_rotation * inverse(new_predicted_rotation)
  3. Set blend_remaining = blend_duration.

Each render frame:
  t = 1.0 - (blend_remaining / blend_duration)
  display_position = predicted_position + visual_offset * (1.0 - t)
  display_rotation = Slerp(rotation_offset * predicted_rotation, predicted_rotation, t)
  blend_remaining -= frame_delta
  If blend_remaining <= 0 → remove CorrectionState component.
```

```plaintext
PredictionConfig (resource)
  max_rollback_ticks:   uint8      // maximum ticks to resimulate, default 16
  correction_blend_ms:  Duration   // misprediction smoothing duration, default 100ms
  checksum_enabled:     bool       // fast-path mismatch detection, default true
  prediction_systems:   []SystemID // which systems to re-run during resimulation
```

### 4.5 Input Delay (Optional)

Some games add a small input delay (1-2 ticks) to reduce misprediction frequency:

```plaintext
InputDelayConfig (resource)
  delay_ticks: uint8    // default: 0 (no delay)

If delay_ticks > 0:
  Input captured at tick T is applied at tick T + delay_ticks.
  This gives the server more time to confirm, reducing rollback frequency.
  Trade-off: adds delay_ticks * fixed_timestep of input latency.
```

Input delay is most useful for fighting games and high-precision competitive games where misprediction artifacts are more disruptive than a consistent 1-2 frame delay.

### 4.6 Predicted vs Interpolated Entities

A single client renders entities using two different strategies simultaneously:

```plaintext
Local player entity:
  NetworkAuthority::Predicted(local_connection_id)
  → Rendered from client prediction (this spec)
  → Misprediction smoothing via CorrectionState

Remote player entities:
  NetworkAuthority::Server or NetworkAuthority::Predicted(other_connection_id)
  → Rendered via snapshot interpolation (snapshot-interpolation.md)
  → No local simulation, no rollback

Server-only entities (AI, projectiles):
  NetworkAuthority::Server
  → Rendered via snapshot interpolation
```

This means the client displays the local player ~RTT/2 in the "future" relative to remote entities. This time gap is inherent to client prediction and is why shooting mechanics need lag compensation (a game-layer concern, not an engine concern).

### 4.7 ClientPredictionPlugin

```plaintext
ClientPredictionPlugin
  config: PredictionConfig

Build(app):
  app.InsertResource(PredictionConfig{...})
  app.InsertResource(InputDelayConfig{ delay_ticks: 0 })
  app.InsertResource(PredictionHistory{ capacity: 64 })
  app.AddSystem(FixedUpdate, PredictionSystem)
  app.AddSystem(PreUpdate, ReconciliationSystem)
  app.AddSystem(PostUpdate, CorrectionSmoothingSystem)
```

**Dependency**: Requires `ReplicationPlugin`, `NetworkPlugin`, and `SnapshotInterpolationPlugin` (for rendering remote entities).

## 5. Open Questions

- Should the engine provide built-in **lag compensation** (rewinding the server world to the client's visual time for hit detection), or is that a game-layer concern?
- How should **entity spawning during rollback** be handled? If the server spawns an entity at tick T, and the client rolls back past T, should the entity be despawned and respawned during resimulation?
- Should prediction extend to **physics** simulation? Rolling back physics is complex (contacts, constraints). The physics solver may need special rollback support.
- How should **variable tick rate** clients handle prediction? If a client runs at 30 Hz but the server at 60 Hz, inputs map to different ticks.
- Should `PredictionHistory` store per-component deltas or full snapshots? Deltas save memory but complicate rollback.

## 6. Implementation Notes

1. `PredictionHistory` ring buffer — core data structure, testable in isolation.
2. `PredictionSystem` with local input → local simulation — works without server.
3. `ReconciliationSystem` with simple snap correction — basic server reconciliation, no smoothing.
4. `CorrectionState` blending — visual polish, add after basic reconciliation works.
5. Input delay support — optional, add for games that request it.
6. Integration test with `LoopbackBackend`: server + client in one process, inject artificial mispredictions.

## 7. Drawbacks & Alternatives

**Complexity**: Client prediction is the most complex networking model. It requires deterministic simulation, input buffering, state comparison, rollback, and resimulation — all running every frame. The alternative (snapshot interpolation only) is simpler but adds input latency.

**Misprediction artifacts**: Even with smoothing, severe mispredictions (server rejects a large movement) produce visible corrections. The only mitigations are reducing misprediction frequency (better prediction logic, input delay) and increasing blend duration (masking the snap).

**CPU cost of resimulation**: Rolling back 8 ticks means running 8 ticks of deterministic simulation in one frame. For complex games, this can cause frame spikes. Mitigation: limit `max_rollback_ticks`, and profile resimulation separately from normal simulation.

**Determinism requirement**: Predicted systems must be deterministic. This constrains what systems can participate in prediction. Floating-point non-determinism across platforms can cause persistent mispredictions. Mitigation: use fixed-point math for predicted systems, or accept that cross-platform determinism is best-effort.

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-28 | Initial draft — prediction loop, reconciliation, misprediction smoothing, input delay |
| — | — | Planned examples: `examples/networking/` |
