# Scripting System

**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview

The Scripting System provides a Lua bridge that allows gameplay logic to be written in Lua scripts and hot-reloaded without recompiling the engine. Scripts access the ECS through a sandboxed API that reads components via the TypeRegistry, mutates the World through CommandBuffers, and responds to events. The bridge is designed for gameplay iteration speed, not raw performance — performance-critical systems remain in Go.

## Related Specifications

- [type-registry.md](type-registry.md) — Scripts discover and access component types at runtime
- [command-system.md](command-system.md) — All World mutations from scripts go through CommandBuffers
- [event-system.md](event-system.md) — Scripts subscribe to and emit events
- [component-system.md](component-system.md) — Component data exposed to scripts as Lua tables
- [query-system.md](query-system.md) — Scripts iterate entities through a query API
- [asset-system.md](asset-system.md) — Lua scripts are loaded as assets with hot-reload
- [hot-reload.md](hot-reload.md) — Script hot-reload is in-process (no restart needed)
- [app-framework.md](app-framework.md) — ScriptingPlugin registration, ServiceRegistry access
- [definition-system.md](definition-system.md) — Flow definitions can reference scripts for custom actions

## 1. Motivation

Go is excellent for engine internals — type safety, performance, and concurrency. But for gameplay iteration, the compile-run cycle (even with hot-reload) creates friction:

- A designer wants to tweak enemy AI behavior and see the result instantly.
- A prototype system needs rapid experimentation before committing to a Go implementation.
- Mod support requires a sandboxed runtime that cannot crash the engine.
- Level scripts (cutscenes, triggers, dialogues) change frequently and are authored by non-programmers.

Lua is the standard choice for game scripting: lightweight (< 200KB), embeddable, fast (LuaJIT-class with gopher-lua), and familiar to game developers. The bridge provides a safe, ergonomic API that maps ECS concepts to Lua idioms.

## 2. Constraints & Assumptions

- Lua is the sole scripting language. Additional languages (e.g., Wren, JavaScript) are out of scope for v1.
- The Lua runtime is embedded via `gopher-lua` (pure Go, no CGo dependency — C24 compliant).
- Scripts cannot directly modify component memory. All mutations go through CommandBuffers.
- Scripts execute on the main thread within the `Update` schedule. No goroutine spawning from scripts.
- Script execution time is bounded by a configurable instruction limit (default: 1M instructions per frame). Exceeding the limit suspends the script and logs a warning.
- Hot-reload of scripts is in-process — the Lua VM reloads the changed file without engine restart.
- Scripts are sandboxed: no file I/O, no network access, no `os.execute`. Only engine API functions are exposed.

## 3. Core Invariants

- **INV-1**: Scripts cannot crash the engine. Lua runtime errors are caught, logged, and the offending script is disabled.
- **INV-2**: All World mutations from scripts are deferred through CommandBuffers, never immediate.
- **INV-3**: Script execution is bounded. A runaway script is suspended, not allowed to freeze the frame.
- **INV-4**: Scripts access only the engine's public API surface. Internal engine state is never exposed.
- **INV-5**: Script hot-reload does not interrupt the current frame. Reload applies at the next frame boundary.

## 4. Detailed Design

### 4.1 Lua VM Lifecycle

```plaintext
ScriptingPlugin.Build(app):
  vm = NewLuaVM()
  vm.SetInstructionLimit(config.MaxInstructions)
  vm.RegisterAPI(ecs_api)          // entity, component, query, event API
  vm.RegisterAPI(math_api)         // vec2, vec3, quat helpers
  vm.RegisterAPI(input_api)        // read-only input state
  vm.RegisterAPI(time_api)         // delta_time, elapsed, frame_count
  app.World().Services().Register[ScriptVM](vm)

Per-frame execution (Update schedule):
  ScriptUpdateSystem(vm: Res[ScriptVM], scripts: Query[ScriptComponent]):
    for entity, script in scripts:
      if script.enabled:
        vm.CallFunction(script.handle, "update", entity)
```

### 4.2 Script Component

Entities with scripted behavior carry a `ScriptComponent`:

```plaintext
ScriptComponent
  asset:         Handle[LuaScript]    // asset reference to the .lua file
  handle:        ScriptHandle         // VM-internal reference to loaded script instance
  enabled:       bool                 // runtime enable/disable
  state:         ScriptState          // Active | Suspended | Error
  error_message: string               // populated if state == Error
```

### 4.3 ECS API for Lua

The bridge exposes ECS operations as Lua functions:

```plaintext
-- Entity operations
local entity = ecs.spawn()                     -- returns entity ID
ecs.despawn(entity)                            -- deferred via Commands
local exists = ecs.is_alive(entity)

-- Component operations (read)
local transform = ecs.get(entity, "Transform")  -- returns Lua table
local has = ecs.has(entity, "Health")
local hp = ecs.get(entity, "Health").current

-- Component operations (write — all deferred)
ecs.set(entity, "Health", { current = 50 })     -- via CommandBuffer
ecs.insert(entity, "Poisoned", { duration = 5.0 })
ecs.remove(entity, "Poisoned")

-- Query iteration
for entity, transform, health in ecs.query("Transform", "Health") do
    if health.current <= 0 then
        ecs.despawn(entity)
    end
end

-- Event operations
ecs.send("PlayerDied", { player = entity })
ecs.on("CollisionEnter", function(event)
    -- handle collision
end)

-- Resource access (read-only)
local time = ecs.resource("Time")
local dt = time.delta

-- Service access (read-only)
local input = ecs.service("Input")
if input.pressed("space") then
    ecs.send("Jump", { entity = self })
end
```

### 4.4 Type Marshalling

Component data crosses the Go/Lua boundary through automatic marshalling:

```plaintext
Go struct → Lua table (on ecs.get):
  TypeRegistry provides field metadata.
  Each field is converted:
    int/float/bool/string → Lua primitive
    Vec2/Vec3/Quat        → Lua table with x, y, z, w fields
    []T                   → Lua table (array)
    map[K]V               → Lua table (hash)
    Handle[T]             → Lua userdata (opaque, passable to engine API)
    Entity                → Lua number (entity ID)

Lua table → Go struct (on ecs.set):
  Fields present in the Lua table overwrite the corresponding Go fields.
  Missing fields retain their current values (partial update).
  Extra fields in the Lua table are ignored with a warning.
  Type mismatches log an error and skip the field.
```

### 4.5 Script Hot-Reload

```plaintext
On AssetEvent[LuaScript]::Modified(handle):
  1. Find all entities with ScriptComponent.asset == handle.
  2. For each entity:
     a. Save the script's Lua-side state table (script.state).
     b. Unload the old script from the VM.
     c. Load the new script source.
     d. Call script:on_reload(saved_state) if defined.
     e. If load fails: set ScriptState = Error, keep old behavior.
  3. Log: "Script reloaded: {path} ({N} instances)"
```

**State preservation**: Scripts can store persistent state in a `self.state` table. This table survives hot-reload — the new script receives it via `on_reload()`. This enables iterating on script logic without losing runtime state.

### 4.6 Sandboxing

The Lua VM runs in a restricted environment:

```plaintext
Blocked standard library modules:
  io        — no file access
  os        — no process/env access
  debug     — no internal VM manipulation
  loadfile  — no arbitrary file loading
  dofile    — no arbitrary file execution

Allowed standard library modules:
  math      — full math library
  string    — string manipulation
  table     — table utilities
  coroutine — coroutines (for script-local async patterns)
  pairs/ipairs/type/tostring/tonumber — basic Lua built-ins

Instruction limit:
  VM checks instruction count every N ops (configurable).
  If limit exceeded: current script is suspended, error logged.
  Script can be resumed next frame if the issue was transient.
```

### 4.7 Coroutine Support

Scripts can use Lua coroutines for sequential logic that spans multiple frames:

```plaintext
-- Example: patrol behavior
function update(entity)
    if not self.patrol_routine then
        self.patrol_routine = coroutine.create(patrol)
    end
    coroutine.resume(self.patrol_routine, entity)
end

function patrol(entity)
    while true do
        move_to(entity, waypoint_a)
        wait_seconds(2.0)           -- yields for 2 seconds of game time
        move_to(entity, waypoint_b)
        wait_seconds(2.0)
    end
end

function wait_seconds(duration)
    local elapsed = 0
    while elapsed < duration do
        elapsed = elapsed + ecs.resource("Time").delta
        coroutine.yield()
    end
end
```

### 4.8 Error Handling

```plaintext
Script errors are contained per-entity:

1. Lua runtime error (syntax, nil access, type error):
   - Error caught by pcall wrapper.
   - ScriptComponent.state = Error
   - ScriptComponent.error_message = Lua error string
   - Entity logged: "Script error on entity {id}: {message}"
   - Script disabled for this entity. Other entities unaffected.

2. Instruction limit exceeded:
   - ScriptComponent.state = Suspended
   - Warning: "Script on entity {id} exceeded instruction limit"
   - Automatically retried next frame.
   - If exceeded 3 frames in a row: escalate to Error state.

3. Type marshalling error:
   - Individual field skipped with warning.
   - Remaining fields still processed.
   - Entity continues running.
```

### 4.9 Performance Budget

```plaintext
Script execution budget per frame:
  Default:      2ms (configurable via ScriptingConfig)
  Measurement:  wall-clock time for all script updates combined
  Enforcement:  if budget exceeded, remaining scripts deferred to next frame
                (round-robin fairness — deferred scripts run first next frame)

Profiling integration:
  Each script update emits a "Script:{asset_name}" profiling span
  (see profiling-protocol.md §4.5 for auto-instrumentation).
```

## 5. Open Questions

- Should scripts be able to define new component types at runtime, or only use pre-registered Go types?
- Should there be a visual scripting layer (node graphs) that compiles to Lua?
- How should script-to-script communication work — direct function calls or event-only?
- Should coroutine state persist across hot-reload, or only `self.state` tables?
- Is `gopher-lua` performance sufficient, or should LuaJIT (via CGo) be an optional backend?
- Should scripts have read access to other entities' script state (for AI coordination)?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-28 | Initial draft: Lua bridge, ECS API, sandboxing, hot-reload, coroutines |
| — | — | Planned examples: `examples/app/` |
