package entity

// DisabledTag is an empty-struct tag component marking entities that should
// be skipped by the default scheduler iteration. Per ECS design rules
// (CLAUDE.md §4.6), tags are zero-size structs and act as filters rather
// than boolean flags inside larger components.
//
// Usage (illustrative — component registration lands in T-1B01):
//
//	world.AddComponent(entity, entity.DisabledTag{})
//
// Queries can filter with `Without[DisabledTag]` once the query layer
// (Track D) is implemented.
type DisabledTag struct{}
