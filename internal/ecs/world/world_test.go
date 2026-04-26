package world

import (
	"testing"
)

// resource fixtures
type Health struct{ HP int }
type Mana struct{ MP int }
type Score struct{ Value int }

func TestNewWorld(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	if w == nil {
		t.Fatal("NewWorld must return non-nil")
	}
	if w.entities == nil || w.components == nil || w.resources == nil {
		t.Fatal("NewWorld must initialise entities, components, and resources")
	}
}

func TestNewWorldWithCapacity(t *testing.T) {
	t.Parallel()

	w := NewWorldWithCapacity(1024, 128)
	if w == nil {
		t.Fatal("NewWorldWithCapacity must return non-nil")
	}
	if w.Entities() == nil || w.Components() == nil || w.Resources() == nil {
		t.Fatal("accessor must return non-nil subsystems")
	}
}

// Entity lifecycle tests.

func TestSpawnEmptyAlive(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	e := w.SpawnEmpty()
	if !e.IsValid() {
		t.Fatal("SpawnEmpty must return a valid entity")
	}
	if !w.Contains(e) {
		t.Fatal("entity must be alive after SpawnEmpty")
	}
}

func TestDespawnMakesEntityDead(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	e := w.SpawnEmpty()
	if err := w.Despawn(e); err != nil {
		t.Fatalf("Despawn alive entity: %v", err)
	}
	if w.Contains(e) {
		t.Fatal("entity must be dead after Despawn")
	}
}

func TestDespawnDeadEntityReturnsError(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	e := w.SpawnEmpty()
	_ = w.Despawn(e)
	if err := w.Despawn(e); err != ErrEntityNotAlive {
		t.Fatalf("double Despawn must return ErrEntityNotAlive; got %v", err)
	}
}

func TestContainsNeverSpawned(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	// Spawn and immediately free to get a stale handle.
	e := w.SpawnEmpty()
	_ = w.Despawn(e)
	if w.Contains(e) {
		t.Fatal("stale entity must not be contained")
	}
}

func TestMultipleEntitiesIndependent(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	a := w.SpawnEmpty()
	b := w.SpawnEmpty()

	_ = w.Despawn(a)

	if w.Contains(a) {
		t.Error("a must be dead after despawn")
	}
	if !w.Contains(b) {
		t.Error("b must still be alive")
	}
}

// Change tick tests.

func TestChangeTickStartsAtZero(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	if w.ChangeTick() != 0 {
		t.Fatalf("initial change tick must be 0; got %d", w.ChangeTick())
	}
}

func TestIncrementChangeTick(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	got := w.IncrementChangeTick()
	if got != 1 {
		t.Fatalf("first increment must yield 1; got %d", got)
	}
	got = w.IncrementChangeTick()
	if got != 2 {
		t.Fatalf("second increment must yield 2; got %d", got)
	}
	if w.ChangeTick() != 2 {
		t.Fatalf("ChangeTick() after two increments = %d, want 2", w.ChangeTick())
	}
}

func TestClearTrackers(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	w.IncrementChangeTick()
	w.IncrementChangeTick()
	w.IncrementChangeTick()
	if w.LastChangeTick() != 0 {
		t.Fatalf("LastChangeTick before ClearTrackers = %d, want 0", w.LastChangeTick())
	}
	w.ClearTrackers()
	if w.LastChangeTick() != w.ChangeTick() {
		t.Fatalf("after ClearTrackers: last=%d, current=%d; must match",
			w.LastChangeTick(), w.ChangeTick())
	}
}

func TestTickIsNewerThan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		t, last Tick
		want    bool
	}{
		{5, 3, true},
		{3, 5, false},
		{3, 3, false},
		{1, 0, true},
	}
	for _, c := range cases {
		if got := c.t.IsNewerThan(c.last); got != c.want {
			t.Errorf("Tick(%d).IsNewerThan(%d) = %v, want %v", c.t, c.last, got, c.want)
		}
	}
}

// ResourceMap tests.

func TestSetAndGetResource(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	SetResource(w, Health{HP: 100})

	got, ok := Resource[Health](w)
	if !ok {
		t.Fatal("Resource[Health] must be found after SetResource")
	}
	if got.HP != 100 {
		t.Fatalf("Resource[Health].HP = %d, want 100", got.HP)
	}
}

func TestResourceOverwrite(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	SetResource(w, Health{HP: 50})
	SetResource(w, Health{HP: 200})

	got, ok := Resource[Health](w)
	if !ok || got.HP != 200 {
		t.Fatalf("overwritten resource = %+v (ok=%v), want HP=200", got, ok)
	}
}

func TestResourceNotFound(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	ptr, ok := Resource[Mana](w)
	if ok || ptr != nil {
		t.Fatal("Resource[Mana] must return nil,false when not set")
	}
}

func TestContainsResource(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	if ContainsResource[Score](w) {
		t.Fatal("ContainsResource must be false before insertion")
	}
	SetResource(w, Score{Value: 42})
	if !ContainsResource[Score](w) {
		t.Fatal("ContainsResource must be true after insertion")
	}
}

func TestRemoveResource(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	SetResource(w, Score{Value: 10})

	if !RemoveResource[Score](w) {
		t.Fatal("RemoveResource must return true when resource existed")
	}
	if RemoveResource[Score](w) {
		t.Fatal("second RemoveResource must return false (already gone)")
	}
	if ContainsResource[Score](w) {
		t.Fatal("resource must not exist after removal")
	}
}

func TestResourceTypeIsolation(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	SetResource(w, Health{HP: 99})
	SetResource(w, Mana{MP: 77})

	h, ok := Resource[Health](w)
	if !ok || h.HP != 99 {
		t.Fatalf("Health resource = %+v (ok=%v), want HP=99", h, ok)
	}
	m, ok := Resource[Mana](w)
	if !ok || m.MP != 77 {
		t.Fatalf("Mana resource = %+v (ok=%v), want MP=77", m, ok)
	}
	if w.resources.Len() != 2 {
		t.Fatalf("resources.Len() = %d, want 2", w.resources.Len())
	}
}

func TestResourceWorldIsolation(t *testing.T) {
	t.Parallel()

	w1 := NewWorld()
	w2 := NewWorld()
	SetResource(w1, Health{HP: 10})

	if ContainsResource[Health](w2) {
		t.Fatal("resource set in w1 must not appear in w2")
	}
}

func TestResourceMutablePointer(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	SetResource(w, Score{Value: 1})

	ptr, _ := Resource[Score](w)
	ptr.Value = 42

	// Re-fetch: the stored pointer is the same object, so mutation is visible.
	got, _ := Resource[Score](w)
	if got.Value != 42 {
		t.Fatalf("mutation via pointer = %d, want 42", got.Value)
	}
}

// ResourceMap direct tests.

func TestResourceMapLen(t *testing.T) {
	t.Parallel()

	w := NewWorld()
	if w.resources.Len() != 0 {
		t.Fatalf("empty ResourceMap.Len() = %d, want 0", w.resources.Len())
	}
	SetResource(w, Health{HP: 1})
	SetResource(w, Mana{MP: 2})
	if w.resources.Len() != 2 {
		t.Fatalf("ResourceMap.Len() = %d, want 2", w.resources.Len())
	}
}
