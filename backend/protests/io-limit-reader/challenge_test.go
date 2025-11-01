//go:build ignore

package challenge

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestLimitReaderStopsAtLimit(t *testing.T) {
	src := strings.NewReader("abcdef")
	lr := &LimitReader{R: src, N: 3}

	buf := make([]byte, 2)
	n, err := lr.Read(buf)
	if err != nil {
		t.Fatalf("first read unexpected error: %v", err)
	}
	if n != 2 || string(buf[:n]) != "ab" {
		t.Fatalf("expected first chunk 'ab', got %q (%d)", string(buf[:n]), n)
	}

	n, err = lr.Read(buf)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF after hitting limit, got %v", err)
	}
	if n != 1 || string(buf[:n]) != "c" {
		t.Fatalf("expected single byte 'c', got %q (%d)", string(buf[:n]), n)
	}
}

func TestLimitReaderPassesThroughEOF(t *testing.T) {
	src := strings.NewReader("go")
	lr := &LimitReader{R: src, N: 10}

	got, err := io.ReadAll(lr)
	if err != nil {
		t.Fatalf("read all: %v", err)
	}
	if string(got) != "go" {
		t.Fatalf("expected to read entire source, got %q", string(got))
	}
}

func TestLimitReaderHandlesNilReader(t *testing.T) {
	var lr LimitReader
	n, err := lr.Read(make([]byte, 8))
	if n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("expected 0, EOF for nil reader, got %d, %v", n, err)
	}
}

func TestLimitReaderNoExtraReads(t *testing.T) {
	src := bytes.NewBufferString("hello world")
	counting := &countReader{r: src}
	lr := &LimitReader{R: counting, N: 5}

	_, err := io.ReadAll(lr)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("unexpected error: %v", err)
	}
	if counting.reads > 3 {
		t.Fatalf("expected minimal read calls, got %d", counting.reads)
	}
}

type countReader struct {
	r     io.Reader
	reads int
}

func (c *countReader) Read(p []byte) (int, error) {
	c.reads++
	return c.r.Read(p)
}
