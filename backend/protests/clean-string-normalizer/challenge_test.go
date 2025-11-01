//go:build ignore

package challenge

import "testing"

func TestNormalizeProducesCleanSentences(t *testing.T) {
	input := "   go   is,  awesome!   concurrency\tis hard. clean code?  yes! "
	want := "Go is, awesome! Concurrency is hard. Clean code? Yes!"

	if got := Normalize(input); got != want {
		t.Fatalf("unexpected normalization result\nwant: %q\ngot:  %q", want, got)
	}
}

func TestNormalizeCollapsesWhitespace(t *testing.T) {
	input := "\n\nmultiple   spaces\tand newlines\nshould  fold."
	want := "Multiple spaces and newlines should fold."
	if got := Normalize(input); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestNormalizeIdempotent(t *testing.T) {
	input := "Clean Code Matters."
	if got := Normalize(Normalize(input)); got != input {
		t.Fatalf("expected idempotent normalization, got %q", got)
	}
}
