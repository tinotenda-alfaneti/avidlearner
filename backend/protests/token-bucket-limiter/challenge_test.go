//go:build ignore

package challenge

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLimiterRespectsBurst(t *testing.T) {
	lim := New(5, 3)
	now := time.Now()

	for i := 0; i < 3; i++ {
		if !lim.Allow(now) {
			t.Fatalf("expected allow on token %d", i)
		}
	}
	if lim.Allow(now) {
		t.Fatal("expected deny once burst tokens exhausted")
	}
}

func TestLimiterRefillsOverTime(t *testing.T) {
	lim := New(2, 2)
	now := time.Now()

	if !lim.Allow(now) || !lim.Allow(now) {
		t.Fatal("expected initial tokens to succeed")
	}
	if lim.Allow(now) {
		t.Fatal("expected deny with zero tokens remaining")
	}

	now = now.Add(600 * time.Millisecond)
	if !lim.Allow(now) {
		t.Fatal("expected token after partial second")
	}

	now = now.Add(600 * time.Millisecond)
	if !lim.Allow(now) {
		t.Fatal("expected another token after second interval")
	}
}

func TestLimiterIsConcurrentSafe(t *testing.T) {
	lim := New(4, 4)
	var granted int32

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if lim.Allow(time.Now()) {
				atomic.AddInt32(&granted, 1)
			}
		}()
	}
	wg.Wait()

	if granted > 4 {
		t.Fatalf("expected at most 4 grants, got %d", granted)
	}
}
