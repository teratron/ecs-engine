// Package world provides the central ECS data store: the World, which owns
// all entities, component storage, resources, archetypes, and schedules.
package world

import (
	"errors"

	"github.com/teratron/ecs-engine/internal/ecs/component"
	"github.com/teratron/ecs-engine/internal/ecs/entity"
)

// Tick is a monotonically increasing counter used for change detection.
// Incremented each time a system runs; never wraps within a single game run.
type Tick uint32

// IsNewerThan returns true when t was set after last (accounting for the
// standard change-detection contract: a component is "changed" when its
// tick is strictly newer than the last-cleared tick).
func (t Tick) IsNewerThan(last Tick) bool { return t > last }

var (
	ErrEntityNotAlive    = errors.New("ecs: entity is not alive")
	ErrScheduleNotFound  = errors.New("ecs: schedule not found")
	ErrComponentNotFound = errors.New("ecs: component not found on entity")
)

const (
	defaultEntityCapacity    = 256
	defaultComponentCapacity = 64
)

// World is the central data store of the ECS engine. It owns all entities,
// component registrations, resources, and the global change tick.
// Not thread-safe — concurrent access must be coordinated by the schedule
// executor.
//
// Archetype storage (tables, sparse sets, archetype graph) is initialized
// during T-1C03 and lives in separate fields added in that task.
type World struct {
	entities       *entity.EntityAllocator
	components     *component.Registry
	resources      *ResourceMap
	changeTick     Tick
	lastChangeTick Tick
}

// NewWorld creates a World with default initial capacities.
func NewWorld() *World {
	return NewWorldWithCapacity(defaultEntityCapacity, defaultComponentCapacity)
}

// NewWorldWithCapacity creates a World pre-allocated for the expected number
// of entities and component types. Both values are hints; the World grows
// automatically beyond them.
func NewWorldWithCapacity(entityCapacity, _ int) *World {
	return &World{
		entities:   entity.NewEntityAllocator(entityCapacity),
		components: component.NewRegistry(),
		resources:  NewResourceMap(),
	}
}

// Entities exposes the underlying EntityAllocator for packages that need
// direct allocator access (archetype graph, commands).
func (w *World) Entities() *entity.EntityAllocator { return w.entities }

// Components exposes the underlying component Registry for packages that
// register or look up component metadata.
func (w *World) Components() *component.Registry { return w.components }

// Resources exposes the ResourceMap for packages that need bulk resource
// iteration or direct map access (e.g., serialization).
func (w *World) Resources() *ResourceMap { return w.resources }

// SpawnEmpty allocates a new entity with no components.
func (w *World) SpawnEmpty() entity.Entity {
	return w.entities.Allocate()
}

// Contains reports whether the entity is currently alive in this World.
func (w *World) Contains(e entity.Entity) bool {
	return w.entities.IsAlive(e)
}

// Despawn removes the entity from the World. Returns ErrEntityNotAlive if
// the entity is already dead or was never part of this World.
// Component removal hooks and archetype migration are handled in T-1C03.
func (w *World) Despawn(e entity.Entity) error {
	if !w.entities.IsAlive(e) {
		return ErrEntityNotAlive
	}
	w.entities.Free(e)
	return nil
}

// ChangeTick returns the current global change tick.
func (w *World) ChangeTick() Tick { return w.changeTick }

// LastChangeTick returns the tick value at the last ClearTrackers call.
func (w *World) LastChangeTick() Tick { return w.lastChangeTick }

// IncrementChangeTick advances the global tick by one and returns the new
// value. Called by the schedule executor before each system run.
func (w *World) IncrementChangeTick() Tick {
	w.changeTick++
	return w.changeTick
}

// ClearTrackers advances lastChangeTick to the current changeTick, resetting
// change-detection state for the next update cycle. Called once per frame,
// typically at the end of the update schedule.
func (w *World) ClearTrackers() {
	w.lastChangeTick = w.changeTick
}
