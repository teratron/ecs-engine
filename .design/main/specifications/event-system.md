# Event System

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

The event system provides three communication mechanisms between systems: **Events** (broadcast, double-buffered), **Messages** (typed system-to-system, cursor-based), and **Observers** (reactive triggers that fire immediately). Each mechanism serves a different communication pattern.

## Related Specifications

- [world-system.md](world-system.md) — Events/messages stored in World resources
- [system-scheduling.md](system-scheduling.md) — Event readers/writers are system parameters
- [command-system.md](command-system.md) — Observers can trigger commands
- [entity-system.md](entity-system.md) — Entity events target specific entities

## 1. Motivation

Systems need to communicate without direct coupling. Three patterns emerge:
1. **Broadcast**: "Something happened" — any interested system can react (Events).
2. **Point-to-point**: "System A tells System B" — ordered, reliable delivery (Messages).
3. **Reactive**: "When X is added to entity Y, do Z immediately" — no frame delay (Observers).

## 2. Constraints & Assumptions

- Events persist for exactly 2 frames (double-buffered), then are dropped.
- Messages persist until all registered readers have consumed them (cursor-based).
- Observers fire synchronously during the trigger — they block the triggering operation.
- All three mechanisms are type-safe: each event/message/trigger has a concrete type.

## 3. Core Invariants

- **INV-1**: An EventReader sees each event exactly once, even if it runs multiple times per frame.
- **INV-2**: Events not consumed within 2 frames are silently dropped.
- **INV-3**: Messages are never dropped until all registered readers advance past them.
- **INV-4**: Observers execute immediately and can trigger other observers (chain), but must terminate (no infinite loops).
- **INV-5**: Entity events only fire observers attached to the target entity (or global observers).

## 4. Detailed Design

### 4.1 Events (Broadcast)

Double-buffered event bus. Writers push events; readers consume them.

```
// Writing
fn spawn_particles(mut writer: EventWriter[CollisionEvent]) {
    writer.Send(CollisionEvent{entity_a, entity_b, point})
}

// Reading
fn play_sound(mut reader: EventReader[CollisionEvent]) {
    for event in reader.Read() {
        audio.play("hit.wav")
    }
}
```

**Buffer rotation**: Each frame, the "current" buffer becomes the "previous" buffer, and a new empty buffer becomes "current". Events live for 2 frames, ensuring systems running in different schedules can see them.

### 4.2 Messages (System-to-System)

Cursor-based communication channel. Each reader maintains its own cursor position.

```
// Writing
fn ai_system(mut writer: MessageWriter[MoveCommand]) {
    writer.Write(MoveCommand{entity, direction})
}

// Reading
fn movement_system(mut reader: MessageReader[MoveCommand]) {
    for cmd in reader.Read() {
        // process movement command
    }
}
```

**Key differences from Events:**
- Messages are not frame-limited — they persist until all readers consume them.
- Each reader tracks its own cursor, so late readers don't miss messages.
- Messages are ordered (FIFO within a single writer).

### 4.3 Observers (Reactive Triggers)

Observers fire immediately when a trigger occurs. They do not wait for the next schedule tick.

```
// Global observer
world.AddObserver(func(trigger: On[OnAdd[Health]]) {
    log("Entity gained health component")
})

// Entity-targeted observer
world.Entity(player).Observe(func(trigger: On[Damage]) {
    health := trigger.Entity.Get[Health]()
    health.current -= trigger.amount
})
```

**Trigger types:**
- **Component lifecycle**: `OnAdd[T]`, `OnInsert[T]`, `OnReplace[T]`, `OnRemove[T]`
- **Custom events**: Any user-defined event type
- **Entity events**: Events with an `Entity` field — dispatched to entity's observers

**Execution model:**
1. Trigger fires (e.g., component added).
2. All matching observers execute immediately, synchronously.
3. Observers can enqueue Commands and trigger additional events.
4. Chained triggers execute depth-first.

### 4.4 Entity Events

Events that target specific entities. They propagate up the entity hierarchy (bubbling):

```
EntityEvent[Damage] triggered on child_entity:
  1. child_entity's observers fire
  2. parent_entity's observers fire (if any)
  3. grandparent_entity's observers fire (if any)
  ... up to root
```

Bubbling can be stopped by an observer calling `event.StopPropagation()`.

### 4.5 Event Registration

Events and messages must be registered with the App/World before use:

```
app.AddEvent[CollisionEvent]()
app.AddMessage[MoveCommand]()
```

Registration sets up the storage buffers and enables the scheduler to track access.

### 4.6 Conditional Event Handling

Run conditions based on events:

```
system.RunIf(on_event[CollisionEvent]())    // run only if events exist
system.RunIf(on_message[MoveCommand]())     // run only if messages exist
```

## 5. Open Questions

- Should event buses support priority ordering?
- Maximum event buffer size — should there be backpressure or just unbounded?
- Should observers support async execution (fire-and-forget to a worker thread)?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft from Bevy analysis |
