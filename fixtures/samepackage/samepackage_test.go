package samepackage

import "testing"

// A white-box test in the interface's own package can use fakes that were
// generated into that package without an import cycle.
func TestUseWithSamePackageFakes(t *testing.T) {
	w := &FakeWidget{}
	w.DoReturns(Thing{Name: "out"}, nil)
	g := &fakeGadget{}

	got, err := Use(w, g)
	if err != nil {
		t.Fatalf("Use returned error: %v", err)
	}
	if got != "out" {
		t.Fatalf("Use returned %q, want %q", got, "out")
	}
	if w.DoCallCount() != 1 || g.SpinCallCount() != 1 {
		t.Fatalf("expected one call each, got Do=%d Spin=%d", w.DoCallCount(), g.SpinCallCount())
	}
	if g.SpinArgsForCall(0) != 3 {
		t.Fatalf("Spin called with %d, want 3", g.SpinArgsForCall(0))
	}

	var asGadget gadget = g
	asGadget.stop()
	if g.StopCallCount() != 1 {
		t.Fatalf("expected stop to be recorded once, got %d", g.StopCallCount())
	}
}
