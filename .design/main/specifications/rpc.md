# Network RPC

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

Network RPC (Remote Procedure Call) provides typed, one-shot message passing between server and clients over the transport layer. Unlike replication — which continuously synchronizes component state — RPC delivers discrete actions: "cast spell X on target Y", "request respawn", "kick player Z". RPCs integrate with the engine's event system: sending an RPC on one peer fires a typed event on the receiving peer. No polling, no manual deserialization in gameplay code.

## Related Specifications

- [event-system.md](event-system.md) — RPCs are delivered as typed ECS events on the receiver
- [transport.md](transport.md) — RPCs sent via ChannelID 1 (ReliableUnordered) by default; configurable per RPC
- [replication.md](replication.md) — Replication handles continuous state; RPC handles discrete actions
- [networking-system.md](networking-system.md) — Network message pipeline (§4.6) for send/receive flow
- [type-registry.md](type-registry.md) — RPC payload serialization via registered types
- [command-system.md](command-system.md) — Server-side RPC handlers may use commands for deferred mutations

## 1. Motivation

Replication synchronizes *state* — what the world looks like. But many gameplay actions are *events* — things that happen once and have no persistent state:

- A client requests to cast a spell. The server validates and executes it.
- The server tells all clients to play a kill-feed animation.
- A client sends a chat message.
- The server forces a client to load a new level.

Without RPC, developers encode these as temporary components (spawn a "SpellRequest" entity, process it, despawn) or as byte blobs stuffed into a generic "message" channel. Both approaches are error-prone and lack type safety. RPC provides a clean, typed interface: define a struct, register it, call `rpc.Send()`, receive it as an event.

## 2. Constraints & Assumptions

- RPCs are fire-and-forget by default. Reliable delivery is handled by the transport channel, not the RPC system.
- RPCs are one-shot messages, not continuous state. For continuous synchronization, use replication.
- RPC payload must be serializable via the type registry. No closures, no function pointers.
- RPCs have a **direction**: Client→Server, Server→Client, or Server→AllClients. Bidirectional RPCs (request-response) are composed from two unidirectional RPCs.
- Maximum RPC payload size is bounded by transport MTU. Large payloads are fragmented by the transport layer.
- The RPC system does not provide authentication or authorization. Validating that a client is allowed to invoke a given RPC is the game's responsibility.

## 3. Core Invariants

- **INV-1**: An RPC registered on both peers produces a typed event on the receiver. The sender never receives its own RPC as an event.
- **INV-2**: RPC delivery order follows the transport channel semantics. ReliableUnordered: guaranteed delivery, arbitrary order. ReliableOrdered: guaranteed delivery, in-order.
- **INV-3**: Unknown or unregistered RPC type IDs received from the network are silently dropped and logged as a warning. They do not crash the receiver.
- **INV-4**: RPC handlers run within the standard ECS schedule (as event readers). They never run on the transport thread.

## 4. Detailed Design

### 4.1 RPC Definition

An RPC is defined as a plain struct registered in the RPC registry:

```plaintext
// Example RPC definitions (pseudo-code):

CastSpellRequest
  caster:     EntityID    // remapped via EntityMap on receive
  spell_id:   uint16
  target:     EntityID    // remapped
  direction:  ClientToServer

ChatMessage
  sender:     ConnectionID
  text:       string
  direction:  ServerToAllClients

ForceLoadLevel
  level_name: string
  direction:  ServerToClient   // targeted to specific client
```

### 4.2 RPC Registry

```plaintext
RpcRegistry (resource)
  rpcs: map[RpcTypeID] -> RpcDefinition

RpcDefinition
  type_id:    RpcTypeID          // unique uint16, assigned at registration
  name:       string             // human-readable, for logging
  direction:  RpcDirection
  channel:    ChannelID          // transport channel, default: ChannelID 1 (ReliableUnordered)
  type_info:  TypeRegistryEntry  // serialization metadata from type registry

RpcDirection:
  ClientToServer
  ServerToClient      // targeted to one client
  ServerToAllClients  // broadcast to all connected clients
  ServerToGroup       // broadcast to a subset (e.g., team, visibility set)
```

Registration happens at plugin build time:

```plaintext
app.RegisterRpc[CastSpellRequest](RpcConfig{
  direction: ClientToServer,
  channel:   ChannelID(1),
})
```

`RpcTypeID` is a compact `uint16` agreed upon by both peers during the connection handshake (exchanged alongside channel configuration). The registry maps between Go types and wire IDs.

### 4.3 Sending RPCs

```plaintext
RpcSender (system parameter, analogous to EventWriter)
  Send[T](target: RpcTarget, payload: T)

RpcTarget:
  Server                          // client → server (only valid direction for clients)
  Client(ConnectionID)            // server → specific client
  AllClients                      // server → all connected clients
  Group([]ConnectionID)           // server → subset of clients
  Except(ConnectionID)            // server → all except one (e.g., don't echo to sender)

Serialization:
  1. Look up RpcTypeID from RpcRegistry for type T.
  2. Serialize payload via type registry (binary format).
  3. Remap EntityID fields from local IDs to server/client IDs via EntityMap.
  4. Construct wire message:
     RpcMessage
       rpc_type_id: uint16
       payload:     []byte
  5. Enqueue in transport outbound for the configured channel.
```

### 4.4 Receiving RPCs

```plaintext
RpcReceiveSystem (runs in PreUpdate, after transport Drain)
  1. Read inbound packets from RPC channel(s).
  2. For each packet:
     a. Deserialize RpcMessage header (rpc_type_id).
     b. Look up RpcDefinition in RpcRegistry.
     c. If unknown → log warning, drop.
     d. Validate direction (client should not receive ClientToServer RPCs, etc.).
     e. Deserialize payload via type registry.
     f. Remap EntityID fields via EntityMap.
     g. Emit as typed ECS event: EventWriter[T].Send(deserialized_payload).
  3. Gameplay systems read RPCs as standard events:
     fn handle_spell_requests(events: EventReader[CastSpellRequest], commands: Commands) {
       for event in events.Iter() {
         // Validate and execute spell
       }
     }
```

### 4.5 Request-Response Pattern

Some RPCs need a response (e.g., "request inventory" → server sends inventory data). This is composed from two unidirectional RPCs:

```plaintext
// Client sends request:
InventoryRequest
  player: EntityID
  direction: ClientToServer

// Server sends response:
InventoryResponse
  player:    EntityID
  items:     []ItemData
  request_id: uint32        // correlates with the request
  direction: ServerToClient

// Client-side usage:
rpc_sender.Send(Server, InventoryRequest{ player: local_entity })

// Server-side handler:
fn handle_inventory_request(
  events: EventReader[InventoryRequest],
  rpc: RpcSender,
  query: Query[Inventory],
) {
  for event in events.Iter() {
    inventory = query.Get(event.player)
    rpc.Send(Client(event.connection_id), InventoryResponse{
      player: event.player,
      items: inventory.items,
      request_id: event.request_id,
    })
  }
}
```

The `request_id` is a client-generated correlation token. The engine does not enforce request-response pairing — it's a convention. A future extension could provide a `RpcFuture` that automatically correlates responses.

### 4.6 RPC Rate Limiting

To prevent abuse (client flooding server with RPCs):

```plaintext
RpcRateLimit (resource, server-side)
  global_limit:       int      // max RPCs per second from any client, default 100
  per_type_limit:     map[RpcTypeID] -> int   // per-RPC-type limit, optional
  violation_action:   RateLimitAction

RateLimitAction:
  Drop         — silently drop excess RPCs
  Warn         — drop and log warning
  Disconnect   — disconnect the offending client

RpcRateLimitSystem (runs in PreUpdate, before RpcReceiveSystem)
  Count RPCs per connection per second.
  If exceeded → apply violation_action.
```

### 4.7 RpcPlugin

```plaintext
RpcPlugin

Build(app):
  app.InsertResource(RpcRegistry{})
  app.InsertResource(RpcRateLimit{ global_limit: 100 })
  app.AddSystem(PreUpdate, RpcRateLimitSystem)
  app.AddSystem(PreUpdate, RpcReceiveSystem)

  // RpcSender is a system parameter, auto-injected like EventWriter
```

**Dependency**: Requires `NetworkPlugin` (transport). Compatible with any sync model (prediction, interpolation, lockstep).

## 5. Open Questions

- Should the engine provide a built-in `RpcFuture` for request-response correlation, or is the `request_id` convention sufficient?
- Should RPCs support **batching** — multiple small RPCs packed into a single transport message to reduce per-message overhead?
- How should RPCs interact with **NetworkAuthority**? Should a client-owned entity be allowed to send server-to-client RPCs on behalf of that entity?
- Should there be a **broadcast with acknowledgement** variant (server waits for all clients to confirm receipt)?
- Should RPC registration be dynamic (register at runtime) or static-only (register at plugin build time)?

## 6. Implementation Notes

1. `RpcRegistry` with compile-time type→ID registration — core data structure.
2. `RpcSender` and `RpcReceiveSystem` with simple serialization — end-to-end message passing.
3. Entity ID remapping in RPC payloads — integrate with EntityMap from replication.md.
4. Rate limiting — server-side protection, add after basic RPC works.
5. Integration test: client sends RPC, server receives as event, server responds — full loop with LoopbackBackend.

## 7. Drawbacks & Alternatives

**No built-in request-response**: The RPC system is unidirectional fire-and-forget. Request-response is a convention (request_id), not a first-class feature. The alternative — a full RPC framework with futures, timeouts, and retries — adds significant complexity. The convention approach is simpler and sufficient for most game RPCs where the "response" is often just a state change visible through replication.

**Flat type ID space**: `uint16` limits to 65,536 RPC types. Sufficient for any game, but if multiple plugins register RPCs independently, ID collisions must be prevented. The registry assigns IDs sequentially; plugins must coordinate or use namespaced ranges.

**No encryption or signing**: RPC payloads are sent in plaintext. A malicious client can forge RPCs. Validation is the game's responsibility. This is consistent with the transport layer's no-encryption-in-v1 stance.

**Entity remapping overhead**: RPCs with EntityID fields require a registry walk to find and remap IDs. For RPCs without entity references, this is wasted work. Mitigation: the type registry marks whether a type contains EntityID fields; skip remapping if not.

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-28 | Initial draft — RPC definition, registry, send/receive, request-response pattern, rate limiting |
| — | — | Planned examples: `examples/networking/` |
