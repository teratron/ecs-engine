package component

import (
	"fmt"
	"reflect"
)

// RequiredComponents is an optional interface a component type can implement
// to declare components that must accompany it on every entity. The Registry
// resolves the dependency graph transitively at registration time and stores
// the result on [Info.RequiredBy].
//
// The implementation is invoked once per registration on a freshly
// allocated zero value of the type; the returned [Data] values supply the
// default payloads inserted by the World when a missing dependency is
// detected.
type RequiredComponents interface {
	Required() []Data
}

var requiredComponentsIface = reflect.TypeOf((*RequiredComponents)(nil)).Elem()

// asRequiredComponents returns the RequiredComponents view of the zero
// value of t, or nil if t (and *t) do not implement the interface.
func asRequiredComponents(t reflect.Type) RequiredComponents {
	ptr := reflect.New(t)
	if ptr.Type().Implements(requiredComponentsIface) {
		return ptr.Interface().(RequiredComponents)
	}
	if t.Implements(requiredComponentsIface) {
		return ptr.Elem().Interface().(RequiredComponents)
	}
	return nil
}

// resolveRequired walks the RequiredComponents graph rooted at info,
// registers any not-yet-seen dependency types, and returns the transitive
// list of required IDs in leaves-first order. Cycles are detected via the
// `visiting` set and trigger a panic with the offending chain.
func (r *Registry) resolveRequired(info *Info, visiting map[reflect.Type]bool) []ID {
	rc := asRequiredComponents(info.Type)
	if rc == nil {
		return nil
	}

	if visiting[info.Type] {
		panic(fmt.Sprintf(
			"component.Registry: circular required-component dependency at %s",
			info.Type,
		))
	}
	visiting[info.Type] = true
	defer delete(visiting, info.Type)

	var out []ID
	seen := make(map[ID]bool)
	for _, dep := range rc.Required() {
		if dep.Value == nil {
			panic(fmt.Sprintf(
				"component.Registry: %s.Required() returned a Data with nil Value",
				info.Type,
			))
		}
		depType := reflect.TypeOf(dep.Value)
		depID := r.registerType(depType, visiting)

		// Leaves-first ordering: first emit the dependency's own transitive
		// requirements, then the dependency itself.
		depInfo := &r.infosByID[depID]
		for _, sub := range depInfo.RequiredBy {
			if !seen[sub] {
				seen[sub] = true
				out = append(out, sub)
			}
		}
		if !seen[depID] {
			seen[depID] = true
			out = append(out, depID)
		}
	}
	return out
}

// registerType is the internal entry point used by required-component
// resolution. It registers a Go type under the default ColumnSpec, threading
// the visiting-set so cycles spanning multiple types are detected. Public
// callers should use [Registry.Register] or [RegisterType].
//
// The cycle check fires before the typeToID short-circuit so that a type
// already present in the registry is still rejected when re-encountered
// inside the same resolution chain.
func (r *Registry) registerType(t reflect.Type, visiting map[reflect.Type]bool) ID {
	if visiting[t] {
		panic(fmt.Sprintf(
			"component.Registry: circular required-component dependency at %s",
			t,
		))
	}
	if existing, ok := r.typeToID[t]; ok {
		return existing
	}
	info := Info{Type: t, Storage: StorageTable}
	id := r.insert(info)
	r.infosByID[id].RequiredBy = r.resolveRequired(&r.infosByID[id], visiting)
	return id
}
