# Transport

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

The Transport layer provides reliable and unreliable message delivery between peers over UDP. It abstracts the network socket behind a `NetworkTransport` interface, manages connection lifecycle (handshake, heartbeat, disconnect), and exposes typed channels with configurable delivery guarantees. All networking in the engine flows through this layer — replication, RPC, snapshots, and lockstep messages are all sent as typed payloads through transport channels. No TCP is used at runtime; reliability is implemented in user-space on top of UDP.

## Related Specifications

- [app-framework.md](app-framework.md) — NetworkPlugin registers as a SubApp; ServiceRegistry for cross-system access
- [event-system.md](event-system.md) — Connection lifecycle events delivered through the standard event bus
- [diagnostic-system.md](diagnostic-system.md) — Latency, jitter, packet loss surfaced as diagnostics
- [platform-system.md](platform-system.md) — Socket backend varies per platform (POSIX, WASM, console)

## 1. Motivation

Every networking model — snapshot interpolation, client prediction, lockstep — needs the same foundation: send bytes from A to B with configurable reliability. Without a shared transport layer each model would reimplement sockets, connection management, heartbeats, and MTU handling. The transport layer solves this once, cleanly, and lets higher-level systems focus on game logic rather than packet headers.

The choice of UDP over TCP is deliberate. TCP's retransmission and head-of-line blocking add latency that is unacceptable for real-time games. A position update that arrives 200 ms late is worse than no update — the game should have sent a newer one by then. UDP lets the engine decide per-channel whether reliability matters, and pay only for the guarantees actually needed.

## 2. Constraints & Assumptions

- The engine uses **UDP exclusively** for game traffic. TCP may be used for out-of-band operations (lobby, matchmaking) handled by external backend services.
- Maximum transmission unit (MTU) is conservatively set at **1200 bytes** per packet, safely below typical internet path MTU of 1500 bytes minus headers.
- Messages larger than MTU are **fragmented** by the transport layer and reassembled on the receiver. Fragmented messages use the `Reliable` channel internally.
- A single `NetworkTransport` instance per engine process. Multiple simultaneous connections are supported (server hosting N clients).
- The transport layer is platform-agnostic. Socket implementation is behind the `SocketBackend` interface (§4.9).
- No encryption in v1. TLS/DTLS is a future concern. Sensitive data (auth tokens) must not travel over game channels.
- The transport layer runs on a **dedicated network goroutine**, separate from the main ECS loop. Interaction with ECS happens through thread-safe queues.

## 3. Core Invariants

- **INV-1**: `Unreliable` channel messages may be dropped or arrive out of order. No retransmission. The receiver must tolerate gaps.
- **INV-2**: `Reliable` channel messages are delivered exactly once, in order, or the connection is declared lost.
- **INV-3**: `ReliableUnordered` channel messages are delivered exactly once but may arrive in any order.
- **INV-4**: Heartbeats are sent automatically every `heartbeat_interval`. A connection with no traffic for `timeout_duration` is closed and a `Disconnected` event is emitted.
- **INV-5**: The transport layer never blocks the ECS main loop. Send and receive operations are non-blocking; results are queued and consumed each frame.
- **INV-6**: Every outbound packet includes a monotonically increasing sequence number per channel. Receivers use this to detect loss and reorder.

## 4. Detailed Design

### 4.1 NetworkTransport Interface

The top-level service registered in `ServiceRegistry`:

```plaintext
NetworkTransport interface:
  // Connection management
  Listen(addr: SocketAddr) -> error
  Connect(addr: SocketAddr) -> (ConnectionID, error)
  Disconnect(id: ConnectionID, reason: string)
  Connections() -> []ConnectionID

  // Sending
  Send(id: ConnectionID, channel: ChannelID, payload: []byte) -> error
  Broadcast(channel: ChannelID, payload: []byte)             // send to all connections

  // Receiving (called each frame by NetworkPlugin)
  Drain() -> []InboundPacket

  // Stats
  Stats(id: ConnectionID) -> ConnectionStats
```

`ConnectionID` is an opaque `uint32`. `ChannelID` is a `uint8` (max 256 channels per connection, more than sufficient).

### 4.2 Channel Types

Every channel is registered at startup with a fixed delivery guarantee:

```plaintext
ChannelConfig
  id:           ChannelID
  delivery:     DeliveryMode
  max_message_size: int        // bytes before fragmentation, default 1200

DeliveryMode:
  Unreliable
    — Fire and forget. No ACK, no retransmission.
    — Use for: positions, rotations, health bars, anything where newest > oldest.
    — Overhead: 4 bytes header per message.

  ReliableUnordered
    — ACK-based delivery. Each message delivered exactly once.
    — Arrival order not guaranteed.
    — Use for: game events, RPC calls, pickups, kills.
    — Overhead: 8 bytes header + ACK bookkeeping.

  ReliableOrdered
    — ACK-based + reordering buffer. Exactly once, in send order.
    — Head-of-line blocking if a packet is lost.
    — Use for: chat messages, critical state changes.
    — Overhead: 12 bytes header + ACK + reorder buffer.
```

Default channels registered by `NetworkPlugin`:

```plaintext
ChannelID 0 — Unreliable       (state replication: positions, rotations)
ChannelID 1 — ReliableUnordered (events, RPC)
ChannelID 2 — ReliableOrdered   (chat, critical notifications)
ChannelID 3 — Unreliable        (snapshot delta, high-frequency updates)
```

Game code can register additional channels at plugin build time. Channel IDs must be agreed on by both peers at connection time (exchanged during handshake).

### 4.3 Packet Structure

Every UDP datagram has a fixed header followed by one or more channel message frames:

```plaintext
UDPDatagram
  [Header: 8 bytes]
    protocol_id:  uint16   // magic value to reject foreign packets
    connection_id: uint16  // sender's ConnectionID on receiver side
    packet_seq:   uint16   // monotonic datagram sequence number
    flags:        uint8    // ACK_PRESENT | FRAGMENT | COMPRESSED
    channel_count: uint8   // number of message frames in this datagram

  [MessageFrame × N]
    channel_id:   uint8
    msg_seq:      uint16   // per-channel sequence number
    payload_len:  uint16
    payload:      []byte
```

Multiple messages can be **batched** into one datagram to reduce per-packet overhead. The transport layer accumulates outbound messages each frame and flushes them in as few datagrams as possible before the end of the network tick.

### 4.4 Reliability Layer

Reliability for `ReliableUnordered` and `ReliableOrdered` channels:

```plaintext
Sender side:
  Maintain a sliding window of unACKed messages (default window size: 256).
  Each outbound packet includes an ACK bitfield for the last 32 received
    datagrams from the peer (piggybacked ACKs — no separate ACK packets).
  Unacknowledged messages are retransmitted after RTO (Retransmission Timeout).
  RTO is computed using EWMA of round-trip time: RTO = SRTT + 4*RTTVAR.

Receiver side:
  Track received msg_seq per channel.
  Unreliable: discard if older than highest seen (sequence gap tolerance: 64).
  ReliableUnordered: deliver immediately, ACK, record to detect duplicates.
  ReliableOrdered: buffer out-of-order messages, deliver in-order sequence.

Duplicate detection:
  Bitfield of last 256 received msg_seq values per channel.
  Duplicates are silently dropped.
```

### 4.5 Connection Lifecycle

```plaintext
stateDiagram-v2
  [*] --> Disconnected
  Disconnected --> Connecting : Connect() called
  Connecting --> Connected : handshake complete
  Connecting --> Disconnected : timeout or rejection
  Connected --> Disconnecting : Disconnect() called
  Connected --> Disconnected : timeout (no heartbeat)
  Disconnecting --> Disconnected : graceful close confirmed
```

**Handshake sequence (client-initiated):**

```plaintext
Client → Server: ConnectRequest { protocol_id, client_version, channels: []ChannelConfig }
Server → Client: ConnectAccept { connection_id, server_channels: []ChannelConfig }
  OR
Server → Client: ConnectReject { reason: string }
Client → Server: ConnectAck
// Connection is now CONNECTED on both sides
```

The handshake is sent over `ReliableOrdered` channel 2 before any game channels are active. Version mismatch causes `ConnectReject`.

**Heartbeat:**

```plaintext
NetworkSettings
  heartbeat_interval: Duration    // default 250 ms
  timeout_duration:   Duration    // default 5 s
```

If no packet is received within `timeout_duration`, the connection transitions to `Disconnected` and a `Disconnected` event is emitted with `reason: Timeout`.

### 4.6 Connection Events

Delivered through the standard event bus each frame after `Drain()`:

```plaintext
Connected
  connection_id: ConnectionID
  remote_addr:   SocketAddr
  is_server:     bool            // true if we accepted this connection as server

Disconnected
  connection_id: ConnectionID
  remote_addr:   SocketAddr
  reason:        DisconnectReason

DisconnectReason:
  Graceful        — peer called Disconnect() cleanly
  Timeout         — no heartbeat within timeout_duration
  ProtocolError   — malformed packet or version mismatch
  Rejected        — server refused the connection
  LocalClose      — we called Disconnect()

InboundPacket (internal, consumed by replication/RPC layers — not exposed directly to game code)
  connection_id: ConnectionID
  channel_id:    ChannelID
  payload:       []byte
```

### 4.7 Bandwidth and Congestion

The transport layer tracks bandwidth usage per connection:

```plaintext
ConnectionStats
  rtt:              Duration     // round-trip time (EWMA)
  rtt_variance:     Duration     // RTTVAR
  packet_loss:      float32      // fraction 0..1 over last 128 packets
  bytes_sent:       uint64       // total bytes sent
  bytes_received:   uint64       // total bytes received
  send_rate:        float32      // bytes/sec (current)
  receive_rate:     float32      // bytes/sec (current)
```

**Send rate limiting:**

```plaintext
NetworkSettings
  max_send_rate: int    // bytes per second per connection, default 256 KB/s
```

If the outbound queue for a connection exceeds `max_send_rate`, lower-priority channels (Unreliable) are throttled first. Reliable messages are never dropped due to rate limiting — they are queued.

**No congestion control in v1.** CUBIC or BBR congestion control is a future addition. For LAN and controlled environments, rate limiting is sufficient.

### 4.8 MTU Discovery

The transport layer performs Path MTU Discovery (PMTUD) on connection establishment:

```plaintext
1. Send probe packet at 1400 bytes. If ACKed → use 1400.
2. If dropped after 3 retries → try 1200 bytes.
3. If dropped → try 576 bytes (minimum guaranteed).
4. Store negotiated MTU per connection.
```

MTU is re-probed every 60 seconds to adapt to path changes. Fragmentation kicks in for messages larger than the negotiated MTU.

### 4.9 SocketBackend Interface

Platform-specific socket implementation:

```plaintext
SocketBackend interface:
  Bind(addr: SocketAddr) -> error
  SendTo(addr: SocketAddr, data: []byte) -> error
  RecvFrom() -> ([]byte, SocketAddr, error)   // non-blocking
  Close()
  LocalAddr() -> SocketAddr
```

Built-in backends:

| Platform | Backend |
| :--- | :--- |
| Windows / Linux / macOS | `UDPSocketBackend` via `net.UDPConn` (stdlib) |
| Web / WASM | `WebRTCDataChannelBackend` (unreliable) or `WebSocketBackend` (reliable) |
| Headless / test | `LoopbackBackend` — in-process, zero latency, configurable loss simulation |

`LoopbackBackend` is essential for testing: two `NetworkTransport` instances in the same process communicate via in-memory queues, with optional simulated latency and packet loss for robustness testing.

### 4.10 Simulated Network Conditions

For testing and development, the transport layer supports network condition simulation:

```plaintext
NetworkSimulation (Resource, debug builds only)
  enabled:         bool
  latency:         Duration     // added to all outbound packets (one-way)
  jitter:          Duration     // random variance around latency
  packet_loss:     float32      // fraction of packets dropped 0..1
  packet_duplicate: float32     // fraction of packets duplicated 0..1
  bandwidth_limit: int          // bytes/sec cap, 0 = unlimited

// Usage:
world.SetResource(NetworkSimulation {
    enabled:      true,
    latency:      50ms,
    jitter:       10ms,
    packet_loss:  0.05,   // 5% loss
})
```

This is the primary tool for validating that higher-level networking systems (client prediction, snapshot interpolation) handle adverse conditions correctly. Disable in release builds via the `networksim` build tag.

### 4.11 NetworkPlugin

```plaintext
NetworkPlugin
  settings: NetworkSettings

Build(app):
  app.InsertResource(NetworkSettings{...})
  app.InsertResource(NetworkSimulation{ enabled: false })
  app.AddEvent[Connected]()
  app.AddEvent[Disconnected]()
  app.InsertSubApp("network", NetworkSubApp{})
  // NetworkSubApp runs on dedicated goroutine:
  //   tick: Drain() → dispatch inbound → flush outbound
  services.Register[NetworkTransport](transport)
```

The network SubApp ticks at a configurable rate (default: every frame, in sync with `FixedUpdate`) but on its own goroutine. ECS systems access the transport via `services.Get[NetworkTransport]()` and thread-safe inbound/outbound queues.

## 5. Open Questions

- Should channel configuration be static (set at `NetworkPlugin` build time) or dynamic (registered per-connection during handshake)? Static is simpler and safer; dynamic allows peers to negotiate custom channels.
- **Encryption**: should v1 include optional DTLS support behind a build tag, or defer entirely to v2? Shipping without encryption is a security risk for any game with competitive elements.
- Should `max_send_rate` throttle per-connection or globally across all connections (important for servers hosting many clients)?
- **IPv6**: the `SocketAddr` type should support both IPv4 and IPv6. Is dual-stack (listening on both simultaneously) required in v1?
- Should `LoopbackBackend` support more than two peers (for simulating server + N clients in a single test process), or is two-peer sufficient?
- **NAT traversal**: STUN/TURN/ICE for peer connections behind NAT — required for any consumer-facing multiplayer. Out of scope for transport layer itself, but should the interface be designed to accommodate a future NAT traversal layer above it?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-27 | Initial draft — UDP transport, channel types, reliability layer, connection lifecycle, MTU discovery, network simulation |
