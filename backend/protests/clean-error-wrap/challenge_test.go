//go:build ignore

package challenge

import (
	"errors"
	"testing"
)

func TestWrapPreservesNil(t *testing.T) {
	if Wrap(nil, "noop") != nil {
		t.Fatal("expected wrapping nil to remain nil")
	}
	if Cause(nil) != nil {
		t.Fatal("expected cause of nil to be nil")
	}
}

func TestWrapChainsErrors(t *testing.T) {
	root := errors.New("disk failure")
	level1 := Wrap(root, "loading config")
	level2 := Wrap(level1, "startup")

	if level1 == nil || level2 == nil {
		t.Fatal("expected wrapped errors")
	}
	if !errors.Is(level2, root) {
		t.Fatalf("expected errors.Is to find root cause through wrappers")
	}
	if got := Cause(level2); got != root {
		t.Fatalf("expected Cause to return root error, got %v", got)
	}
	if got := Cause(level1); got != root {
		t.Fatalf("expected Cause to return root error, got %v", got)
	}
	wantMsg := "startup: loading config: disk failure"
	if level2.Error() != wantMsg {
		t.Fatalf("unexpected error string.\nwant: %q\ngot:  %q", wantMsg, level2.Error())
	}
}

func TestWrapKeepsContextOrder(t *testing.T) {
	root := errors.New("validation failed")
	wrapped := Wrap(Wrap(root, "parse request"), "api handler")

	if !errors.Is(wrapped, root) {
		t.Fatalf("expected errors.Is to find %q", root)
	}
	// ensure Cause walks entire chain
	if got := Cause(wrapped); got != root {
		t.Fatalf("expected root %v, got %v", root, got)
	}
}
