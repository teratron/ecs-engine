# Replication

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

The Replication system synchronizes entity and component state between a server World and one or more client Worlds over the network. It answers three questions: *what* to replicate (component markers and filters), *to whom* (interest management / visibility), and *how often* (priority-based frequency control). The system builds on top of the transport layer for delivery, the change detection system for dirty tracking, the scene system for serialization, and the type registry for dynamic component handling. It does not dictate a netcode model — rollback, lockstep, and server-authoritative architectures all use the same replication primitives.

## Related Specifications

- [networking-system.md](networking-system.md) — Parent networking boundary; SnapshotManager, InputBuffer, DeterministicSchedule
- [transport.md](transport.md) — Channel-based message delivery; ChannelID 0/3 for unreliable state, ChannelID 1 for reliable events
- [change-detection.md](change-detection.md) — Tick-based dirty tracking (ChangedTick, AddedTick), Changed[T]/Added[T] filters
- [entity-system.md](entity-system.md) — Generational Entity IDs, EntityAllocator, entity remapping
- [component-system.md](component-system.md) — ComponentID, registration, clone behavior, lifecycle hooks
- [scene-system.md](scene-system.md) — DynamicSceneBuilder, SceneFilter, entity remapping during instantiation
- [type-registry.md](type-registry.md) — Runtime type metadata, serialization hooks, ReflectComponent
- [event-system.md](event-system.md) — Replication events delivered through the standard event bus
- [world-system.md](world-system.md) — World as the canonical state store, resources

## 1. Motivation

Multiplayer games must keep multiple Worlds in sync. Without a dedicated replication system, every project reimplements the same concerns:

- **What to sync**: Manually tagging components for network transfer, with no standard filter mechanism. Forgetting to exclude a server-only component leaks implementation details (or secrets) to clients.
- **To whom**: Sending full world state to every client wastes bandwidth. A server with 10,000 entities and 64 clients cannot afford to replicate everything to everyone — relevance filtering is essential.
- **How often**: Not all components change at the same rate or have the same importance. Position updates at 60 Hz are critical; inventory changes can wait. Without frequency control, either bandwidth is wasted or important updates are delayed.

The replication system solves these once at the engine level, giving netcode plugins (rollback, server-authoritative, lockstep) a clean interface to build on.

## 2. Constraints & Assumptions

- The replication system is **authoritative-agnostic**. It provides the primitives for sending and receiving state; the authority model (who is allowed to write) is a policy layer above.
- Replication uses the **existing change detection system** (tick-based). No parallel dirty-tracking mechanism.
- Only components explicitly marked for replication are sent. This is a whitelist model — nothing replicates by default.
- Entity ID mapping between server and client is managed by the replication system. Server entities and client entities have different IDs.
- Serialization uses the **type registry and reflection**, same as DynamicScene. Binary format for wire, not text.
- The replication system runs as ECS systems within the standard schedule, not on a separate thread. It reads from the transport layer's inbound queue and writes to the outbound queue.
- No encryption at this layer (see transport.md §2).

## 3. Core Invariants

- **INV-1**: Only components with an explicit `Replicated` marker (or registered in the replication whitelist) are sent over the network. Unmarked components are never serialized or transmitted.
- **INV-2**: Every replicated entity on the server has exactly one corresponding entity on each relevant client. The mapping is bijective within the client's visibility set.
- **INV-3**: Entity references inside replicated components are remapped from server IDs to client IDs before delivery to the game layer. No server-side EntityID ever leaks into client gameplay code.
- **INV-4**: A client never receives component data for entities outside its visibility set.
- **INV-5**: Replication never mutates the source World. It reads state and produces outbound messages. Inbound messages are applied through the command system (deferred mutations).
- **INV-6**: If a replicated entity is despawned on the server, a despawn message is sent to all clients that had the entity in their visibility set.

## 4. Detailed Design

### 4.1 Replication Markers

Components opt into replication via marker components and registration:

```plaintext
Replicated (marker component — zero-sized)
  Attached to entities that should be replicated.
  Entities without this marker are invisible to the replication system.

ServerOnly (marker component — zero-sized)
  Explicitly excludes a component from replication even if the entity is Replicated.
  Use for: server-side AI state, physics internals, authority bookkeeping.

ClientOnly (marker component — zero-sized)
  Marks components that exist only on clients and are never sent to the server.
  Use for: interpolation state, local visual effects, prediction buffers.
```

Component-level replication control is registered at plugin build time:

```plaintext
ReplicationConfig (resource)
  component_rules: map[ComponentID] -> ReplicationRule

ReplicationRule:
  Replicate           — send this component when the entity is Replicated
  ServerOnly          — never send to clients
  ClientOnly          — never send to server
  OnChange            — only send when Changed[T] is true (default for Replicate)
  OnInterval(ticks)   — send at most once every N simulation ticks
  OnEvent             — send only when an explicit replication event is triggered
```

Default rules registered by `ReplicationPlugin`:

```plaintext
Transform       → Replicate, OnChange
GlobalTransform → ServerOnly (derived, recomputed on client)
Name            → Replicate, OnEvent (rarely changes)
Replicated      → ServerOnly (marker, not sent as data)
```

### 4.2 Entity Mapping

Server and client Worlds have independent EntityAllocators. The replication system maintains a bidirectional mapping:

```plaintext
EntityMap (resource, per-connection on server; global on client)
  server_to_client: map[EntityID] -> EntityID
  client_to_server: map[EntityID] -> EntityID

  Map(server_id: EntityID) -> EntityID
    If server_to_client[server_id] exists → return it
    Else → allocate new client entity, record mapping, return new ID

  Unmap(server_id: EntityID)
    Remove from both maps. Client entity is despawned via command.

  Remap(component: ReflectedComponent) -> ReflectedComponent
    Walk all EntityID fields using type registry metadata.
    Replace each server EntityID with its client-side mapping.
    If a referenced entity has no mapping → use EntityID::PLACEHOLDER
      and queue a deferred resolution (the referenced entity may arrive
      in a later packet).
```

The `Map` pattern follows Bevy's `SceneEntityMapper` approach: new client entities are allocated on demand the first time a server entity is seen.

### 4.3 Visibility (Interest Management)

Not every client needs every entity. The visibility system determines which entities each client can see:

```plaintext
Visibility (server-side concept)
  fn is_visible(entity: EntityID, client: ConnectionID) -> bool

VisibilitySet (resource, per-connection on server)
  entities: EntitySet    // set of entities visible to this client
  added:    EntitySet    // entities that became visible this tick
  removed:  EntitySet    // entities that left visibility this tick
```

Visibility is computed by **VisibilityPolicy** — a pluggable interface:

```plaintext
VisibilityPolicy interface:
  Update(world: &World, sets: map[ConnectionID] -> &mut VisibilitySet)
```

Built-in policies:

```plaintext
ReplicateAll
  Every Replicated entity is visible to every client.
  Simplest policy. Suitable for small games (< 100 entities).

GridVisibility
  Spatial partitioning: 2D grid cells of configurable size.
  Each client has a "view position" (typically their controlled entity's position).
  Entities in cells within a configurable radius are visible.
  Entities outside are culled.

  GridVisibilityConfig (resource)
    cell_size:    float32   // world units per cell, default 64.0
    view_radius:  int       // cells in each direction, default 3
    y_culling:    bool      // apply vertical distance check, default false

CustomVisibility
  Game provides a function: fn(entity, client_view) -> bool
  For complex rules: faction visibility, fog of war, stealth.
```

**Visibility transitions** produce replication events:

```plaintext
When entity enters visibility for a client:
  → Send full component snapshot (all replicated components)
  → VisibilitySet.added includes the entity

When entity leaves visibility for a client:
  → Send EntityDespawn message (client removes the entity)
  → VisibilitySet.removed includes the entity
  → EntityMap.Unmap(entity) for this connection
```

### 4.4 Replication Messages

Messages sent over the transport layer:

```plaintext
ReplicationMessage (enum, sent via transport channels)

  EntitySpawn
    server_entity: EntityID
    components:    []SerializedComponent   // full snapshot of all replicated components
    // Sent when: entity first enters client's visibility set

  EntityDespawn
    server_entity: EntityID
    // Sent when: entity leaves visibility or is despawned on server

  ComponentUpdate
    server_entity: EntityID
    components:    []SerializedComponent   // only changed components
    // Sent when: replicated component changes (per ReplicationRule frequency)

  ComponentRemove
    server_entity: EntityID
    component_ids: []ComponentID
    // Sent when: a replicated component is removed from a Replicated entity

SerializedComponent
  component_id: ComponentID
  data:         []byte          // binary-serialized via type registry
  tick:         uint32          // server ChangeTick when this data was captured
```

**Channel assignment:**

```plaintext
EntitySpawn     → ChannelID 1 (ReliableUnordered) — must arrive, order doesn't matter
EntityDespawn   → ChannelID 1 (ReliableUnordered) — must arrive
ComponentUpdate → ChannelID 0 (Unreliable) for high-frequency (Transform)
                  ChannelID 1 (ReliableUnordered) for low-frequency (Health, Inventory)
ComponentRemove → ChannelID 1 (ReliableUnordered) — must arrive
```

The channel choice per component is determined by `ReplicationRule`. Components with `OnChange` and no explicit channel preference default to Unreliable (newest-wins). Components with `OnEvent` default to Reliable.

### 4.5 Delta Compression

Full component snapshots are expensive. The replication system compresses updates by tracking what each client has already acknowledged:

```plaintext
ClientAckState (per-connection on server)
  last_acked_tick: map[EntityID] -> uint32
    // The last server tick for which this client confirmed receipt, per entity.

Delta computation:
  For each entity visible to this client:
    acked_tick = client_ack_state[entity]
    for each replicated component on entity:
      if component.ChangedTick > acked_tick:
        include in outbound ComponentUpdate

  Client sends periodic ACK:
    AckMessage { tick: uint32 }  // "I have processed up to this server tick"
    Sent via ChannelID 1 (ReliableUnordered)
```

For components with internal structure (e.g., large inventories), the replication system supports **field-level delta** via type registry metadata:

```plaintext
DeltaSerializer interface (optional, registered per component type):
  Diff(old: []byte, new: []byte) -> []byte   // compute minimal diff
  Patch(base: []byte, diff: []byte) -> []byte // apply diff to base

// If no DeltaSerializer is registered → send full component data.
// Built-in DeltaSerializer for common patterns:
//   - Slice delta: added/removed/changed indices
//   - Struct delta: changed fields only (bitmask + values)
```

### 4.6 Priority and Frequency Control

Not all components are equally important. The replication system uses a priority queue to allocate bandwidth:

```plaintext
ReplicationPriority (per component type or per entity)
  base_priority:  float32    // default 1.0
  accumulator:    float32    // grows each tick when update is deferred

Each tick, for each connection:
  1. Collect all pending updates (entities with changed replicated components).
  2. For each update, compute effective_priority = base_priority + accumulator.
  3. Sort by effective_priority descending.
  4. Serialize and enqueue updates until bandwidth budget is exhausted.
  5. Updates that fit: reset accumulator to 0.
  6. Updates that don't fit: accumulator += base_priority (will have higher priority next tick).
```

This is the **priority accumulator** pattern: important updates go first, but less important updates eventually accumulate enough priority to be sent. No update is starved indefinitely.

**Bandwidth budget** is derived from transport layer settings:

```plaintext
max_replication_rate = NetworkSettings.max_send_rate * replication_budget_fraction
  replication_budget_fraction: float32, default 0.8
  // 80% of send rate is allocated to replication; 20% reserved for reliable messages, heartbeats
```

### 4.7 Receive Pipeline (Client-Side)

Client-side systems process inbound replication messages:

```plaintext
ReplicationReceiveSystem (runs in PreUpdate)
  1. Drain transport inbound queue for replication channels.
  2. Deserialize ReplicationMessage from each packet.
  3. For each message:

     EntitySpawn:
       client_entity = entity_map.Map(msg.server_entity)
       commands.Spawn(client_entity)
       for each component in msg.components:
         deserialized = type_registry.Deserialize(component.component_id, component.data)
         entity_map.Remap(deserialized)   // fix EntityID fields
         commands.Insert(client_entity, deserialized)
       fire_event(EntityReplicated { entity: client_entity, is_new: true })

     EntityDespawn:
       client_entity = entity_map.server_to_client[msg.server_entity]
       commands.Despawn(client_entity)
       entity_map.Unmap(msg.server_entity)

     ComponentUpdate:
       client_entity = entity_map.server_to_client[msg.server_entity]
       if client_entity not found → queue for deferred processing (entity may arrive later)
       for each component in msg.components:
         deserialized = type_registry.Deserialize(component.component_id, component.data)
         entity_map.Remap(deserialized)
         commands.Insert(client_entity, deserialized)  // overwrite existing
       fire_event(EntityReplicated { entity: client_entity, is_new: false })

     ComponentRemove:
       client_entity = entity_map.server_to_client[msg.server_entity]
       for each component_id in msg.component_ids:
         commands.Remove(client_entity, component_id)
```

All mutations go through the command system (INV-5): deferred, batched, safe.

### 4.8 Send Pipeline (Server-Side)

```plaintext
ReplicationSendSystem (runs in PostUpdate, after game logic)
  1. Update visibility sets via VisibilityPolicy.Update().
  2. For each connection:
     a. Process visibility transitions:
        - added entities → queue EntitySpawn (full snapshot)
        - removed entities → queue EntityDespawn
     b. For each visible entity:
        - Check Changed[T] for each replicated component.
        - Check ReplicationRule frequency constraints.
        - If update needed → compute priority, add to priority queue.
     c. Drain priority queue within bandwidth budget.
        - Serialize ComponentUpdate messages.
        - Enqueue in transport outbound.
  3. Process ACK messages → update ClientAckState.
```

### 4.9 Replication Events

Delivered through the standard event bus:

```plaintext
EntityReplicated
  entity:   EntityID       // client-side entity
  is_new:   bool           // true for first spawn, false for updates
  tick:     uint32         // server tick of the replicated state

EntityReplicationLost
  entity:   EntityID       // client-side entity
  reason:   LostReason     // LeftVisibility | ServerDespawn | ConnectionLost

ReplicationStats (resource, updated each frame)
  entities_replicated: uint32
  bytes_sent:          uint64   // replication bytes this frame
  bytes_received:      uint64
  updates_deferred:    uint32   // updates that didn't fit in bandwidth budget
  entity_map_size:     uint32   // number of mapped entities
```

### 4.10 NetworkAuthority

While the replication system is authority-agnostic (INV does not dictate who owns what), it provides a lightweight `NetworkAuthority` component for netcode plugins to build on:

```plaintext
NetworkAuthority (component)
  owner: AuthorityOwner

AuthorityOwner:
  Server                    — server is authoritative (default for all replicated entities)
  Client(ConnectionID)      — specific client owns this entity (client-authoritative movement)
  Predicted(ConnectionID)   — client predicts locally, server is authoritative but tolerates prediction

// Replication respects authority:
// - Server-owned: server sends updates, clients apply without question.
// - Client-owned: client sends updates for this entity, server relays to other clients.
// - Predicted: both sides send; reconciliation is a netcode plugin concern.
```

The replication system uses `NetworkAuthority` to determine **send direction**:

- `Server`-owned entities: server → clients (standard replication)
- `Client(id)`-owned entities: client `id` → server → other clients (relay)
- `Predicted(id)`: bidirectional; the netcode plugin's rollback system resolves conflicts

### 4.11 ReplicationPlugin

```plaintext
ReplicationPlugin
  config: ReplicationConfig

Build(app):
  app.InsertResource(ReplicationConfig{...})
  app.InsertResource(ReplicationStats{})
  app.AddEvent[EntityReplicated]()
  app.AddEvent[EntityReplicationLost]()
  app.AddSystem(PreUpdate, ReplicationReceiveSystem)
  app.AddSystem(PostUpdate, ReplicationSendSystem)
  app.AddSystem(PostUpdate, VisibilityUpdateSystem)

  // Register default component rules
  config.Register[Transform](Replicate, OnChange)
  config.Register[Name](Replicate, OnEvent)
  config.Register[GlobalTransform](ServerOnly)
```

`ReplicationPlugin` depends on `NetworkPlugin` (transport.md) and is typically added alongside a netcode plugin (rollback, server-authoritative) that provides the authority and reconciliation logic.

## 5. Open Questions

- **Snapshot vs delta**: Should the system support full World snapshots (for initial join / resync) in addition to per-entity deltas? The SnapshotManager in networking-system.md handles this at the World level, but replication-level snapshots (filtered by visibility) may be more bandwidth-efficient.
- **Component grouping**: Should frequently co-updated components (Transform + Velocity) be serialized together in one message to reduce per-component overhead?
- **Interpolation integration**: Should the replication system insert `PreviousTransform` and `InterpolationTarget` components automatically, or is that a separate concern?
- **Partial entity updates**: If only 1 of 5 replicated components changed, should the ComponentUpdate include just that one, or batch all? (Current design: just the changed one.)
- **Dormancy**: Should entities that haven't changed in N ticks be marked "dormant" and excluded from per-tick visibility checks to reduce server CPU?
- **Client-to-server input replication**: Should input be sent through the replication system or through a separate InputBuffer channel (as defined in networking-system.md §4.3)?

## 6. Implementation Notes

1. `ReplicationConfig` and marker components (`Replicated`, `ServerOnly`, `ClientOnly`) first — enables registration and tagging without any networking.
2. `EntityMap` and entity remapping — the core data structure, testable in isolation with `LoopbackBackend`.
3. `ReplicationSendSystem` with `ReplicateAll` visibility — simplest end-to-end path.
4. `ReplicationReceiveSystem` — completes the loop, testable with loopback transport.
5. `VisibilityPolicy` and `GridVisibility` — optimization layer, can be iterated after basic replication works.
6. Delta compression and priority accumulator — bandwidth optimization, add after functional correctness.

## 7. Drawbacks & Alternatives

**Whitelist-only model**: Requires explicit opt-in for every replicated component. The alternative — replicate everything by default with a blacklist — is simpler for prototyping but dangerous in production: new components accidentally leak to clients, wasting bandwidth or exposing server internals. The whitelist model is safer and standard in production engines (Unreal's `DOREPLIFETIME`, Source Engine's SendProxy).

**Reflection-based serialization**: Flexible but slower than code-generated serializers. The alternative — compile-time code generation for each replicated component — is faster but requires a build step and reduces runtime flexibility (editor integration, hot-reload). The reflection path is the default; performance-critical projects can register custom `DeltaSerializer` implementations per component.

**Per-connection EntityMap**: Each connected client maintains its own entity mapping. This uses O(N * E) memory (N clients, E entities). The alternative — shared entity IDs across all clients — avoids the mapping cost but requires a global allocator and leaks information (client can infer server entity count from IDs). Per-connection mapping is standard practice.

**Priority accumulator vs fixed frequency**: The accumulator adapts to bandwidth pressure automatically, but adds complexity compared to simple "send Transform every 2 ticks". The accumulator is strictly better when bandwidth is constrained; fixed frequency is a special case where accumulator base_priority equals the inverse of the desired interval.

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-28 | Initial draft — replication markers, entity mapping, visibility, delta compression, priority control |
| — | — | Planned examples: `examples/networking/` |
