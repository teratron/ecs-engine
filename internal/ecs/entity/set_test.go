package entity

import (
	"sort"
	"testing"
	"unsafe"
)

func TestEntitySetInsertContainsLen(t *testing.T) {
	t.Parallel()

	s := NewEntitySet()
	if s.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", s.Len())
	}

	e1 := NewEntity(1, 1)
	e2 := NewEntity(2, 1)

	if !s.Insert(e1) {
		t.Fatal("Insert(e1) must return true on first add")
	}
	if s.Insert(e1) {
		t.Fatal("Insert(e1) must return false on duplicate")
	}
	if !s.Insert(e2) {
		t.Fatal("Insert(e2) must return true on first add")
	}
	if s.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", s.Len())
	}
	if !s.Contains(e1) || !s.Contains(e2) {
		t.Fatal("Contains must report true for inserted entities")
	}
}

func TestEntitySetRejectsNullEntity(t *testing.T) {
	t.Parallel()

	s := NewEntitySet()
	if s.Insert(Entity{}) {
		t.Fatal("Insert(null) must return false")
	}
	if s.Contains(Entity{}) {
		t.Fatal("Contains(null) must return false")
	}
	if s.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", s.Len())
	}
}

func TestEntitySetRemoveSwapAndPop(t *testing.T) {
	t.Parallel()

	s := NewEntitySet()
	e1 := NewEntity(1, 1)
	e2 := NewEntity(2, 1)
	e3 := NewEntity(3, 1)
	s.Insert(e1)
	s.Insert(e2)
	s.Insert(e3)

	if !s.Remove(e1) {
		t.Fatal("Remove(e1) must return true")
	}
	if s.Remove(e1) {
		t.Fatal("Remove(e1) on absent must return false")
	}
	if s.Contains(e1) {
		t.Fatal("e1 must be gone after Remove")
	}
	if !s.Contains(e2) || !s.Contains(e3) {
		t.Fatal("e2/e3 must still be present after removing e1")
	}
	if s.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", s.Len())
	}

	if !s.Remove(e3) { // remove last (no swap)
		t.Fatal("Remove(e3) must return true")
	}
	if s.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", s.Len())
	}
}

func TestEntitySetIterCoversAll(t *testing.T) {
	t.Parallel()

	s := NewEntitySet()
	want := []Entity{
		NewEntity(10, 1),
		NewEntity(20, 1),
		NewEntity(30, 1),
	}
	for _, e := range want {
		s.Insert(e)
	}

	got := make([]Entity, 0, len(want))
	s.Iter(func(e Entity) { got = append(got, e) })

	if len(got) != len(want) {
		t.Fatalf("Iter visited %d entities, want %d", len(got), len(want))
	}
	sortEntities(got)
	sortEntities(want)
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("Iter element %d = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestEntitySetClear(t *testing.T) {
	t.Parallel()

	s := NewEntitySet()
	for i := uint32(1); i <= 5; i++ {
		s.Insert(NewEntity(i, 1))
	}
	s.Clear()
	if s.Len() != 0 {
		t.Fatalf("Len() after Clear = %d, want 0", s.Len())
	}
	if s.Contains(NewEntity(1, 1)) {
		t.Fatal("Contains must be false after Clear")
	}

	if !s.Insert(NewEntity(7, 1)) {
		t.Fatal("Insert after Clear must succeed")
	}
}

func TestEntitySetWithCapacity(t *testing.T) {
	t.Parallel()

	s := NewEntitySetWithCapacity(-3)
	if s == nil {
		t.Fatal("negative capacity must not yield nil")
	}
	if cap(s.dense) != 0 {
		t.Fatalf("clamped capacity expected 0, got %d", cap(s.dense))
	}

	s2 := NewEntitySetWithCapacity(64)
	if cap(s2.dense) < 64 {
		t.Fatalf("dense capacity = %d, want ≥ 64", cap(s2.dense))
	}
}

func TestEntityMapSetGetRemove(t *testing.T) {
	t.Parallel()

	m := NewEntityMap[string]()
	e := NewEntity(1, 1)

	if v, ok := m.Get(e); ok || v != "" {
		t.Fatal("Get on empty map must return zero / false")
	}

	m.Set(e, "hello")
	v, ok := m.Get(e)
	if !ok || v != "hello" {
		t.Fatalf("Get(e) = (%q, %v), want (\"hello\", true)", v, ok)
	}
	if !m.Contains(e) {
		t.Fatal("Contains must be true after Set")
	}
	if m.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", m.Len())
	}

	if !m.Remove(e) {
		t.Fatal("Remove must return true on present entry")
	}
	if m.Remove(e) {
		t.Fatal("Remove must return false on absent entry")
	}
	if m.Contains(e) {
		t.Fatal("Contains must be false after Remove")
	}
}

func TestEntityMapRejectsNullEntity(t *testing.T) {
	t.Parallel()

	m := NewEntityMap[int]()
	m.Set(Entity{}, 42)
	if m.Len() != 0 {
		t.Fatalf("null-entity Set must be ignored; Len() = %d", m.Len())
	}
	if _, ok := m.Get(Entity{}); ok {
		t.Fatal("null-entity Get must report not-present")
	}
}

func TestEntityMapIter(t *testing.T) {
	t.Parallel()

	m := NewEntityMap[int]()
	for i := uint32(1); i <= 5; i++ {
		m.Set(NewEntity(i, 1), int(i*10))
	}
	visited := map[uint32]int{}
	m.Iter(func(e Entity, v int) {
		visited[e.Index()] = v
	})
	if len(visited) != 5 {
		t.Fatalf("Iter visited %d entries, want 5", len(visited))
	}
	for i := uint32(1); i <= 5; i++ {
		if visited[i] != int(i*10) {
			t.Fatalf("entry %d = %d, want %d", i, visited[i], i*10)
		}
	}
}

func TestEntityMapClear(t *testing.T) {
	t.Parallel()

	m := NewEntityMap[string]()
	for i := uint32(1); i <= 3; i++ {
		m.Set(NewEntity(i, 1), "v")
	}
	m.Clear()
	if m.Len() != 0 {
		t.Fatalf("Len() after Clear = %d, want 0", m.Len())
	}
}

func TestNewEntityMapWithCapacity(t *testing.T) {
	t.Parallel()

	if m := NewEntityMapWithCapacity[int](-5); m == nil || m.Len() != 0 {
		t.Fatal("negative capacity must produce empty map")
	}
	m := NewEntityMapWithCapacity[int](32)
	if m.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", m.Len())
	}
}

func TestDisabledTagIsZeroSize(t *testing.T) {
	t.Parallel()

	if got := unsafe.Sizeof(DisabledTag{}); got != 0 {
		t.Fatalf("DisabledTag size = %d, want 0 (zero-size tag component)", got)
	}
}

func sortEntities(es []Entity) {
	sort.Slice(es, func(i, j int) bool { return es[i].ID() < es[j].ID() })
}
