package query

import (
	"errors"
	"iter"

	"github.com/teratron/ecs-engine/internal/ecs/component"
	"github.com/teratron/ecs-engine/internal/ecs/entity"
	"github.com/teratron/ecs-engine/internal/ecs/world"
)

// ErrQueryNoMatch is returned by Single-style helpers when zero entities
// match the query.
var ErrQueryNoMatch = errors.New("ecs: query matched zero entities")

// ErrQueryMultipleMatches is returned by Single-style helpers when more than
// one entity matches the query.
var ErrQueryMultipleMatches = errors.New("ecs: query matched more than one entity")

// Query1 is a single-component query. It iterates every live entity whose
// archetype contains T and yields (Entity, *T) pairs via [Query1.All].
//
// The matched-archetype list is grown lazily — every call to [Query1.All]
// (and other terminal methods) scans archetypes created since the last
// invocation. Phase 1 archetypes are append-only, so a watermark suffices;
// future phases will adopt the generation-counter invalidation contract
// described in [QueryState].
type Query1[T any] struct {
	state    *QueryState
	id       component.ID
	matched  []world.ArchetypeID
	nextScan int
}

// NewQuery1 builds a single-component query, auto-registering T as a
// [component.StorageTable] component if it is not yet known to the world's
// registry. Returns an error only when the implicit [Access] declaration
// fails validation (currently impossible with the default policy, but kept
// in the signature to match T-1D03 filter integration).
func NewQuery1[T any](w *world.World) (*Query1[T], error) {
	id := componentIDFor[T](w)
	state, err := NewQueryState([]component.ID{id}, nil, Access{})
	if err != nil {
		return nil, err
	}
	return &Query1[T]{state: state, id: id}, nil
}

// State returns the underlying [QueryState] (used by the scheduler to read
// access metadata for conflict detection).
func (q *Query1[T]) State() *QueryState { return q.state }

// All returns an iterator over every (entity, *T) pair matching the query.
// The pointer is valid for the duration of the iteration step; storing it
// past a structural mutation (Spawn/Insert/Remove/Despawn) is undefined.
//
// Usage:
//
//	for e, t := range q.All(world) {
//	    _ = e
//	    t.Field = ...
//	}
func (q *Query1[T]) All(w *world.World) iter.Seq2[entity.Entity, *T] {
	q.refresh(w)
	return func(yield func(entity.Entity, *T) bool) {
		for _, archID := range q.matched {
			arch := w.Archetypes().At(archID)
			entities := arch.Entities()
			for row, e := range entities {
				ptr := fetchComponent(w, arch, e, row, q.id)
				if !yield(e, (*T)(ptr)) {
					return
				}
			}
		}
	}
}

// Count returns the number of entities currently matching the query.
func (q *Query1[T]) Count(w *world.World) int {
	q.refresh(w)
	n := 0
	for _, archID := range q.matched {
		n += w.Archetypes().At(archID).Len()
	}
	return n
}

// Single asserts that exactly one entity matches the query and returns it.
// Returns [ErrQueryNoMatch] or [ErrQueryMultipleMatches] when the count is
// zero or greater than one, respectively.
func (q *Query1[T]) Single(w *world.World) (entity.Entity, *T, error) {
	q.refresh(w)
	var (
		found     bool
		gotEntity entity.Entity
		gotPtr    *T
	)
	for _, archID := range q.matched {
		arch := w.Archetypes().At(archID)
		entities := arch.Entities()
		for row, e := range entities {
			if found {
				return entity.Entity{}, nil, ErrQueryMultipleMatches
			}
			found = true
			gotEntity = e
			gotPtr = (*T)(fetchComponent(w, arch, e, row, q.id))
		}
	}
	if !found {
		return entity.Entity{}, nil, ErrQueryNoMatch
	}
	return gotEntity, gotPtr, nil
}

// refresh appends every newly created archetype that matches the query's
// state to q.matched. Idempotent when no new archetypes have appeared.
func (q *Query1[T]) refresh(w *world.World) {
	store := w.Archetypes()
	if q.nextScan == store.Len() {
		return
	}
	store.EachFrom(q.nextScan, func(a *world.Archetype) bool {
		if q.state.MatchesIDs(a.ComponentIDs()) {
			q.matched = append(q.matched, a.ID())
		}
		return true
	})
	q.nextScan = store.Len()
}
