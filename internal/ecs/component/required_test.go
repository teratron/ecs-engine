package component

import (
	"reflect"
	"testing"
)

// Test fixtures for the required-component graph.

type Transform struct{ X, Y, Z float32 }

// Renderable requires Transform.
type Renderable struct{ Mesh uint32 }

func (Renderable) Required() []Data {
	return []Data{{Value: Transform{}}}
}

// Animator requires Renderable (transitively also Transform).
type Animator struct{ Clip uint32 }

func (Animator) Required() []Data {
	return []Data{{Value: Renderable{}}}
}

// Cycle: A requires B, B requires A.
type cycA struct{}

func (cycA) Required() []Data { return []Data{{Value: cycB{}}} }

type cycB struct{}

func (cycB) Required() []Data { return []Data{{Value: cycA{}}} }

// Self-cycle: SelfRef requires itself.
type SelfRef struct{}

func (SelfRef) Required() []Data { return []Data{{Value: SelfRef{}}} }

// Bad: Required returns a Data with a nil Value.
type BadNilValue struct{}

func (BadNilValue) Required() []Data { return []Data{{Value: nil}} }

// Pointer-receiver implementation should still be detected.
type ptrRcvr struct{}

func (*ptrRcvr) Required() []Data { return []Data{{Value: Transform{}}} }

func TestRequiredEmptyForPlainComponent(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	id := RegisterType[Position](r)
	if got := r.Info(id).RequiredBy; len(got) != 0 {
		t.Fatalf("Position must have no requirements; got %v", got)
	}
}

func TestRequiredSingleDependencyAutoRegistersDep(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	id := RegisterType[Renderable](r)

	depID, ok := r.Lookup(reflect.TypeOf(Transform{}))
	if !ok {
		t.Fatal("Transform must be auto-registered as a dependency of Renderable")
	}
	got := r.Info(id).RequiredBy
	if len(got) != 1 || got[0] != depID {
		t.Fatalf("Renderable.RequiredBy = %v, want [%d]", got, depID)
	}
}

func TestRequiredTransitiveLeavesFirst(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	id := RegisterType[Animator](r)

	transformID, _ := r.Lookup(reflect.TypeOf(Transform{}))
	renderID, _ := r.Lookup(reflect.TypeOf(Renderable{}))

	got := r.Info(id).RequiredBy
	want := []ID{transformID, renderID}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Animator transitive deps = %v, want %v (leaves first)", got, want)
	}
}

func TestRequiredCycleAcrossTypesPanics(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	defer func() {
		if recover() == nil {
			t.Fatal("cycle A↔B must panic at registration")
		}
	}()
	RegisterType[cycA](r)
}

func TestRequiredSelfCyclePanics(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	defer func() {
		if recover() == nil {
			t.Fatal("self-referential Required must panic")
		}
	}()
	RegisterType[SelfRef](r)
}

func TestRequiredNilValuePanics(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	defer func() {
		if recover() == nil {
			t.Fatal("Required() returning nil Value must panic")
		}
	}()
	RegisterType[BadNilValue](r)
}

func TestRequiredPointerReceiverDetected(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	id := RegisterType[ptrRcvr](r)
	transformID, ok := r.Lookup(reflect.TypeOf(Transform{}))
	if !ok {
		t.Fatal("pointer-receiver Required() must still register Transform")
	}
	if got := r.Info(id).RequiredBy; len(got) != 1 || got[0] != transformID {
		t.Fatalf("ptrRcvr.RequiredBy = %v, want [%d]", got, transformID)
	}
}

func TestRequiredIsIdempotentAcrossDuplicateRegistrations(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	id1 := RegisterType[Animator](r)
	id2 := RegisterType[Animator](r)
	if id1 != id2 {
		t.Fatalf("re-registering Animator must yield same ID; got %d / %d", id1, id2)
	}
	if r.Len() != 3 {
		t.Fatalf("registry size after duplicate Animator = %d, want 3", r.Len())
	}
}
