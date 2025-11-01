//go:build ignore

package challenge

import "testing"

func TestMapInts(t *testing.T) {
	in := []int{1, 2, 3, 4}
	out := Map(in, func(v int) int { return v * v })

	if len(out) != len(in) {
		t.Fatalf("expected len %d, got %d", len(in), len(out))
	}
	if cap(out) != len(in) {
		t.Fatalf("expected cap %d, got %d", len(in), cap(out))
	}
	for i, v := range in {
		if out[i] != v*v {
			t.Fatalf("expected %d, got %d", v*v, out[i])
		}
	}
	// ensure input not mutated
	want := []int{1, 2, 3, 4}
	for i := range in {
		if in[i] != want[i] {
			t.Fatalf("input mutated at %d: want %d, got %d", i, want[i], in[i])
		}
	}
}

func TestMapDifferentTypes(t *testing.T) {
	in := []string{"go", "lang", "rocks"}
	out := Map(in, func(s string) int { return len(s) })
	want := []int{2, 4, 5}
	if len(out) != len(want) {
		t.Fatalf("expected len %d, got %d", len(want), len(out))
	}
	for i, v := range want {
		if out[i] != v {
			t.Fatalf("index %d: want %d, got %d", i, v, out[i])
		}
	}
}

func TestFilterBasics(t *testing.T) {
	in := []int{1, 2, 3, 4, 5}
	out := Filter(in, func(v int) bool { return v%2 == 0 })

	want := []int{2, 4}
	if len(out) != len(want) {
		t.Fatalf("expected len %d, got %d", len(want), len(out))
	}
	if cap(out) != len(in) {
		t.Fatalf("expected capacity reused len(in)=%d, got %d", len(in), cap(out))
	}
	for i, v := range want {
		if out[i] != v {
			t.Fatalf("index %d: want %d, got %d", i, v, out[i])
		}
	}
}

func TestFilterStructs(t *testing.T) {
	type user struct {
		Name string
		Role string
	}
	in := []user{
		{"Ana", "admin"},
		{"Bob", "user"},
		{"Cyd", "admin"},
	}
	out := Filter(in, func(u user) bool { return u.Role == "admin" })
	if len(out) != 2 {
		t.Fatalf("expected 2 admins, got %d", len(out))
	}
	if out[0].Name != "Ana" || out[1].Name != "Cyd" {
		t.Fatalf("unexpected ordering: %#v", out)
	}
}
