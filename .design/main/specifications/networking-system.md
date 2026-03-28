# Networking System

**Version:** 0.2.0
**Status:** Draft
**Layer:** concept

## Overview

The Networking System defines the engine's boundary with multiplayer infrastructure. The engine itself is NOT a netcode framework — it provides deterministic primitives that netcode libraries build upon: a fixed-timestep simulation loop, snapshot/rollback state management, input serialization, and a transport abstraction. Game-layer networking (lobbies, matchmaking, authoritative servers) is explicitly out of scope and lives in separate binaries.

## Related Specifications

- [app-framework.md](app-framework.md) — FixedUpdate schedule, SubApp isolation, deployment architecture (§4.11)
- [time-system.md](time-system.md) — FixedTime drives deterministic simulation rate
- [scene-system.md](scene-system.md) — World serialization for snapshots
- [event-system.md](event-system.md) — Network events delivered through the standard event bus
- [input-system.md](input-system.md) — Input serialization for rollback
- [entity-system.md](entity-system.md) — Entity ID stability across snapshots
- [definition-system.md](definition-system.md) — Network boundary rule (§4.11)
- [task-system.md](task-system.md) — IO pool for network transport threads
- [transport.md](transport.md) — Detailed UDP transport layer: channels, reliability, packet structure, connection lifecycle

## 1. Motivation

Multiplayer games require deterministic simulation, state synchronization, and latency compensation. These needs conflict with a typical game engine's variable-rate rendering and non-deterministic system execution. Without engine-level support:

- Developers reimplement fixed-timestep loops per project, often with subtle bugs.
- Rollback requires snapshotting the entire World, but the engine provides no efficient delta mechanism.
- Input must be serialized and replayed identically — but input systems typically consume and discard input each frame.
- Transport layer code (TCP/UDP sockets) ends up interleaved with gameplay logic, violating the no-network-in-game-loop rule.

The Networking System provides the building blocks so that netcode libraries (rollback, lockstep, client-server) can be implemented as plugins without modifying engine internals.

## 2. Constraints & Assumptions

- The engine does NOT implement any specific netcode model (rollback, lockstep, server-authoritative). It provides primitives.
- No network I/O occurs inside the main game loop (`First` through `Last` schedules). All transport runs on IO pool threads.
- Determinism is opt-in per system. Systems that declare `Deterministic: true` are scheduled in a fixed order with fixed-point math.
- Snapshot serialization uses the same binary format as scene saves and hot-reload (see [scene-system.md](scene-system.md), [hot-reload.md](hot-reload.md)).
- The transport layer is pluggable — UDP, WebSocket, WebRTC, or custom protocols.
- Entity IDs must remain stable across snapshot/restore cycles (see [entity-system.md §4.8](entity-system.md)).

## 3. Core Invariants

- **INV-1**: No network I/O occurs on the main thread or inside any schedule from `First` through `Last`. Transport runs exclusively on IO pool threads.
- **INV-2**: Deterministic systems produce identical output given identical input, regardless of platform or frame rate.
- **INV-3**: World snapshots are self-contained — restoring a snapshot does not require any state beyond the snapshot itself.
- **INV-4**: Input replay produces identical simulation results as live execution (input determinism).
- **INV-5**: The transport abstraction never leaks protocol-specific types into gameplay code.

## 4. Detailed Design

### 4.1 Deterministic Simulation Layer

The engine provides a `DeterministicSchedule` that guarantees identical execution across clients:

```plaintext
DeterministicSchedule
  fixed_timestep:    Duration           // e.g., 16.666ms (60 Hz)
  system_order:      []SystemID         // explicit, deterministic order
  rng_seed:          uint64             // shared seed for deterministic RNG
  tick_number:       uint64             // monotonic simulation tick

Guarantees:
  1. Systems execute in declared order (no parallel execution).
  2. All math uses fixed-point or deterministic float operations.
  3. RNG is seeded per-tick from a shared seed + tick_number.
  4. No system reads wall-clock time — only tick_number and fixed_timestep.
```

**Opt-in**: Only systems registered with `AddDeterministicSystems()` run in this schedule. Non-deterministic systems (particles, audio) run in the regular `Update` schedule.

### 4.2 Snapshot Manager

Efficient World state capture for rollback and state synchronization:

```plaintext
SnapshotManager (resource)
  TakeSnapshot(tick: uint64) -> SnapshotHandle
  RestoreSnapshot(handle: SnapshotHandle)
  GetDelta(from: SnapshotHandle, to: SnapshotHandle) -> DeltaSnapshot

SnapshotHandle
  tick:          uint64
  data:          []byte           // serialized world state
  checksum:      uint32           // CRC32 for desync detection
  entity_count:  uint32

DeltaSnapshot
  tick_from:     uint64
  tick_to:       uint64
  changed:       []EntityDelta    // only entities that changed
  removed:       []EntityID
  added:         []EntitySnapshot
```

**Ring buffer**: The SnapshotManager maintains a configurable ring buffer of recent snapshots (default: 16). Older snapshots are evicted. This supports rollback windows of up to 16 ticks (~266ms at 60 Hz).

**Delta compression**: `GetDelta()` compares two snapshots component-by-component and returns only changed data. This is the primary mechanism for network state synchronization — send deltas, not full snapshots.

**Checksum verification**: Each snapshot includes a CRC32 checksum. Clients exchange checksums to detect desynchronization early.

### 4.3 Input Buffer

Input serialization and buffering for rollback replay:

```plaintext
InputBuffer (resource)
  RecordInput(tick: uint64, player: PlayerID, input: SerializedInput)
  GetInput(tick: uint64, player: PlayerID) -> (SerializedInput, bool)
  PredictInput(tick: uint64, player: PlayerID) -> SerializedInput

SerializedInput
  tick:      uint64
  player:    PlayerID
  data:      []byte              // deterministic serialization of input state
  checksum:  uint16              // input integrity check

InputBuffer internals:
  ring:          [MAX_BUFFER]map[PlayerID]SerializedInput
  MAX_BUFFER:    128             // ~2 seconds at 60 Hz
```

**Prediction**: When a remote player's input hasn't arrived yet, `PredictInput()` returns the last known input (repeat-last-input prediction). Netcode plugins can override this with more sophisticated prediction.

**Serialization**: Inputs are serialized to a compact binary format that is platform-independent. Button states are packed as bitfields. Axis values use fixed-point representation.

### 4.4 Rollback Coordinator

Orchestrates the rollback-resimulate cycle:

```plaintext
RollbackCoordinator
  confirmed_tick:    uint64      // last tick where all inputs are confirmed
  current_tick:      uint64      // current simulation tick
  max_rollback:      uint8       // maximum rollback depth (default: 8)

  OnRemoteInput(tick: uint64, player: PlayerID, input: SerializedInput):
    if tick <= confirmed_tick:
        return  // already confirmed, ignore late input

    // Store the confirmed input
    input_buffer.RecordInput(tick, player, input)

    if tick < current_tick:
        // Input arrived late — rollback needed
        rollback_to = tick
        snapshot = snapshot_manager.RestoreSnapshot(rollback_to)
        // Resimulate from rollback_to to current_tick
        for t := rollback_to; t <= current_tick; t++ {
            deterministic_schedule.RunTick(t, input_buffer)
            snapshot_manager.TakeSnapshot(t)
        }

    // Advance confirmed_tick if possible
    confirmed_tick = findLatestFullyConfirmedTick()
```

**This is a reference implementation**. Netcode plugins can replace the RollbackCoordinator entirely via the ServiceRegistry. The engine provides it as a starting point, not a mandate.

### 4.5 Transport Abstraction

Network transport runs on IO pool threads, completely isolated from the game loop. The full transport API — handle-based connection management, pre-configured delivery channels, packet structure, reliability layer, MTU discovery, and platform backends — is defined in [transport.md](transport.md).

Key design decisions (see transport.md for rationale):

- **Handle-based connections**: `ConnectionID` (uint32) instead of stateful connection objects — consistent with the engine's ID-centric ECS model (EntityID, AssetID).
- **Per-channel delivery**: Each `ChannelID` has a fixed `DeliveryMode` (Unreliable, ReliableUnordered, ReliableOrdered) configured at startup, not per-send. Enables per-channel buffer optimization and eliminates hot-path branching.
- **Separated concerns**: Connection lifecycle events (`Connected`, `Disconnected`) flow through the standard event bus. Payload data is consumed via `Drain()`. The transport never leaks protocol types into gameplay code (INV-5 of this spec).

**Implementations** (as separate plugins):

| Transport | Protocol | Use Case |
| :--- | :--- | :--- |
| UDPTransport | Raw UDP + custom reliability | Low-latency gameplay |
| WebSocketTransport | WebSocket over TCP | Browser clients |
| WebRTCTransport | WebRTC data channels | Browser P2P |
| SteamTransport | Steam Networking Sockets | Steam platform |

### 4.6 Network Message Pipeline

Messages flow between the transport layer and the game loop through a channel-based pipeline:

```plaintext
Transport Thread (IO Pool):
  ┌──────────────┐
  │  UDP Socket   │
  │  recv loop    │──→ inbound_channel ──→  Game Loop (PreUpdate):
  │               │                          NetworkReceiveSystem reads channel
  │               │                          Deserializes messages
  └──────────────┘                          Delivers as ECS events

Game Loop (PostUpdate):                     Transport Thread (IO Pool):
  NetworkSendSystem                          ┌──────────────┐
  Collects outbound messages ──→             │  UDP Socket   │
  outbound_channel ──→                       │  send loop    │
                                             └──────────────┘
```

**Channel-based isolation**: `inbound_channel` and `outbound_channel` are bounded Go channels. The transport thread never touches World state directly. The game loop never performs socket operations.

### 4.7 Desync Detection

Clients periodically exchange state checksums to detect simulation divergence:

```plaintext
DesyncDetector
  check_interval:    uint8       // compare checksums every N ticks (default: 10)
  tolerance:         uint8       // allow N consecutive mismatches before alert

  OnTick(tick: uint64):
    if tick % check_interval == 0:
        local_checksum = snapshot_manager.GetChecksum(tick)
        send_to_peers(ChecksumMessage{tick, local_checksum})

  OnRemoteChecksum(peer: PeerID, tick: uint64, remote_checksum: uint32):
    local_checksum = snapshot_manager.GetChecksum(tick)
    if local_checksum != remote_checksum:
        mismatch_count[peer]++
        if mismatch_count[peer] > tolerance:
            fire_event(DesyncDetected{peer, tick, local_checksum, remote_checksum})
            // Netcode plugin decides: force resync, disconnect, or log
```

### 4.8 Deterministic RNG

A seedable, platform-independent random number generator for deterministic simulation:

```plaintext
DeterministicRNG (resource)
  seed:      uint64
  state:     [4]uint64       // xoshiro256** state

  Next() -> uint64           // deterministic, portable
  Float01() -> float64       // [0.0, 1.0)
  Range(min, max: int) -> int

  ForkForEntity(entity: EntityID) -> DeterministicRNG
    // Derives a child RNG seeded from parent state + entity ID
    // Ensures entity-specific randomness is reproducible
```

**No `math/rand`**: The standard library's `math/rand` is NOT used in deterministic systems because its implementation may change between Go versions. The engine provides its own portable PRNG.

## 5. Open Questions

- Should the engine provide a built-in lobby/session system, or is that always game-layer?
- How should snapshot size be bounded for games with large worlds (streaming/chunking)?
- Should deterministic systems support SIMD operations, given SIMD behavior varies across architectures?
- How should spectators be handled — full state stream or delta-only?
- Should the transport abstraction support multicast for LAN discovery?
- Is interest management (spatial relevance filtering) an engine primitive or a netcode plugin concern?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-28 | Initial draft: deterministic simulation, snapshot/rollback, transport abstraction |
| 0.2.0 | 2026-03-28 | Replaced inline transport API with reference to transport.md; documented key design decisions |
| — | — | Planned examples: `examples/movement/` |
