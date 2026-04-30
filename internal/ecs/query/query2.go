package query

import (
	"iter"

	"github.com/teratron/ecs-engine/internal/ecs/component"
	"github.com/teratron/ecs-engine/internal/ecs/entity"
	"github.com/teratron/ecs-engine/internal/ecs/world"
)

// Query2 is a two-component query. It iterates every live entity whose
// archetype contains both A and B and yields (Entity, [Tuple2]) pairs via
// [Query2.All].
type Query2[A, B any] struct {
	state    *QueryState
	ids      [2]component.ID
	matched  []world.ArchetypeID
	nextScan int
}

// NewQuery2 builds a two-component query, auto-registering A and B as
// [component.StorageTable] components when first seen. Returns an error if
// A and B are the same Go type (a query asking for the same component twice
// is a programmer error, not a runtime condition).
func NewQuery2[A, B any](w *world.World) (*Query2[A, B], error) {
	idA := componentIDFor[A](w)
	idB := componentIDFor[B](w)
	if idA == idB {
		return nil, errSameTypeInQuery
	}
	state, err := NewQueryState([]component.ID{idA, idB}, nil, Access{})
	if err != nil {
		return nil, err
	}
	return &Query2[A, B]{state: state, ids: [2]component.ID{idA, idB}}, nil
}

// State returns the underlying [QueryState].
func (q *Query2[A, B]) State() *QueryState { return q.state }

// All returns an iterator over every (entity, Tuple2[*A, *B]) pair matching
// the query.
//
// Usage:
//
//	for e, t := range q.All(world) {
//	    _ = e
//	    t.A.X += t.B.DX
//	}
func (q *Query2[A, B]) All(w *world.World) iter.Seq2[entity.Entity, Tuple2[A, B]] {
	q.refresh(w)
	return func(yield func(entity.Entity, Tuple2[A, B]) bool) {
		for _, archID := range q.matched {
			arch := w.Archetypes().At(archID)
			entities := arch.Entities()
			for row, e := range entities {
				tup := Tuple2[A, B]{
					A: (*A)(fetchComponent(w, arch, e, row, q.ids[0])),
					B: (*B)(fetchComponent(w, arch, e, row, q.ids[1])),
				}
				if !yield(e, tup) {
					return
				}
			}
		}
	}
}

// Count returns the number of entities currently matching the query.
func (q *Query2[A, B]) Count(w *world.World) int {
	q.refresh(w)
	n := 0
	for _, archID := range q.matched {
		n += w.Archetypes().At(archID).Len()
	}
	return n
}

func (q *Query2[A, B]) refresh(w *world.World) {
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
