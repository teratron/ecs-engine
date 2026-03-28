# Deterministic Lockstep

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

Deterministic Lockstep is a synchronization model where all peers run the same deterministic simulation and only exchange inputs, not state. Each simulation tick advances only when inputs from all players for that tick have been received. Because the simulation is identical on all machines given the same inputs, no state replication is needed — bandwidth scales with player count, not world complexity. This model is ideal for RTS games, fighting games, and any genre where world state is too large or complex to replicate but determinism can be guaranteed.

## Related Specifications

- [networking-system.md](networking-system.md) — DeterministicSchedule (§4.1) guarantees identical execution, DeterministicRNG (§4.8) for reproducible randomness, DesyncDetector (§4.7) for checksum verification, InputBuffer (§4.3) for input serialization
- [transport.md](transport.md) — Inputs sent via ChannelID 1 (ReliableUnordered); checksums via ChannelID 1
- [time-system.md](time-system.md) — FixedTime drives simulation at a constant rate
- [input-system.md](input-system.md) — Input serialization for network transmission
- [replication.md](replication.md) — Not used in pure lockstep (no state replication); may be used for initial state sync on late join

## 1. Motivation

In an RTS with 10,000 units, replicating the position, health, and AI state of every unit 30 times per second is prohibitive. But if every client runs the same simulation with the same inputs, the state is implicitly synchronized — no replication needed.

Lockstep exploits this: the only data sent over the network is player inputs (~50-200 bytes per player per tick). A 4-player RTS with 10,000 units uses the same bandwidth as a 4-player RTS with 10 units. This makes lockstep the only practical synchronization model for games with large simulation state.

The trade-off is strict: the simulation must be 100% deterministic. A single floating-point rounding difference causes a desync that accumulates over time and corrupts the game. The engine's `DeterministicSchedule` provides the foundation; this spec builds the synchronization protocol on top.

## 2. Constraints & Assumptions

- **All** gameplay systems must be registered with `DeterministicSchedule`. Non-deterministic systems (particles, sound, UI) run separately and are not part of the lockstep simulation.
- All peers must use the same engine version, the same system registration order, and the same fixed timestep. Version mismatch is detected during connection handshake (transport.md §4.5).
- Pure lockstep stalls if any player's input is late — one slow player freezes all others. Input delay and speculative execution mitigate this.
- Late join (joining a game in progress) requires a full state snapshot transfer, since lockstep does not continuously replicate state.
- Lockstep is most natural for peer-to-peer topologies but also works with a relay server that aggregates and distributes inputs.

## 3. Core Invariants

- **INV-1**: No simulation tick advances until inputs for that tick from all connected peers are available.
- **INV-2**: Given identical inputs and identical initial state, all peers produce identical simulation state at every tick.
- **INV-3**: Inputs are the only gameplay data sent over the network. World state is never replicated during normal gameplay (only during late join).
- **INV-4**: Desync detection runs periodically. A confirmed desync halts the game and reports which tick diverged.
- **INV-5**: Input delay is configurable and bounded. The system never waits indefinitely for a peer's input — a timeout triggers disconnect.

## 4. Detailed Design

### 4.1 Lockstep Tick Loop

The core lockstep loop replaces the standard FixedUpdate:

```plaintext
LockstepScheduler (resource)
  current_tick:     uint64
  input_delay:      uint8       // ticks of deliberate input delay, default 2
  input_timeout:    Duration    // max wait for a peer's input, default 5s
  peers:            []ConnectionID

Each frame:
  1. Check if inputs from ALL peers are available for tick (current_tick + input_delay):
     - If yes → proceed to step 2.
     - If no → wait (do NOT advance simulation). Display "waiting for players" if stalled > 200ms.
  2. Run DeterministicSchedule for current_tick:
     - Apply all peer inputs for current_tick.
     - Execute all deterministic systems in declared order.
  3. Save state checksum: SnapshotManager.TakeChecksum(current_tick).
  4. Advance current_tick++.
  5. Repeat if time budget allows (catch-up for multiple ticks per frame).
```

### 4.2 Input Flow

```plaintext
Local input pipeline:
  1. Capture raw input.
  2. Serialize to SerializedInput (networking-system.md §4.3).
  3. Tag with tick = current_tick + input_delay.
  4. Store locally in InputBuffer.
  5. Broadcast to all peers via transport (ChannelID 1, ReliableUnordered).

Remote input pipeline:
  1. Receive SerializedInput from peer.
  2. Store in InputBuffer under peer's PlayerID and tagged tick.
  3. If all peers have submitted input for the next tick → simulation can advance.
```

**Input delay** (default: 2 ticks): The local player's input is scheduled for 2 ticks in the future, giving remote inputs time to arrive before the tick is due. At 60 Hz with 2-tick delay, the player experiences ~33ms of input latency — acceptable for most genres.

### 4.3 Speculative Execution (Opt-In)

Pure lockstep freezes if any input is late. Speculative execution avoids freezing by predicting missing inputs:

```plaintext
SpeculativeConfig (resource)
  enabled:          bool        // default: false
  max_speculative:  uint8       // max ticks to run speculatively, default 4

When a peer's input is missing for the current tick:
  1. Predict input using InputBuffer.PredictInput() (repeat-last-input).
  2. Run the tick speculatively with predicted input.
  3. Mark the tick as speculative in the history.
  4. When the real input arrives:
     a. If it matches the prediction → confirm the speculative tick. Done.
     b. If it differs → rollback to the last confirmed tick:
        - Restore state from SnapshotManager.
        - Replay from confirmed tick to current tick with correct inputs.
        - This is identical to client-prediction rollback
          (see client-prediction.md §4.3).
```

Speculative execution converts lockstep into a hybrid model that tolerates jitter at the cost of occasional rollbacks. Recommended for fighting games and fast-paced action where freezing is worse than rare visual corrections.

### 4.4 Desync Detection

Since all peers should have identical state, checksums detect divergence:

```plaintext
DesyncProtocol (uses networking-system.md §4.7 DesyncDetector)
  check_interval: uint8    // compare checksums every N ticks, default 10

Each check_interval ticks:
  1. Compute CRC32 checksum of the full deterministic World state.
  2. Broadcast ChecksumMessage { tick, checksum } to all peers.

On receiving remote checksum:
  1. Compare against local checksum for the same tick.
  2. Match → ok.
  3. Mismatch → increment mismatch counter.
  4. If mismatch_count > tolerance (default: 3):
     → fire_event(DesyncDetected { peer, tick, local_checksum, remote_checksum })
     → Default behavior: pause game, display desync error with tick number.
     → Optional: attempt full state resync (send snapshot from designated authority).
```

**Desync debugging**: When a desync is detected, the system can optionally dump the full World state at the diverging tick to disk for offline comparison. This is the primary debugging tool for determinism bugs.

### 4.5 Late Join

A player joining a game in progress cannot replay from tick 0. Late join requires a state transfer:

```plaintext
Late Join Protocol:
  1. Joining peer connects and sends JoinRequest.
  2. Authority peer (host or server) pauses simulation for all peers.
  3. Authority takes a full snapshot: SnapshotManager.TakeSnapshot(current_tick).
  4. Authority sends snapshot to joining peer via ChannelID 2 (ReliableOrdered).
  5. Joining peer receives and restores snapshot.
  6. Authority broadcasts ResumeMessage with the tick to resume at.
  7. All peers resume lockstep from that tick.
```

**Pause-on-join** is the simplest approach. For games that cannot pause (competitive), the alternative is continuous background streaming — but this is complex and deferred to future work.

### 4.6 Topology Support

```plaintext
Peer-to-Peer (default for lockstep):
  Each peer sends inputs directly to all other peers.
  No authority — all peers are equal.
  Desync detection relies on majority consensus.
  Scales to ~8 players (N² connections).

Relay Server:
  All peers send inputs to a central relay.
  Relay aggregates inputs per tick and broadcasts to all peers.
  Relay does NOT simulate — it's a dumb forwarder.
  Advantages: simpler NAT traversal, single point for input ordering.
  Scales to ~16+ players.

Host-Authoritative:
  One peer is designated as authority.
  Authority collects inputs, simulates, broadcasts confirmed inputs.
  Authority can also validate inputs (anti-cheat).
  Trade-off: host has zero latency advantage.
```

`LockstepConfig` specifies the topology:

```plaintext
LockstepConfig (resource)
  topology:       Topology       // PeerToPeer | Relay | HostAuthoritative
  input_delay:    uint8          // default: 2
  input_timeout:  Duration       // default: 5s
  check_interval: uint8          // desync check every N ticks, default: 10
```

### 4.7 LockstepPlugin

```plaintext
LockstepPlugin
  config: LockstepConfig

Build(app):
  app.InsertResource(LockstepConfig{...})
  app.InsertResource(SpeculativeConfig{ enabled: false })
  app.AddSystem(FixedUpdate, LockstepScheduler)
  app.AddSystem(PreUpdate, LockstepInputReceiveSystem)
  app.AddSystem(PostUpdate, LockstepInputBroadcastSystem)
  app.AddSystem(PostUpdate, DesyncCheckSystem)
  app.AddEvent[DesyncDetected]()
  app.AddEvent[LockstepStalled]()  // emitted when waiting for peer input > 200ms
```

**Dependency**: Requires `NetworkPlugin`. Does NOT require `ReplicationPlugin` (no state replication in pure lockstep). May use `ReplicationPlugin` only for late join snapshot transfer.

## 5. Open Questions

- Should the engine provide **determinism testing tools** (run two identical simulations side by side, compare state each tick) as a first-party development utility?
- How should **AI bots** be handled? If each peer simulates all AI, the AI code must be deterministic. Alternatively, one peer runs AI and broadcasts AI inputs.
- Should lockstep support **variable input delay** that adapts to network conditions (increase delay when jitter rises, decrease when stable)?
- How should **replay** work? Lockstep replays are naturally just input recordings — but the replay must use the exact same engine version for determinism.
- Should **speculative execution** be in this spec or split into a separate micro-spec? It significantly increases complexity.

## 6. Implementation Notes

1. `LockstepScheduler` with 2-player local test (LoopbackBackend) — basic tick advancement on input receipt.
2. Input serialization and broadcast — end-to-end input exchange.
3. Desync detection — checksum generation and comparison.
4. Input delay tuning — configurable delay with UI feedback.
5. Late join — snapshot transfer protocol.
6. Speculative execution — opt-in, add last (effectively adds rollback to lockstep).

## 7. Drawbacks & Alternatives

**100% determinism requirement**: The strictest requirement of any netcode model. A single non-deterministic operation (unordered map iteration, platform-dependent float rounding, system clock access) causes a desync. The engine mitigates this with `DeterministicSchedule` and `DeterministicRNG`, but the game developer must also ensure their systems are deterministic. No engine can fully enforce this.

**Wait-for-slowest**: In pure lockstep, the game runs at the speed of the slowest connection. One player with packet loss causes all players to stutter. Input delay and speculative execution mitigate this but don't eliminate it. For games with >8 players or unreliable connections, client-server with prediction is more robust.

**No late join without pause**: Late join in lockstep requires either pausing (disruptive) or continuous state streaming (complex). Client-server models handle late join trivially because the server always has the full state.

**Bandwidth efficiency vs complexity**: Lockstep uses minimal bandwidth (inputs only) but requires every client to run the full simulation, including AI, physics, and all game logic. This means every client needs sufficient CPU power to simulate the entire world, not just what's visible.

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-28 | Initial draft — lockstep tick loop, input delay, speculative execution, desync detection, late join, topology |
| — | — | Planned examples: `examples/networking/` |
