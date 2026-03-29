# Contributing

## 📂 Repository Structure (This Project)

```plaintext
ecs-engine/                 # Root of the engine project
├── cmd/                    # CLI tools and standalone executables (Go 1.26.1)
│   └── cli/                # Scaffolding and project management tool
│
├── examples/               # Validating implementations (required by C26/C29)
│   ├── ecs/                # Entity/component/system core patterns
│   ├── world/              # Resources, events, hierarchy, change detection
│   ├── app/                # Plugin system, schedules, state machines
│   ├── 2d/                 # 2D rendering pipeline
│   ├── 3d/                 # 3D rendering pipeline
│   ├── physics/            # Collision, rigid bodies, character controller
│   ├── audio/              # Spatial and non-spatial audio
│   ├── ui/                 # Layout engine and widgets
│   ├── networking/         # Snapshot sync, prediction, lockstep
│   ├── asset/              # Asset loading, hot-reload, VFS
│   ├── diagnostic/         # Profiling, gizmos, debug overlay
│   └── stress_test/        # Performance benchmarks
│
├── internal/               # Core engine implementation (Private)
│   │
│   ├── ecs/                # Central entity-component-system kernel
│   │   ├── archetype/      # Table-based contiguous memory layout
│   │   ├── component/      # Registries, bundles, and sparse-set storage
│   │   ├── entity/         # Generational ID allocation and recycling
│   │   ├── query/          # Archetype filters and parallel iterators
│   │   ├── command/        # CommandBuffer, deferred mutations, EntityCommands
│   │   ├── change/         # Tick-based change detection, Ref[T]/Mut[T], ClearTrackers
│   │   ├── world/          # Main data store and implementation coordinator
│   │   └── scheduler/      # Parallel DAG execution and system scheduling
│   │
│   ├── app/                # Application framework and plugin orchestrator
│   ├── asset/              # Asynchronous asset server and hot-reloader
│   ├── hotreload/          # Engine hot-swap orchestrator, state snapshots, and VFS watchers
│   ├── scene/              # DynamicScene, StaticScene, entity remapping, prefabs
│   ├── hierarchy/          # ChildOf, Children, transform propagation
│   ├── input/              # ButtonInput[T], AxisInput, action mapping, picking
│   ├── state/              # State[S], NextState[S], SubState, ComputedState
│   ├── time/               # Fixed-timestep loop and virtual timers
│   ├── events/             # Observables and reactive event bus
│   ├── registry/           # Runtime type registry, reflection bridge, and metadata
│   ├── definition/         # JSON declarative layer, interpreters, hot-reload binding
│   │
│   ├── render/             # Render SubApp, render graph, extract pattern
│   │   ├── core/           # RenderGraph, RenderPass, backend abstraction, RID
│   │   ├── mesh/           # Mesh assets, vertex layout, skinning, morph targets
│   │   ├── material/       # PBR materials, shaders, pipeline specialization
│   │   ├── camera/         # Camera, projections, frustum, visibility culling
│   │   ├── light/          # Light types, shadow maps, IBL, irradiance volumes
│   │   ├── pipeline2d/     # Sprite batching, TextureAtlas, Text2D
│   │   ├── pipeline3d/     # 3D render phases, instancing, deferred
│   │   └── postprocess/    # Bloom, tonemapping, AA, DOF, SSAO
│   │
│   ├── physics/            # Physics SubApp, backend abstraction
│   │   ├── server/         # PhysicsServer, command queue, sync/writeback phases
│   │   ├── body/           # RigidBody component, BodyType, MassProperties
│   │   ├── collider/       # ColliderShape, CollisionGroups, compound colliders
│   │   ├── query/          # RayCast, ShapeCast, OverlapShape, ContactsBetween
│   │   ├── joints/         # Revolute, Prismatic, Spherical, Distance, Generic
│   │   ├── events/         # CollisionStarted/Ended, TriggerEntered/Exited
│   │   ├── material/       # PhysicsMaterial, CombineRule, surface tags
│   │   └── character/      # CharacterController, iterative sweep, step-up
│   │
│   ├── audio/              # Audio SubApp, backend abstraction
│   │   ├── server/         # AudioServer, bus graph, AudioDriver interface
│   │   ├── source/         # AudioPlayer, PlaybackSettings, SpatialAudioSink
│   │   └── effect/         # AudioEffect factory/instance split, bus effects
│   │
│   ├── ui/                 # UI system (built on engine's own ECS — dogfooding)
│   │   ├── layout/         # Flexbox engine (Taffy-equivalent), dirty flags
│   │   ├── style/          # Style component, Val, UiRect, BackgroundColor
│   │   ├── text/           # Font loading, glyph atlas, Text2D pipeline
│   │   ├── interaction/    # Hover/press state, MouseFilter, focus management
│   │   └── widget/         # Node, ImageNode, ScrollView convenience bundles
│   │
│   ├── window/             # OS window management (entity-based)
│   │   ├── component/      # Window component, WindowMode, CursorOptions
│   │   └── backend/        # WindowBackend interface, platform implementations
│   │
│   └── network/            # Networking — primitives only, no specific model
│       ├── transport/      # UDP, channels, reliability layer, connection lifecycle
│       ├── replication/    # Component markers, EntityMap, visibility, delta compression
│       ├── snapshot/       # SnapshotData, ring buffer, delta compression
│       ├── prediction/     # InputBuffer, RollbackCoordinator, misprediction smoothing
│       ├── interpolation/  # SnapshotBuffer, per-component lerp, adaptive delay
│       ├── lockstep/       # LockstepScheduler, input delay, desync detection
│       └── rpc/            # RpcRegistry, RpcSender/Receiver, rate limiting
│
└── pkg/                    # Exportable utility packages (Reusable)
    ├── math/               # SIMD-optimized vectors, matrices, and geometry
    ├── platform/           # Cross-platform capability negotiation
    │   ├── profile/        # PlatformProfile, PlatformCaps bitfield
    │   └── backend/        # SocketBackend, RenderBackend, AudioDriver interfaces
    ├── diagnostic/         # Observability layer
    │   ├── store/          # DiagnosticsStore, Diagnostic, rolling history
    │   ├── gizmo/          # Immediate-mode debug drawing, GizmoConfigStore
    │   ├── profiling/      # Span API, Tracy/pprof/chrome exporters
    │   └── error/          # EngineError, E-series codes, localization registry
    ├── codegen/            # ecs-gen: code generation for components/queries
    ├── editor/             # Plugin interfaces: InspectorPlugin, GizmoPlugin, etc.
    └── protocol/           # IPC wire format: hot-reload and diagnostic messages
```

## 🏗️ Game Project Structure (User Project)

```plaintext
my-game/                    # Typical project using the ecs-engine
├── assets/                 # Raw assets (glTF, images, audio, scenes)
├── config/                 # Declarative definitions (UI, logic flows, templates)
├── src/                    # Game-specific systems and components
│   └── main.go             # App builder, plugin registration and game loop
└── go.mod                  # Go dependency management
```

## 🗝️ Core Architectural Decisions

### ECS Kernel Isolation

- `internal/ecs/command/` and `internal/ecs/change/`: Explicitly isolated sub-packages for CommandBuffer and Change Detection (Ref/Mut) logic.
- `internal/ecs/scheduler/`: Integrated directly into the ECS core, utilizing access descriptors from queries and commands for DAG construction.

### SubApp Modularity

- `internal/render/`: Decomposed into 8 specialized sub-packages to manage the high complexity of the render graph and pipeline specialization (P4/P5 specs).
- `internal/physics/`: Mirrored 8-package structure (P8 specs) with `server/` isolating backend implementations from pure data components.
- `internal/network/`: Layer-based sub-packages (P7 specs) ensuring that transport remains unaware of high-level synchronization logic (prediction/lockstep).

### Exported Packages (pkg/)

- `pkg/platform/`: Limited to interfaces and capability negotiation; physical implementations remain in `internal/window/backend/`.
- `pkg/diagnostic/`: Atomic split into Store, Gizmo, Profiling, and Error packages to minimize cross-dependency.
- `pkg/editor/` & `pkg/protocol/`: Stable architectural boundaries for communication with the `ecs-editor` repository.

### Multi-Repo Architecture

- **Repo Split**: The GUI editor resides in a separate `ecs-editor` repository to ensure the engine's public API is properly "dogfooded" and to isolate internal engine dependencies.
- **IPC Protocol**: `pkg/protocol/` defines the shared newline-delimited JSON wire format for hot-reload and live diagnostics, maintained in the engine repository to ensure state-synchronization accuracy.

### SDD Integration

- `.design/`: The singular source of truth for architectural specifications (Magic SDD), integrated directly into the repository for visibility and build-time validation.
