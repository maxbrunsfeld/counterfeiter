package externaltest_test

import (
	"testing"

	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/externaltest"
	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/samepackage"
)

// A black-box test in the external test package uses the fakes that were
// generated into that package without qualifying them.
func TestUseWithExternalTestPackageFakes(t *testing.T) {
	thing := &FakeThing{}
	thing.DoReturns(42, nil)

	got, err := externaltest.Use(thing)
	if err != nil {
		t.Fatalf("Use returned error: %v", err)
	}
	if got != 42 {
		t.Fatalf("Use returned %d, want 42", got)
	}
	if thing.DoCallCount() != 1 || thing.DoArgsForCall(0) != "thing" {
		t.Fatalf("unexpected Do calls: %d, %q", thing.DoCallCount(), thing.DoArgsForCall(0))
	}
}

func TestFakesOfOtherPackages(t *testing.T) {
	var _ samepackage.Widget = &FakeWidget{}

	wc := &FakeWriteCloser{}
	if _, err := wc.Write([]byte("x")); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if wc.WriteCallCount() != 1 {
		t.Fatalf("expected one Write call, got %d", wc.WriteCallCount())
	}

	m := &FakeGomegaMatcher{}
	m.MatchReturns(true, nil)
	ok, err := m.Match("anything")
	if err != nil || !ok {
		t.Fatalf("Match returned %v, %v", ok, err)
	}
}
