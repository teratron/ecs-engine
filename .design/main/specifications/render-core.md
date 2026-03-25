# Render Core
**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview
The render core defines how frames are produced. A dedicated render SubApp owns its own World, receives extracted data each frame, and executes a render graph — a DAG of passes that declare their resource dependencies. A pluggable backend interface allows multiple GPU APIs without changing upper layers.

## Related Specifications
- [App Framework](app-framework.md)
- [Mesh & Image](mesh-and-image.md)
- [Materials & Lighting](materials-and-lighting.md)

## 1. Motivation
Rendering must be decoupled from gameplay so the main World never touches GPU state directly. A graph-based approach lets the engine reorder, merge, or cull passes automatically. A backend interface future-proofs the engine against API churn.

## 2. Constraints & Assumptions
- The render SubApp runs **after** the main app finishes its Update schedule for the current frame.
- Extract functions are the **only** bridge between the main World and the render World — no shared mutable state.
- All GPU resource handles are opaque to code outside the render SubApp.
- The render graph must be rebuildable each frame (dynamic pass insertion is allowed).

## 3. Core Invariants
1. Every render pass must declare its inputs and outputs before execution.
2. The graph must be acyclic; a cycle is a hard error at graph-build time.
3. GPU resources referenced by a pass must be alive for the duration of that pass.
4. Extract runs exactly once per frame, before any render pass executes.
5. Backend implementations must satisfy the full `RenderBackend` interface — partial implementations are not permitted.

## 4. Detailed Design

### 4.1 Render SubApp & Extract Pattern
The render SubApp maintains its own World, schedule, and resource table. Each frame:

```plaintext
Main App Update
      │
      ▼
  Extract Phase ── copies relevant data ──► Render World
      │
      ▼
Render Schedule (Render World)
      │
      ▼
  Present
```

Extract functions are registered per plugin. Each receives a read-only view of the main World and a mutable reference to the render World. Data is copied, not shared.

### 4.2 Render Graph
The graph is a directed acyclic graph of `RenderPass` nodes. Each node declares:
- **Inputs**: texture handles or buffer handles it reads.
- **Outputs**: texture handles or buffer handles it writes.
- **Phase**: which render phase it belongs to (or `None` for utility passes).

The graph compiler topologically sorts passes, inserts barriers, and allocates transient resources.

### 4.3 Render Phases
Phases execute in a fixed order within the main scene pass:

| Phase | Sort Strategy | Notes |
| :--- | :--- | :--- |
| Opaque | Front-to-back by depth | Minimises overdraw |
| AlphaMask | Front-to-back by depth | Discard fragments below threshold |
| Transparent | Back-to-front by depth | Requires correct blend order |
| UI | Submission order | Screen-space overlay |

Each phase owns a list of draw functions. Draw functions are batched: items sharing the same pipeline, bind group, and vertex buffer are merged into a single draw call where possible.

### 4.4 Backend Abstraction
`RenderBackend` is an interface with the following surface (pseudo-code):

```plaintext
RenderBackend
  CreateBuffer(descriptor) -> BufferHandle
  CreateTexture(descriptor) -> TextureHandle
  CreatePipeline(descriptor) -> PipelineHandle
  CreateBindGroup(descriptor) -> BindGroupHandle
  BeginRenderPass(descriptor)
  Draw(pipeline, bind_groups, vertex_buffers, index_buffer, instances)
  EndRenderPass()
  Submit()
  Present()
```

Concrete backends (OpenGL, Vulkan, WebGPU, software rasteriser) implement this interface. The active backend is selected at app initialisation and cannot change at runtime.

### 4.5 GPU Resource Management
Buffers, textures, pipelines, and bind groups are stored as resources in the render World. A resource tracker performs reference counting; resources with zero references at the end of a frame are queued for deferred deletion (next frame) to avoid destroying in-flight GPU objects.

### 4.6 Pipeline Specialization
A pipeline key is derived from (shader_id, material_properties, vertex_layout, render_phase). The engine maintains a pipeline cache keyed on this tuple. Cache misses trigger asynchronous compilation; a fallback pipeline is used until the specialised variant is ready.

### 4.7 Draw Functions
A draw function is a registered callable per render phase. During the render schedule, the engine collects visible entities for each phase, sorts them, batches by pipeline + bind group, and invokes the corresponding draw function with the batched data.

## 5. Open Questions
1. Should transient texture allocation use a pool or per-frame linear allocator?
2. How should async pipeline compilation failures be surfaced to the developer?
3. What is the maximum number of render passes before performance degrades on target hardware?

## Document History
| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft from architecture analysis |
