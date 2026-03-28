# System Scheduling — Go Implementation

**Version:** 0.1.0
**Status:** Draft
**Layer:** go
**L1 Reference:** [system-scheduling.md](system-scheduling.md)

## Overview

Go-level design for the system scheduling subsystem. Defines the `System` interface, function-system wrappers with parameter injection, schedule construction with DAG-based ordering, single-threaded and multi-threaded executors, run conditions, and deferred command sync points.

## Related Specifications

- [system-scheduling.md](system-scheduling.md) — L1 concept specification (parent)

## Go Package

```
internal/ecs
```

All scheduling types live alongside `World`, `Entity`, and `Query` in the core ECS package. No external dependencies; multi-threaded executor uses `sync` and `golang.org/x/sync/errgroup` (justified: stdlib has no `errgroup` equivalent).

**Correction**: `errgroup` is in `golang.org/x/sync`, which is a quasi-stdlib module maintained by the Go team. If strict zero-dependency is required, a minimal internal implementation (~40 lines) can replace it.

## Type Definitions

### System Interface

```go
// System is the base interface for all executable units in a schedule.
type System interface {
    Run(world *World)
}

// SystemParam is implemented by types that can be injected into function systems.
// Each param knows how to fetch itself from the World.
type SystemParam interface {
    // Fetch initializes the param from the given World.
    Fetch(world *World)

    // Access returns the data access descriptors for conflict detection.
    Access() []AccessDescriptor
}
```

### Access Tracking

```go
// AccessMode describes whether a system reads or writes a resource.
type AccessMode uint8

const (
    AccessRead AccessMode = iota
    AccessWrite
)

// AccessDescriptor identifies a single data access by a system.
type AccessDescriptor struct {
    TypeID TypeID
    Mode   AccessMode
}
```

### Function System

```go
// FuncSystem wraps a user function and its extracted parameter metadata.
// It implements the System interface.
type FuncSystem struct {
    name       string
    fn         any                    // the original function value
    paramTypes []SystemParamFactory   // factories to create params per invocation
    access     []AccessDescriptor     // pre-computed at registration time
    conditions []RunCondition         // run conditions
    sets       []SystemSet            // set memberships
    ordering   []OrderingConstraint   // before/after edges
}

// SystemParamFactory creates a SystemParam instance for injection.
type SystemParamFactory interface {
    Create() SystemParam
    Access() []AccessDescriptor
}
```

### Built-in System Parameters

```go
// Res[T] provides shared (read-only) access to a resource.
type Res[T any] struct {
    value *T
}

// ResMut[T] provides exclusive (read-write) access to a resource.
type ResMut[T any] struct {
    value *T
}

// Local[T] holds per-system local state, not shared with other systems.
type Local[T any] struct {
    value T
}

// Commands is a system parameter that provides access to a command buffer.
type Commands struct {
    buffer *CommandBuffer
}
```

### Schedule

```go
// ScheduleName identifies a named schedule.
type ScheduleName string

// Standard schedule names.
const (
    SchedulePreStartup  ScheduleName = "PreStartup"
    ScheduleStartup     ScheduleName = "Startup"
    SchedulePostStartup ScheduleName = "PostStartup"
    ScheduleFirst       ScheduleName = "First"
    SchedulePreUpdate   ScheduleName = "PreUpdate"
    ScheduleUpdate      ScheduleName = "Update"
    SchedulePostUpdate  ScheduleName = "PostUpdate"
    ScheduleLast        ScheduleName = "Last"
    ScheduleFixedFirst  ScheduleName = "FixedFirst"
    ScheduleFixedUpdate ScheduleName = "FixedUpdate"
    ScheduleFixedLast   ScheduleName = "FixedLast"
)

// Schedule is a named, ordered collection of systems.
type Schedule struct {
    name     ScheduleName
    nodes    []SystemNode
    sets     map[SystemSet]*SystemSetConfig
    dag      *DAG
    executor Executor
    built    bool // true after Build() completes
}

// SystemNode wraps a system with its scheduling metadata.
type SystemNode struct {
    id         SystemNodeID
    system     System
    name       string
    sets       []SystemSet
    ordering   []OrderingConstraint
    conditions []RunCondition
    access     []AccessDescriptor
}

// SystemNodeID is a unique index within a schedule.
type SystemNodeID uint32
```

### System Sets

```go
// SystemSet is a named group identifier for organizing systems.
type SystemSet string

// SystemSetConfig holds configuration applied to all members of a set.
type SystemSetConfig struct {
    ordering   []OrderingConstraint
    conditions []RunCondition
}

// OrderingConstraint defines a before/after relationship.
type OrderingConstraint struct {
    kind   OrderingKind
    target SystemSet
}

// OrderingKind is the type of ordering constraint.
type OrderingKind uint8

const (
    OrderBefore OrderingKind = iota
    OrderAfter
)
```

### DAG

```go
// DAG is a directed acyclic graph of system nodes for execution ordering.
type DAG struct {
    nodeCount int
    adj       [][]SystemNodeID // adjacency list: adj[from] = [to, ...]
    inDegree  []int            // in-degree per node
    sorted    []SystemNodeID   // topologically sorted order (populated after Build)
}
```

### Executor

```go
// Executor defines the interface for running a built schedule.
type Executor interface {
    Run(dag *DAG, nodes []SystemNode, world *World) error
}

// SequentialExecutor runs systems one at a time in topological order.
type SequentialExecutor struct{}

// ParallelExecutor runs non-conflicting systems concurrently.
type ParallelExecutor struct {
    maxWorkers int
}
```

### Run Conditions

```go
// RunCondition is a predicate evaluated before a system runs.
// If it returns false, the system is skipped.
type RunCondition func(world *World) bool
```

## Key Methods

### Schedule Building

```
func NewSchedule(name ScheduleName) *Schedule

func (s *Schedule) AddSystem(system System) *SystemNodeBuilder
func (s *Schedule) ConfigureSet(set SystemSet) *SystemSetBuilder
func (s *Schedule) AddApplyDeferred() // insert explicit sync point

// Build resolves ordering, builds DAG, detects cycles.
// Returns error with cycle details if Tarjan's SCC finds cycles.
func (s *Schedule) Build() error

// Run executes the schedule: evaluate conditions, run systems, apply deferred.
func (s *Schedule) Run(world *World) error
```

### DAG Construction

```
func NewDAG(nodeCount int) *DAG

// AddEdge adds a directed edge (from must run before to).
func (d *DAG) AddEdge(from, to SystemNodeID)

// Build performs:
//   1. Tarjan's SCC for cycle detection — O(V+E)
//   2. Kahn's algorithm for topological sort — O(V+E)
//   3. Ambiguity detection: conflicting access without ordering edge
// Returns error with cycle members if cycles detected.
func (d *DAG) Build() error

// TopologicalOrder returns the sorted execution order.
func (d *DAG) TopologicalOrder() []SystemNodeID
```

### Tarjan's SCC (Cycle Detection)

```
// tarjanSCC runs Tarjan's strongly connected components algorithm.
// Returns all SCCs with more than one node (cycles).
// Time complexity: O(V+E), space: O(V).
func tarjanSCC(adj [][]SystemNodeID, nodeCount int) [][]SystemNodeID
```

### Executor Implementations

```
// SequentialExecutor.Run iterates topological order, evaluates conditions,
// runs systems, applies deferred commands at sync points.
func (e *SequentialExecutor) Run(dag *DAG, nodes []SystemNode, world *World) error

// ParallelExecutor.Run maintains a ready-queue of systems whose
// dependencies are satisfied. For each ready system:
//   1. Check access conflicts with currently-running systems.
//   2. If no conflict, dispatch to a goroutine via errgroup.
//   3. On completion, decrement in-degree of dependents.
//   4. Apply deferred at sync points (barrier — wait for all running systems).
func (e *ParallelExecutor) Run(dag *DAG, nodes []SystemNode, world *World) error
```

### System Parameter Injection

```
// IntoSystem converts a user function into a FuncSystem by reflecting
// on parameter types and building SystemParamFactory instances.
// Panics at registration time (not at runtime) if params are invalid.
func IntoSystem(fn any, name string) *FuncSystem

// FuncSystem.Run fetches all params, invokes the function, returns.
func (fs *FuncSystem) Run(world *World)
```

### Chaining Utility

```
// Chain creates sequential ordering constraints for a list of systems.
// Equivalent to: systems[0].Before(systems[1]).Before(systems[2])...
func Chain(systems ...System) []OrderingConstraint
```

## Performance Strategy

- **DAG built once** at schedule initialization (`Build()`), not per frame.
- **Access descriptors pre-computed** at system registration, stored on `SystemNode`.
- **Parallel executor** uses a fixed goroutine pool (default: `runtime.GOMAXPROCS`), not one goroutine per system.
- **Condition evaluation** is cheap (function pointer call), no allocation.
- **Command buffer apply** at sync points reuses pre-allocated slices (see command-system-go).
- **SystemParam.Fetch** performs pointer lookups into World internals — no map lookups on hot path after initial resolution.

## Error Handling

| Condition | Behavior |
| :--- | :--- |
| Cycle in DAG | `Build()` returns `ErrScheduleCycle` with involved system names |
| Ambiguous access (warning) | Logged via `slog.Warn`, not a hard error |
| System panics during Run | Recovered by executor, wrapped as `ErrSystemPanic`, schedule halts |
| Missing resource for Res[T] | Panic at `Fetch` time with descriptive message (programming error) |
| Duplicate system name | `AddSystem` returns error |

```go
var (
    ErrScheduleCycle  = errors.New("ecs: cycle detected in schedule DAG")
    ErrSystemPanic    = errors.New("ecs: system panicked during execution")
    ErrDuplicateSystem = errors.New("ecs: duplicate system name in schedule")
)
```

## Testing Strategy

- **Unit tests**: DAG construction, topological sort, cycle detection with known graphs.
- **Property tests**: Random DAG generation, verify topological order respects all edges.
- **Integration tests**: Schedule with multiple systems, verify execution order and deferred command application.
- **Benchmarks**: `BenchmarkDAGBuild` (1000 nodes), `BenchmarkScheduleRun` (sequential vs parallel), `BenchmarkParamFetch`.
- **Race detection**: All parallel executor tests run with `-race` flag.

## Open Questions

- Should `IntoSystem` use code generation instead of runtime reflection for param injection?
- How to surface ambiguity warnings to the user in a structured way (beyond slog)?
- Should the parallel executor support work-stealing or is a simple ready-queue sufficient?
- Exact strategy for `errgroup` vs internal goroutine pool for zero-dependency compliance.

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-26 | Initial L2 draft |
