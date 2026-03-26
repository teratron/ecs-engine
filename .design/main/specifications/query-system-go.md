# Query System — Go Implementation

**Version:** 0.1.0
**Status:** Draft
**Layer:** go
**L1 Reference:** [query-system.md](query-system.md)

## Overview

This specification defines the Go implementation of the query system. Queries are the primary way systems access entity data. Due to Go's lack of variadic generics, multi-component queries require explicit arity types (`Query1`, `Query2`, `Query3`, etc.). Each query tracks its read/write access for schedule-time conflict detection and caches matched archetypes for efficient iteration.

## Go Package

```
internal/ecs/
```

All types in this spec belong to package `ecs`.

## Type Definitions

### Access

```go
// Access describes the component read/write requirements of a query or system.
// Used by the scheduler to detect conflicts and enable parallelism.
type Access struct {
    ReadSet  []ComponentID  // components read (immutable access)
    WriteSet []ComponentID  // components written (mutable access)
}

// IsDisjoint reports whether two Access sets have no conflicting overlap.
// Conflict = one writes a component that the other reads or writes.
func (a Access) IsDisjoint(other Access) bool

// Merge combines two Access values into one (union of sets).
func (a Access) Merge(other Access) Access
```

### QueryFilter

```go
// QueryFilter is an interface for filtering which archetypes/entities
// a query matches.
type QueryFilter interface {
    // Matches reports whether the given archetype passes this filter.
    Matches(archetype *Archetype) bool
}

// TickFilter extends QueryFilter with tick-based per-entity filtering.
// Used by Changed and Added filters that need per-row tick comparison.
type TickFilter interface {
    QueryFilter
    // MatchesEntity reports whether a specific entity row passes
    // the tick-based condition.
    MatchesEntity(column *Column, row int, lastRun Tick, thisTick Tick) bool
}
```

### Concrete Filters

```go
// With filters for entities that have component T.
// T is not fetched — only used for archetype matching.
type With[T any] struct{}

func (f With[T]) Matches(archetype *Archetype) bool
// Pseudo-code: return archetype.HasComponent(componentIDOf[T]())

// Without filters for entities that do NOT have component T.
type Without[T any] struct{}

func (f Without[T]) Matches(archetype *Archetype) bool
// Pseudo-code: return !archetype.HasComponent(componentIDOf[T]())

// Changed filters for entities where component T was mutated since the
// system's last run. Archetype-level: matches if archetype has T.
// Entity-level: compares the component's change tick against lastRun tick.
type Changed[T any] struct{}

func (f Changed[T]) Matches(archetype *Archetype) bool
func (f Changed[T]) MatchesEntity(column *Column, row int, lastRun Tick, thisTick Tick) bool
// Pseudo-code: return column.changeTicks[row] > lastRun

// Added filters for entities where component T was added since the
// system's last run.
type Added[T any] struct{}

func (f Added[T]) Matches(archetype *Archetype) bool
func (f Added[T]) MatchesEntity(column *Column, row int, lastRun Tick, thisTick Tick) bool
// Pseudo-code: return column.addedTicks[row] > lastRun
```

### QueryState (Single Component)

```go
// QueryState1 is the cached state for a single-component query.
// It tracks which archetypes match and caches column pointers.
type QueryState1[T any] struct {
    componentID      ComponentID
    access           Access
    filters          []QueryFilter
    matchedArchetypes []matchedArchetype
    archetypeGen     uint32  // archetype generation at last cache update
    lastRunTick      Tick    // tick when this query's system last ran
}

type matchedArchetype struct {
    archetypeID  ArchetypeID
    tableID      TableID
    columnIndex  int  // index of the target component column in the table
}

// NewQueryState1 creates a new single-component query state.
// Registers the component type if needed.
func NewQueryState1[T any](world *World, filters ...QueryFilter) *QueryState1[T]
```

### Multi-Component Query Types

Due to Go's lack of variadic generics, multi-component queries use separate types per arity:

```go
// QueryState2 is the cached state for a two-component query.
type QueryState2[A, B any] struct {
    componentIDs     [2]ComponentID
    access           Access
    filters          []QueryFilter
    matchedArchetypes []matchedArchetype2
    archetypeGen     uint32
    lastRunTick      Tick
}

// QueryState3 is the cached state for a three-component query.
type QueryState3[A, B, C any] struct {
    componentIDs     [3]ComponentID
    access           Access
    filters          []QueryFilter
    matchedArchetypes []matchedArchetype3
    archetypeGen     uint32
    lastRunTick      Tick
}

// QueryState4 is the cached state for a four-component query.
type QueryState4[A, B, C, D any] struct {
    componentIDs     [4]ComponentID
    access           Access
    filters          []QueryFilter
    matchedArchetypes []matchedArchetype4
    archetypeGen     uint32
    lastRunTick      Tick
}

// Additional arity types up to QueryState8 follow the same pattern.
// Beyond 8, users should restructure their data or use multiple queries.
```

### Query Mutable Access Markers

```go
// Mut is a marker type wrapping a component type to indicate mutable access.
// Used as a type parameter: QueryState2[Position, Mut[Velocity]]
type Mut[T any] struct{}
```

When `Mut[T]` is used as a type parameter, the query registers T in the WriteSet instead of ReadSet. The iteration callback receives `*T` (same pointer type) but the component's change tick is updated.

## Key Methods

### Iteration (Callback-Based)

```go
// Iter iterates all matching entities sequentially.
// Updates archetype cache if new archetypes have been created.
//
// Pseudo-code:
//   updateCache(world)
//   for each matched archetype:
//     table = world.tables.Get(archetype.tableID)
//     column = table.columns[componentID]
//     for row in 0..table.len:
//       if tick filters present and !matchesEntity(row): skip
//       entity = table.entities[row]
//       ptr = (*T)(unsafe.Pointer(&column.data[row * elemSize]))
//       fn(entity, ptr)
func (q *QueryState1[T]) Iter(world *World, fn func(Entity, *T))

// Iter for two-component query.
func (q *QueryState2[A, B]) Iter(world *World, fn func(Entity, *A, *B))

// Iter for three-component query.
func (q *QueryState3[A, B, C]) Iter(world *World, fn func(Entity, *A, *B, *C))
```

### Single Entity Lookup

```go
// Get retrieves the component for a specific entity.
// Returns an error if the entity is dead or does not match the query.
func (q *QueryState1[T]) Get(world *World, entity Entity) (*T, error)

// Get for two-component query.
func (q *QueryState2[A, B]) Get(world *World, entity Entity) (*A, *B, error)
```

### Parallel Iteration

```go
// ParIter divides work across goroutines by splitting matched archetypes
// into chunks. Each goroutine processes a chunk independently.
//
// Pseudo-code:
//   updateCache(world)
//   collect all (archetype, table) pairs
//   divide rows into chunks of size >= minChunkSize
//   launch one goroutine per chunk via errgroup
//   each goroutine iterates its row range and calls fn
//
// Safety: fn must not perform structural World changes. The query's Access
// must not conflict with any concurrently running query.
func (q *QueryState1[T]) ParIter(world *World, fn func(Entity, *T))

// minChunkSize is the minimum number of entities per goroutine to avoid
// excessive goroutine overhead.
const minChunkSize = 256
```

### Cache Invalidation

```go
// updateCache checks the World's archetype generation counter.
// If new archetypes have been created since the last update,
// scans new archetypes for matches and appends to the cache.
//
// Pseudo-code:
//   if world.archetypes.generation == q.archetypeGen: return
//   for each new archetype since q.archetypeGen:
//     if archetype contains all required components:
//       if all filters match:
//         append to matchedArchetypes
//   q.archetypeGen = world.archetypes.generation
func (q *QueryState1[T]) updateCache(world *World)
```

### Access Conflict Detection

```go
// Conflicts reports whether two query accesses conflict.
// Used at schedule build time to determine system ordering.
//
// Conflict rules:
//   - Write-Write on same ComponentID: conflict
//   - Write-Read on same ComponentID: conflict
//   - Read-Read: no conflict
func (a Access) Conflicts(other Access) bool
```

Schedule-time validation (pseudo-code):

```
func validateSchedule(systems []System) error:
    for each pair (sysA, sysB) in systems:
        if sysA.Access().Conflicts(sysB.Access()):
            if no explicit ordering between sysA and sysB:
                return error("ambiguous system order: {sysA} and {sysB} conflict on {component}")
    return nil
```

### Count and Single

```go
// Count returns the number of entities matching the query.
func (q *QueryState1[T]) Count(world *World) int

// Single asserts that exactly one entity matches. Returns it directly.
// Returns an error if zero or more than one entity matches.
func (q *QueryState1[T]) Single(world *World) (Entity, *T, error)
```

## Performance Strategy

- **Archetype cache**: `matchedArchetypes` stores pre-resolved archetype/table/column references. Iteration avoids any map lookups.
- **Incremental cache update**: Only new archetypes are scanned (since last `archetypeGen`), not all archetypes.
- **Column pointer arithmetic**: Component access via `unsafe.Pointer` on contiguous `[]byte`. No interface boxing during iteration.
- **Tick filter short-circuit**: Archetype-level `Matches()` filters entire archetypes before row-level tick checks. Empty archetypes skipped.
- **ParIter chunk sizing**: Minimum 256 entities per goroutine prevents overhead from dominating small workloads.
- **Zero allocations in Iter**: Callback pattern avoids slice/iterator allocation. The `fn` closure captures state without heap escape (compiler-dependent; benchmarks will verify).
- **Access sets as sorted slices**: Conflict detection uses merge-join on sorted `[]ComponentID`, O(n+m) instead of O(n*m).

## Error Handling

- `Get` on dead entity: returns `ErrEntityNotAlive`.
- `Get` on entity without matching components: returns `ErrComponentNotFound`.
- `Single` with zero matches: returns `ErrQueryNoMatch`.
- `Single` with multiple matches: returns `ErrQueryMultipleMatches`.
- Access conflict detected at schedule build time: returns error with descriptive message naming the conflicting systems and components.
- `ParIter` panics are caught per-goroutine and returned as errors via `errgroup`.

```go
var (
    ErrQueryNoMatch        = errors.New("ecs: query matched zero entities")
    ErrQueryMultipleMatches = errors.New("ecs: query matched more than one entity")
)
```

## Testing Strategy

- **Basic iteration**: Spawn entities with known components, iterate via query, verify all and only matching entities are visited.
- **Filter correctness**: `With`, `Without`, `Changed`, `Added` — verify correct filtering at archetype and entity level.
- **Cache invalidation**: Spawn entities, iterate, spawn more entities with new component sets (creating new archetypes), iterate again — verify new entities included.
- **Get/Single**: Verify correct return for existing, missing, and dead entities.
- **Access conflict detection**: Build test schedules with overlapping/disjoint access sets, verify conflict detection.
- **ParIter correctness**: Run parallel iteration on 10K+ entities, verify every entity visited exactly once (use atomic counter).
- **Multi-arity queries**: Test QueryState2, QueryState3, QueryState4 with various component combinations.
- **Mut marker**: Verify mutable access updates change ticks, verify access set reflects write.
- **Benchmarks**: `BenchmarkIter1Component`, `BenchmarkIter3Components`, `BenchmarkParIter`, `BenchmarkQueryGet`, `BenchmarkCacheUpdate`. Compare callback iteration vs. hypothetical iterator alternatives.
- **Fuzz tests**: Random spawn/despawn/query sequences to verify no panics or missed entities.

## Open Questions

- Should we provide a code generator for `QueryState5` through `QueryState8`, or define them manually?
- Should `ParIter` accept a `context.Context` for cancellation, or rely on the schedule executor's context?
- Should query construction happen lazily (on first `Iter` call) or eagerly (at system registration)?
- Is callback-based iteration ergonomic enough, or should we explore Go 1.23+ range-over-func iterators?
- Should `Optional[T]` be a query fetch type (returns `*T` or nil for entities without T)?

## Document History

| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-26 | Initial L2 draft |
