# Contributing

## 📂 Repository Structure (This Project)

```plaintext
ecs-engine/                 # Root of the engine project
├── cmd/                    # CLI tools and standalone executables (Go 1.26.1)
│   ├── ecs-cli/            # Scaffolding and project management tool
│   └── editor/             # Entry point for the GUI-based engine editor
├── examples/               # Validating implementations (Required for C29)
│   ├── basic/              # Minimal entity/component/system workflow
│   ├── rendering/          # Render-graph, sprite batching, and text-rendering
│   ├── physics/            # Collision queries and kinematic movement demos
│   └── networking/         # Snapshot sync and client-side prediction tests
├── internal/               # Core engine implementation (Private)
│   ├── ecs/                # Central entity-component-system kernel
│   │   ├── archetype/      # Table-based contiguous memory layout
│   │   ├── component/      # Registries, bundles, and sparse-set storage
│   │   ├── entity/         # Generational ID allocation and recycling
│   │   ├── query/          # Archetype filters and parallel iterators
│   │   └── world/          # Main data store and implementation coordinator
│   ├── app/                # Application framework and plugin orchestrator
│   ├── asset/              # Asynchronous asset server and hot-reloader
│   ├── render/             # Multithreaded render-graph and GPU extraction
│   ├── scheduler/          # Parallel DAG execution and system scheduling
│   ├── events/             # Observables and reactive event bus
│   ├── hierarchy/          # Transform propagation and child management
│   ├── input/              # Unified keyboard/mouse/gamepad abstraction
│   ├── state/              # Hierarchical Finite State Machines (HFSM)
│   ├── time/               # Fixed-timestep loop and virtual timers
│   └── type/               # Runtime type registry and reflection bridge
└── pkg/                    # Exportable utility packages (Reusable)
    ├── math/               # SIMD-optimized vectors, matrices, and geometry
    ├── diagnostics/        # Profiling, logging, and error taxonomy
    └── platform/           # OS-level windowing and platform abstractions
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
