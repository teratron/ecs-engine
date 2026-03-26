# Workspace Specifications Registry

**Version:** 2.5.0
**Status:** Active

## Overview

Local registry of specifications for this workspace. Organized by priority batch (P1–P6).

## P1 — ECS Core

| File | Description | Status | Layer | Version |
| :--- | :--- | :--- | :--- | :--- |
| [world-system.md](specifications/world-system.md) | Central data store: entities, components, resources, change tracking | Draft | concept | 0.1.0 |
| [world-system-go.md](specifications/world-system-go.md) | World Go implementation: World struct, DeferredWorld, ResourceMap, archetypes, tables | Draft | go | 0.1.0 |
| [entity-system.md](specifications/entity-system.md) | Entity lifecycle, generational IDs, allocation, disabling | Draft | concept | 0.1.0 |
| [entity-system-go.md](specifications/entity-system-go.md) | Entity Go implementation: EntityID, Entity, EntityAllocator, EntitySet, EntityMap | Draft | go | 0.1.0 |
| [component-system.md](specifications/component-system.md) | Component registration, storage strategies, hooks, required components | Draft | concept | 0.1.0 |
| [component-system-go.md](specifications/component-system-go.md) | Component Go implementation: ComponentID, ComponentRegistry, hooks, bundles, storage types | Draft | go | 0.1.0 |
| [query-system.md](specifications/query-system.md) | Data access: queries, filters, iteration, access tracking | Draft | concept | 0.1.0 |
| [query-system-go.md](specifications/query-system-go.md) | Query Go implementation: QueryState, filters, Access, ParIter, multi-arity generics | Draft | go | 0.1.0 |
| [system-scheduling.md](specifications/system-scheduling.md) | System execution, DAG scheduling, parallel executor, system sets | Draft | concept | 0.1.0 |
| [system-scheduling-go.md](specifications/system-scheduling-go.md) | Go impl: System interface, DAG scheduler, executors, run conditions | Draft | go | 0.1.0 |
| [command-system.md](specifications/command-system.md) | Deferred mutations, command buffers, apply points | Draft | concept | 0.1.0 |
| [command-system-go.md](specifications/command-system-go.md) | Go impl: Command interface, CommandBuffer, entity reservation, flush | Draft | go | 0.1.0 |
| [event-system.md](specifications/event-system.md) | Events, messages, observers, reactive triggers | Draft | concept | 0.1.0 |
| [event-system-go.md](specifications/event-system-go.md) | Go impl: EventBus, MessageChannel, Observers, entity event bubbling | Draft | go | 0.1.0 |
| [type-registry.md](specifications/type-registry.md) | Runtime introspection, field metadata, dynamic type mapping | Draft | concept | 0.1.0 |
| [type-registry-go.md](specifications/type-registry-go.md) | Go impl: TypeRegistry, FieldInfo, DynamicObject, serialization hooks | Draft | go | 0.1.0 |

## P2 — Framework

| File | Description | Status | Layer | Version |
| :--- | :--- | :--- | :--- | :--- |
| [hierarchy-system.md](specifications/hierarchy-system.md) | Parent-child relationships, transform propagation, traversal | Draft | concept | 0.1.0 |
| [hierarchy-system-go.md](specifications/hierarchy-system-go.md) | Go impl: ChildOf, Children, Transform, GlobalTransform, propagation | Draft | go | 0.1.0 |
| [time-system.md](specifications/time-system.md) | Real/virtual/fixed time, timers, fixed timestep loop | Draft | concept | 0.1.0 |
| [time-system-go.md](specifications/time-system-go.md) | Go impl: gametime package, Time/RealTime/VirtualTime/FixedTime, Timer, Stopwatch | Draft | go | 0.1.0 |
| [input-system.md](specifications/input-system.md) | Keyboard, mouse, gamepad, touch; polling + events; picking | Draft | concept | 0.1.0 |
| [input-system-go.md](specifications/input-system-go.md) | Go impl: ButtonInput[T], AxisInput[T], KeyCode, MouseButton, GamepadButton | Draft | go | 0.1.0 |
| [state-system.md](specifications/state-system.md) | Hierarchical state machines, transitions, computed states | Draft | concept | 0.1.0 |
| [state-system-go.md](specifications/state-system-go.md) | Go impl: State[S], NextState[S], SubState, ComputedState, DespawnOnExit | Draft | go | 0.1.0 |
| [change-detection.md](specifications/change-detection.md) | Tick-based change tracking, Added/Changed filters, Ref/Mut wrappers | Draft | concept | 0.1.0 |
| [change-detection-go.md](specifications/change-detection-go.md) | Go impl: Tick, ComponentTicks, Ref[T], Mut[T], RemovedComponents[T] | Draft | go | 0.1.0 |
| [app-framework.md](specifications/app-framework.md) | App builder, plugins, plugin groups, sub-apps, game loop | Draft | concept | 0.1.0 |
| [app-framework-go.md](specifications/app-framework-go.md) | Go impl: App, Plugin, PluginGroup, SubApp, RunMode, DefaultPlugins | Draft | go | 0.1.0 |

## P3 — Assets & Math

| File | Description | Status | Layer | Version |
| :--- | :--- | :--- | :--- | :--- |
| [task-system.md](specifications/task-system.md) | Parallelism: worker pools, scoped tasks, parallel iteration | Draft | concept | 0.1.0 |
| [asset-system.md](specifications/asset-system.md) | Asset server, loaders, handles, hot-reload, IO abstraction | Draft | concept | 0.1.0 |
| [scene-system.md](specifications/scene-system.md) | Scene serialization, dynamic scenes, spawning, entity remapping | Draft | concept | 0.1.0 |
| [math-system.md](specifications/math-system.md) | Vectors, matrices, quaternions, colors, geometric primitives | Draft | concept | 0.1.0 |

## P4 — Render Pipeline

| File | Description | Status | Layer | Version |
| :--- | :--- | :--- | :--- | :--- |
| [render-core.md](specifications/render-core.md) | Render graph, extract pattern, render world, backend abstraction | Draft | concept | 0.1.0 |
| [mesh-and-image.md](specifications/mesh-and-image.md) | Mesh assets, vertex layout, image/texture, texture atlases | Draft | concept | 0.1.0 |
| [materials-and-lighting.md](specifications/materials-and-lighting.md) | Material system, PBR, light types, shadows, environment maps | Draft | concept | 0.1.0 |
| [camera-and-visibility.md](specifications/camera-and-visibility.md) | Camera, projections, visibility hierarchy, frustum culling | Draft | concept | 0.1.0 |
| [post-processing.md](specifications/post-processing.md) | Post-process effects, anti-aliasing, tonemapping, bloom | Draft | concept | 0.1.0 |

## P5 — Content Systems

| File | Description | Status | Layer | Version |
| :--- | :--- | :--- | :--- | :--- |
| [audio-system.md](specifications/audio-system.md) | Audio playback, spatial audio, backend abstraction | Draft | concept | 0.1.0 |
| [asset-formats.md](specifications/asset-formats.md) | Asset loaders: glTF, images, audio codecs, scene files | Draft | concept | 0.1.0 |
| [2d-rendering.md](specifications/2d-rendering.md) | Sprites, texture slicing, text rendering, 2D pipeline | Draft | concept | 0.1.0 |
| [animation-system.md](specifications/animation-system.md) | Animation graphs, clips, curves, skeletal animation, morph targets | Draft | concept | 0.1.0 |

## P6 — UI & Tools

| File | Description | Status | Layer | Version |
| :--- | :--- | :--- | :--- | :--- |
| [definition-system.md](specifications/definition-system.md) | JSON declarative layer: UI, scenes, flows, templates — data-driven bridge | Draft | concept | 0.1.0 |
| [window-system.md](specifications/window-system.md) | Window management, multi-window, platform abstraction | Draft | concept | 0.1.0 |
| [diagnostic-system.md](specifications/diagnostic-system.md) | Diagnostics, profiling, gizmos, error codes, debug overlay | Draft | concept | 0.1.0 |
| [ui-system.md](specifications/ui-system.md) | Layout engine, interaction, text, widgets, styling | Draft | concept | 0.1.0 |
| [build-tooling.md](specifications/build-tooling.md) | CI pipeline, golden file testing, benchmarks, migration/release doc formats | Draft | concept | 0.1.0 |
| [examples-framework.md](specifications/examples-framework.md) | Examples directory structure, conventions, and lifecycle | Draft | concept | 0.3.0 |

## Meta Information

- **Maintainer**: Core Team
- **Last Updated**: 2026-03-26
