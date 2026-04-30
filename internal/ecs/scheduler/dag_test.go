package scheduler

import (
	"errors"
	"strings"
	"testing"
)

func TestDAGEmptyBuild(t *testing.T) {
	t.Parallel()

	d := NewDAG(0)
	if err := d.Build(); err != nil {
		t.Fatalf("empty Build err = %v, want nil", err)
	}
	if got := d.TopologicalOrder(); len(got) != 0 {
		t.Fatalf("empty TopologicalOrder = %v, want []", got)
	}
}

func TestDAGSingleNode(t *testing.T) {
	t.Parallel()

	d := NewDAG(1)
	if err := d.Build(); err != nil {
		t.Fatalf("Build err = %v", err)
	}
	got := d.TopologicalOrder()
	if len(got) != 1 || got[0] != 0 {
		t.Fatalf("TopologicalOrder = %v, want [0]", got)
	}
}

func TestDAGLinearChain(t *testing.T) {
	t.Parallel()

	d := NewDAG(4)
	d.AddEdge(0, 1)
	d.AddEdge(1, 2)
	d.AddEdge(2, 3)
	if err := d.Build(); err != nil {
		t.Fatal(err)
	}
	got := d.TopologicalOrder()
	want := []SystemNodeID{0, 1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("Order len = %d, want %d", len(got), len(want))
	}
	for i, v := range want {
		if got[i] != v {
			t.Fatalf("Order[%d] = %d, want %d (full=%v)", i, got[i], v, got)
		}
	}
}

func TestDAGDeterministicOrderOnTies(t *testing.T) {
	t.Parallel()

	// Three roots with no constraints; Build must yield them in ID order.
	d := NewDAG(3)
	if err := d.Build(); err != nil {
		t.Fatal(err)
	}
	got := d.TopologicalOrder()
	for i, v := range []SystemNodeID{0, 1, 2} {
		if got[i] != v {
			t.Fatalf("Order[%d] = %d, want %d", i, got[i], v)
		}
	}
}

func TestDAGTopologicalOrderRespectsEdges(t *testing.T) {
	t.Parallel()

	// 0 → 2, 1 → 2, 2 → 3, 0 → 1
	// 0 must precede 1, 2; 1 must precede 2; 2 must precede 3.
	d := NewDAG(4)
	d.AddEdge(0, 2)
	d.AddEdge(1, 2)
	d.AddEdge(2, 3)
	d.AddEdge(0, 1)
	if err := d.Build(); err != nil {
		t.Fatal(err)
	}
	got := d.TopologicalOrder()
	pos := make(map[SystemNodeID]int, len(got))
	for i, n := range got {
		pos[n] = i
	}
	if pos[0] >= pos[1] || pos[0] >= pos[2] || pos[1] >= pos[2] || pos[2] >= pos[3] {
		t.Fatalf("Order %v violates edges", got)
	}
}

func TestDAGDuplicateEdgesIgnored(t *testing.T) {
	t.Parallel()

	d := NewDAG(2)
	d.AddEdge(0, 1)
	d.AddEdge(0, 1)
	d.AddEdge(0, 1)
	if err := d.Build(); err != nil {
		t.Fatal(err)
	}
	got := d.TopologicalOrder()
	if len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatalf("Order = %v, want [0 1]", got)
	}
}

func TestDAGSimpleCycleRejected(t *testing.T) {
	t.Parallel()

	d := NewDAG(2)
	d.AddEdge(0, 1)
	d.AddEdge(1, 0)
	err := d.Build()
	if err == nil {
		t.Fatal("expected cycle error")
	}
	if !errors.Is(err, ErrScheduleCycle) {
		t.Fatalf("err = %v, want errors.Is(ErrScheduleCycle)", err)
	}
	if !strings.Contains(err.Error(), "0") || !strings.Contains(err.Error(), "1") {
		t.Fatalf("cycle error must list participants; got %q", err.Error())
	}
}

func TestDAGSelfLoopRejected(t *testing.T) {
	t.Parallel()

	d := NewDAG(1)
	d.AddEdge(0, 0)
	err := d.Build()
	if err == nil {
		t.Fatal("self-loop must be rejected")
	}
	if !errors.Is(err, ErrScheduleCycle) {
		t.Fatalf("err = %v, want errors.Is(ErrScheduleCycle)", err)
	}
}

func TestDAGLargerCycleRejected(t *testing.T) {
	t.Parallel()

	// 0 → 1 → 2 → 3 → 1 (3-node cycle nested under 0).
	d := NewDAG(4)
	d.AddEdge(0, 1)
	d.AddEdge(1, 2)
	d.AddEdge(2, 3)
	d.AddEdge(3, 1)
	if err := d.Build(); !errors.Is(err, ErrScheduleCycle) {
		t.Fatalf("err = %v, want cycle", err)
	}
}

func TestDAGAddEdgeOutOfRangePanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("AddEdge with out-of-range id must panic")
		}
	}()
	d := NewDAG(2)
	d.AddEdge(0, 5)
}

func TestDAGTopologicalOrderBeforeBuildIsNil(t *testing.T) {
	t.Parallel()

	d := NewDAG(2)
	if d.TopologicalOrder() != nil {
		t.Fatal("TopologicalOrder before Build must be nil")
	}
}

func TestDAGRebuildAfterEdgeAdd(t *testing.T) {
	t.Parallel()

	d := NewDAG(3)
	d.AddEdge(0, 1)
	if err := d.Build(); err != nil {
		t.Fatal(err)
	}
	first := d.TopologicalOrder()
	if len(first) != 3 {
		t.Fatalf("first Order len = %d, want 3", len(first))
	}

	// Add an edge that flips ordering of 1 vs 2.
	d.AddEdge(2, 1)
	// Build must be re-runnable; old `built` state is invalidated.
	if err := d.Build(); err != nil {
		t.Fatal(err)
	}
	second := d.TopologicalOrder()
	pos := make(map[SystemNodeID]int)
	for i, n := range second {
		pos[n] = i
	}
	if pos[2] >= pos[1] {
		t.Fatalf("after rebuild expected pos[2] < pos[1]; got %v", second)
	}
}

func TestDAGHasEdge(t *testing.T) {
	t.Parallel()

	d := NewDAG(3)
	d.AddEdge(0, 1)
	if !d.HasEdge(0, 1) {
		t.Fatal("HasEdge(0,1) = false, want true")
	}
	if d.HasEdge(1, 0) {
		t.Fatal("HasEdge(1,0) = true, want false")
	}
	if d.HasEdge(0, 2) {
		t.Fatal("HasEdge(0,2) = true, want false")
	}
}

func TestDAGNodeCount(t *testing.T) {
	t.Parallel()

	d := NewDAG(7)
	if d.NodeCount() != 7 {
		t.Fatalf("NodeCount = %d, want 7", d.NodeCount())
	}
}
