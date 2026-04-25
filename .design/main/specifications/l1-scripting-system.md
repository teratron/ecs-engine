# Scripting System

**Version:** 0.2.0
**Status:** Draft
**Layer:** concept
**Priority:** Deferred (post-v1)

## Overview

The Scripting System is **deferred** until the core engine is validated. The current architecture relies on two existing mechanisms for rapid iteration:

1. **[definition-system.md](l1-definition-system.md)** — JSON-based data-driven content (UI, scenes, flows, templates) with in-process hot-reload (~100ms).
2. **[hot-reload.md](l1-hot-reload.md)** — Go code hot-restart with World state snapshot (~1.2s).

Together, these cover the primary iteration workflows without introducing a scripting runtime. A scripting layer becomes valuable when:

- **Mod support** is needed (sandboxed user-generated content).
- **Non-programmer authoring** of gameplay logic (designers writing AI, triggers, cutscenes).
- **Sub-second code iteration** is required beyond what Go hot-restart provides.

This specification preserves the architectural design and API shape for future implementation, and identifies two candidate runtimes for evaluation.

## Related Specifications

- [type-registry.md](l1-type-registry.md) — Scripts would discover and access component types at runtime
- [command-system.md](l1-command-system.md) — All World mutations from scripts go through CommandBuffers
- [event-system.md](l1-event-system.md) — Scripts subscribe to and emit events
- [query-system.md](l1-query-system.md) — Scripts iterate entities through a query API
- [asset-system.md](l1-asset-system.md) — Scripts loaded as assets with hot-reload
- [hot-reload.md](l1-hot-reload.md) — Go code hot-restart (the primary iteration mechanism for now)
- [definition-system.md](l1-definition-system.md) — JSON data-driven layer (covers most non-code iteration needs)
- [app-framework.md](l1-app-framework.md) — ScriptingPlugin registration, ServiceRegistry access

## 1. Motivation

Go is excellent for engine internals — type safety, performance, and concurrency. The definition-system and hot-reload cover most iteration workflows. However, a scripting layer would address gaps that neither mechanism fills:

- **In-process reload (~10ms)** vs hot-restart (~1.2s) — critical for AI behavior tuning.
- **Sandboxed execution** — mods cannot crash the engine or access the filesystem.
- **Coroutines** — sequential multi-frame logic (patrol → wait → attack → flee) expressed naturally.
- **Lower barrier** — designers write scripts without understanding Go build tooling.

**Decision**: Deferred to post-v1. Revisit when the engine has a working editor and real users requesting scripting.

## 2. Constraints & Assumptions

- The scripting runtime MUST be pure Go (no CGo) to comply with C24.
- Scripts MUST NOT directly modify component memory. All mutations go through CommandBuffers.
- Scripts execute on the main thread within the `Update` schedule.
- Script execution time MUST be bounded by a configurable instruction limit.
- The scripting VM is a plugin — the engine core has zero dependency on any scripting runtime.

## 3. Core Invariants

- **INV-1**: Scripts cannot crash the engine. Runtime errors are caught, logged, and the offending script is disabled.
- **INV-2**: All World mutations from scripts are deferred through CommandBuffers, never immediate.
- **INV-3**: Script execution is bounded. A runaway script is suspended, not allowed to freeze the frame.
- **INV-4**: Scripts access only the engine's public API surface. Internal engine state is never exposed.
- **INV-5**: The scripting system is fully optional — removing it has zero impact on engine functionality.

## 4. Detailed Design

### 4.1 Runtime Candidates

Two runtimes are shortlisted for future evaluation. Both are pure Go (C24 compliant):

```plaintext
Candidate 1: Lua (via gopher-lua)
  Language:     Lua 5.1
  Runtime:      gopher-lua (pure Go, no CGo)
  Size:         ~200KB embedded
  Coroutines:   Native (Lua coroutines)
  Typing:       Dynamic
  Ecosystem:    Large — industry standard (WoW, Roblox, Defold, LÖVE)
  Pros:         Familiar to gamedevs, proven in production, rich tooling
  Cons:         No static typing, gopher-lua slower than LuaJIT,
                Lua 5.1 only (no goto, no integers)

Candidate 2: Tengo (via tengo)
  Language:     Tengo (Go-like syntax)
  Runtime:      tengo (pure Go, no CGo)
  Size:         ~150KB embedded
  Coroutines:   No native coroutines (workaround via closures)
  Typing:       Dynamic with Go-like feel
  Ecosystem:    Small — niche but growing
  Pros:         Go-like syntax (zero learning curve for Go devs),
                compiled bytecode (faster than gopher-lua),
                immutable strings, first-class errors
  Cons:         No coroutines, small community, limited tooling,
                less known in gamedev
```

**Evaluation criteria** (to be applied when the decision is made):

1. **Performance**: benchmark both runtimes with 1000 scripted entities per frame.
2. **Coroutine support**: critical for sequential gameplay logic (patrol, dialogue, cutscenes).
3. **Community**: documentation quality, editor support (syntax highlighting, LSP).
4. **Marshalling overhead**: cost of Go struct → script table → Go struct round-trip.
5. **Sandbox safety**: ability to restrict file/network/OS access.

### 4.2 Script-ECS Bridge (Reference Design)

The bridge API is runtime-agnostic. When a runtime is chosen, this interface is implemented:

```plaintext
ScriptRuntime (interface)
  LoadScript(source: []byte) -> (ScriptHandle, error)
  UnloadScript(handle: ScriptHandle)
  CallFunction(handle: ScriptHandle, name: string, args: ...any) -> error
  SetInstructionLimit(limit: uint64)
  ReloadScript(handle: ScriptHandle, newSource: []byte) -> error

ScriptComponent
  asset:         Handle[Script]       // asset reference to the script file
  handle:        ScriptHandle         // VM-internal reference
  enabled:       bool                 // runtime enable/disable
  state:         ScriptState          // Active | Suspended | Error
  error_message: string               // populated if state == Error
```

### 4.3 ECS API Surface (Reference Design)

Regardless of the chosen runtime, scripts interact with the ECS through this API:

```plaintext
// Entity operations
entity = ecs.spawn()
ecs.despawn(entity)
exists = ecs.is_alive(entity)

// Component operations
data     = ecs.get(entity, "ComponentName")     // read → script table
ecs.set(entity, "ComponentName", { ... })       // write → CommandBuffer
ecs.insert(entity, "ComponentName", { ... })
ecs.remove(entity, "ComponentName")
has      = ecs.has(entity, "ComponentName")

// Query iteration
for entity, comp_a, comp_b in ecs.query("CompA", "CompB") do ... end

// Events
ecs.send("EventName", { ... })
ecs.on("EventName", callback)

// Read-only access
time  = ecs.resource("Time")          // delta, elapsed, frame_count
input = ecs.service("Input")          // key/button/axis state
```

### 4.4 Why Deferred

The current stack already covers primary use cases:

```plaintext
Use Case                     Current Solution              Scripting Adds
─────────────────────────────────────────────────────────────────────────
Tweak UI layout              definition-system (JSON)      nothing
Adjust game flow             definition-system (flow)      nothing
Change entity templates      definition-system (template)  nothing
Iterate on Go system logic   hot-reload (1.2s restart)     ~10ms reload
AI behavior prototyping      hot-reload                    coroutines, faster
Cutscenes / triggers         definition-system (flow)      more flexibility
Mod support                  NOT possible                  sandboxed runtime
Designer-authored logic      NOT possible                  lower barrier
```

Scripting becomes high-priority when "NOT possible" rows become requirements.

## 5. Open Questions

- Lua vs Tengo: which runtime best fits the engine's Go-first philosophy?
- Should the engine support multiple runtimes simultaneously (plugin per runtime)?
- Is coroutine support a hard requirement, or can sequential logic use state machines?
- Should the scripting API support creating new component types at runtime?
- Should there be a visual scripting layer (node graphs) that compiles to the chosen runtime?
- How should deterministic simulation (networking-system.md) interact with scripts?

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
| 0.1.0 | 2026-03-28 | Initial draft: full Lua bridge design with ECS API, sandboxing, coroutines |
| 0.2.0 | 2026-03-28 | Deferred to post-v1. Replaced Lua-specific design with runtime-agnostic reference. Added Tengo as candidate. Minimalism strategy: rely on definition-system + hot-reload |
| — | — | Planned examples: `examples/app/` |
