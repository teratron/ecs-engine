package world

import "testing"

func TestRegisterDeferredFlusher_NilIgnored(t *testing.T) {
	t.Parallel()
	w := NewWorld()
	w.RegisterDeferredFlusher(nil)
	w.ApplyDeferred() // must not panic
}

func TestApplyDeferred_NoFlushers(t *testing.T) {
	t.Parallel()
	w := NewWorld()
	w.ApplyDeferred() // must not panic when nothing is registered
}

func TestApplyDeferred_RunsAllInOrder(t *testing.T) {
	t.Parallel()
	w := NewWorld()
	var order []int
	w.RegisterDeferredFlusher(func(_ *World) { order = append(order, 1) })
	w.RegisterDeferredFlusher(func(_ *World) { order = append(order, 2) })
	w.RegisterDeferredFlusher(func(_ *World) { order = append(order, 3) })

	w.ApplyDeferred()

	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Fatalf("flushers must run in registration order; got %v", order)
	}
}

func TestResetDeferredFlushers(t *testing.T) {
	t.Parallel()
	w := NewWorld()
	called := 0
	w.RegisterDeferredFlusher(func(_ *World) { called++ })
	w.ApplyDeferred()
	if called != 1 {
		t.Fatalf("called = %d, want 1", called)
	}

	w.ResetDeferredFlushers()
	w.ApplyDeferred()
	if called != 1 {
		t.Fatalf("after Reset, ApplyDeferred must not invoke dropped flushers (called=%d)", called)
	}
}

func TestDeferredWorld_ApplyDeferredDelegates(t *testing.T) {
	t.Parallel()
	w := NewWorld()
	called := false
	w.RegisterDeferredFlusher(func(_ *World) { called = true })

	dw := w.NewDeferred()
	dw.ApplyDeferred()
	if !called {
		t.Fatal("DeferredWorld.ApplyDeferred must delegate to World.ApplyDeferred")
	}
}
