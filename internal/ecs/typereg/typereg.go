// Package typereg is the runtime type registry: a metadata store that pairs
// a Go reflect.Type with a dense numeric [TypeID], cached struct-field
// metadata, and parsed struct-tag attributes. The registry is the single
// source of truth used by serialization, editor introspection, and any code
// that needs to walk fields of a type-erased value.
//
// # Phase 1 scope (T-1H01)
//
// Pure registry work: register a type, resolve by ID/type/name, walk fields
// with their pre-computed offsets and parsed tags. [DynamicObject] proxies,
// the serialization-hook contract, and the integration with the
// component/resource/event registries land in T-1H02.
//
// # Concurrency
//
// Registration is intended to happen at app setup on a single goroutine —
// just like component and event registration. Reads (Resolve*) are safe for
// concurrent use after every type has been registered, since the underlying
// maps and dense slice are not mutated post-registration.
package typereg

import "errors"

// TypeID is a dense, monotonically-assigned numeric identifier for a
// registered Go type. Index 0 is reserved as the invalid sentinel —
// [TypeRegistry.ResolveByID] returns nil for it.
type TypeID uint32

// IsValid reports whether id refers to a registered type (i.e. is non-zero).
// It does not perform a range check against any specific registry.
func (id TypeID) IsValid() bool { return id != 0 }

// Sentinel errors returned by registry operations.
var (
	// ErrTypeNotRegistered indicates that the queried type has never been
	// registered. Callers that prefer panics should use [MustResolve].
	ErrTypeNotRegistered = errors.New("typereg: type not registered")

	// ErrDuplicateTypeName is returned when two distinct Go types resolve to
	// the same fully-qualified name (the registry's name index is unique).
	ErrDuplicateTypeName = errors.New("typereg: duplicate fully-qualified type name")
)
