//go:build ignore

package challenge

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFanProcessesValues(t *testing.T) {
	ctx := context.Background()
	in := make(chan int)
	out := Fan(ctx, in, 3, func(ctx context.Context, v int) (int, error) {
		time.Sleep(5 * time.Millisecond)
		return v * 2, nil
	})

	go func() {
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)
	}()

	results := map[int]int{}
	for res := range out {
		if res.Err != nil {
			t.Fatalf("unexpected error: %v", res.Err)
		}
		results[res.Value]++
	}

	if len(results) != 5 {
		t.Fatalf("expected 5 unique results, got %d", len(results))
	}
	for i := 1; i <= 5; i++ {
		if results[i*2] != 1 {
			t.Fatalf("expected result %d once, got %d", i*2, results[i*2])
		}
	}
}

func TestFanCancelsWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan int)
	out := Fan(ctx, in, 4, func(ctx context.Context, v int) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return v, nil
		}
	})

	go func() {
		for i := 0; i < 6; i++ {
			in <- i
		}
		close(in)
	}()

	done := make(chan struct{})
	var seen int
	go func() {
		defer close(done)
		for res := range out {
			seen++
			if res.Err != nil && !errors.Is(res.Err, context.Canceled) {
				t.Errorf("unexpected worker error: %v", res.Err)
			}
			if seen == 2 {
				cancel()
			}
		}
	}()

	select {
	case <-done:
	case <-time.After(400 * time.Millisecond):
		t.Fatal("fan did not close output after cancellation")
	}
	if seen == 0 {
		t.Fatal("expected to see some results before cancellation")
	}
}
