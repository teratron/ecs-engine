package query

import (
	"errors"
	"iter"

	"github.com/teratron/ecs-engine/internal/ecs/component"
	"github.com/teratron/ecs-engine/internal/ecs/entity"
	"github.com/teratron/ecs-engine/internal/ecs/world"
)

// errSameTypeInQuery is returned when a multi-arity query is constructed
// with two identical type parameters. Hoisted to a sentinel so callers can
// match it via errors.Is.
var errSameTypeInQuery = errors.New("ecs: query type parameters must be distinct components")

// Query3 is a three-component query. It iterates every live entity whose
// archetype contains A, B, and C and yields (Entity, [Tuple3]) pairs via
// [Query3.All].
type Query3[A, B, C any] struct {
	state    *QueryState
	ids      [3]component.ID
	matched  []world.ArchetypeID
	nextScan int
}

// NewQuery3 builds a three-component query, auto-registering A, B, and C as
// [component.StorageTable] components when first seen. Returns an error if
// any two of the type parameters resolve to the same component.
func NewQuery3[A, B, C any](w *world.World) (*Query3[A, B, C], error) {
	idA := componentIDFor[A](w)
	idB := componentIDFor[B](w)
	idC := componentIDFor[C](w)
	if idA == idB || idA == idC || idB == idC {
		return nil, errSameTypeInQuery
	}
	state, err := NewQueryState([]component.ID{idA, idB, idC}, nil, Access{})
	if err != nil {
		return nil, err
	}
	return &Query3[A, B, C]{state: state, ids: [3]component.ID{idA, idB, idC}}, nil
}

// State returns the underlying [QueryState].
func (q *Query3[A, B, C]) State() *QueryState { return q.state }

// All returns an iterator over every (entity, Tuple3[*A, *B, *C]) tuple.
func (q *Query3[A, B, C]) All(w *world.World) iter.Seq2[entity.Entity, Tuple3[A, B, C]] {
	q.refresh(w)
	return func(yield func(entity.Entity, Tuple3[A, B, C]) bool) {
		for _, archID := range q.matched {
			arch := w.Archetypes().At(archID)
			entities := arch.Entities()
			for row, e := range entities {
				tup := Tuple3[A, B, C]{
					A: (*A)(fetchComponent(w, arch, e, row, q.ids[0])),
					B: (*B)(fetchComponent(w, arch, e, row, q.ids[1])),
					C: (*C)(fetchComponent(w, arch, e, row, q.ids[2])),
				}
				if !yield(e, tup) {
					return
				}
			}
		}
	}
}

// Count returns the number of entities currently matching the query.
func (q *Query3[A, B, C]) Count(w *world.World) int {
	q.refresh(w)
	n := 0
	for _, archID := range q.matched {
		n += w.Archetypes().At(archID).Len()
	}
	return n
}

func (q *Query3[A, B, C]) refresh(w *world.World) {
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
