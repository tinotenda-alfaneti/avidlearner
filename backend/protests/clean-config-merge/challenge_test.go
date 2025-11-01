//go:build ignore

package challenge

import (
	"reflect"
	"testing"
	"time"
)

func TestMergeOverridesMeaningfulValues(t *testing.T) {
	base := Config{
		Timeout:  2 * time.Second,
		Endpoint: "https://api.internal",
		Retries:  3,
		Tags:     []string{"core", "stable"},
	}
	override := Config{
		Timeout:  5 * time.Second,
		Endpoint: "",
		Retries:  0,
		Tags:     []string{"stable", "beta", "beta"},
	}

	got := Merge(base, override)

	if got.Timeout != 5*time.Second {
		t.Fatalf("expected timeout override to win, got %v", got.Timeout)
	}
	if got.Endpoint != base.Endpoint {
		t.Fatalf("expected empty override endpoint to keep base, got %q", got.Endpoint)
	}
	if got.Retries != base.Retries {
		t.Fatalf("expected retries to remain %d, got %d", base.Retries, got.Retries)
	}

	wantTags := []string{"core", "stable", "beta"}
	if !reflect.DeepEqual(got.Tags, wantTags) {
		t.Fatalf("expected tags %v, got %v", wantTags, got.Tags)
	}
}

func TestMergeDoesNotMutateInputs(t *testing.T) {
	base := Config{
		Timeout:  500 * time.Millisecond,
		Endpoint: "svc",
		Retries:  2,
		Tags:     []string{"svc", "edge"},
	}
	override := Config{
		Timeout:  0,
		Endpoint: "svc.v2",
		Retries:  5,
		Tags:     []string{"edge", "fast"},
	}

	baseCopy := deepCopyConfig(base)
	overrideCopy := deepCopyConfig(override)

	got := Merge(base, override)

	if !reflect.DeepEqual(base, baseCopy) {
		t.Fatal("base config mutated by Merge")
	}
	if !reflect.DeepEqual(override, overrideCopy) {
		t.Fatal("override config mutated by Merge")
	}
	if len(got.Tags) == 0 {
		t.Fatal("expected merged tags")
	}
	if &got.Tags[0] == &base.Tags[0] {
		t.Fatal("expected merged tags to be a new slice, but shares backing array with base")
	}
	if &got.Tags[len(got.Tags)-1] == &override.Tags[len(override.Tags)-1] {
		t.Fatal("expected merged tags to be a new slice, but shares backing array with override")
	}
	if got.Endpoint != override.Endpoint {
		t.Fatalf("expected override endpoint to win when provided, got %q", got.Endpoint)
	}
	if got.Retries != override.Retries {
		t.Fatalf("expected override retries to win when non-zero, got %d", got.Retries)
	}
}

func deepCopyConfig(c Config) Config {
	copyTags := append([]string(nil), c.Tags...)
	c.Tags = copyTags
	return c
}
