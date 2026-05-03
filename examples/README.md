# Engine Examples

This directory contains examples categorized by engine subsystem. Currently, the project is in the **Specification Phase (Phase 0)**, and most examples are placeholders representing planned validation targets.

## Example Categories

| Category | Description | Status |
| :--- | :--- | :--- |
| [**ECS**](ecs/README.md) | Core Entity-Component-System patterns | Placeholder (P1) |
| [**World**](world/README.md) | Resources, Events, and Hierarchy | Placeholder (P1/2) |
| [**App**](app/README.md) | Application lifecycle and Plugins | Placeholder (P1/2) |
| [**Diagnostic**](diagnostic/README.md) | Profiling and debug visualization | Placeholder (P2) |
| [**Asset**](asset/README.md) | Asynchronous loading and handles | Placeholder (P2) |
| [**2D**](2d/README.md) | 2D rendering and sprites | Placeholder (P3) |
| [**3D**](3d/README.md) | 3D meshes and lighting | Placeholder (P3) |
| [**UI**](ui/README.md) | User interface and layouts | Placeholder (P3) |
| [**Physics**](physics/README.md) | Rigid body dynamics and collisions | Placeholder (P3/4) |
| [**Audio**](audio/README.md) | 2D/3D audio and spatialization | Placeholder (P2/3) |
| [**Networking**](networking/README.md) | Replication and synchronization | Placeholder (P4+) |
| [**Stress Test**](stress_test/README.md) | Performance and scalability benchmarks | Placeholder (P2/3) |

## Staged Rollout

Examples are introduced following the engine's implementation phases:

1. **Phase 1**: Core ECS validation (`ecs/`, `world/`, `app/`).
2. **Phase 2**: Subsystem foundations (`asset/`, `diagnostic/`, `audio/`).
3. **Phase 3**: Visual and Physical modules (`2d/`, `3d/`, `ui/`, `physics/`).
4. **Phase 4+**: Complex systems (`networking/`).

For more details on the example framework, see the [Examples Framework Specification](../../.design/main/specifications/l1-examples-framework.md).
