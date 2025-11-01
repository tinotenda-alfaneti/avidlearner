//go:build ignore

package challenge

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolProcessesTasks(t *testing.T) {
	p := NewPool(3, 5)
	t.Cleanup(p.Close)

	var count int32
	ctx := context.Background()
	for i := 0; i < 12; i++ {
		if err := p.Submit(ctx, func(context.Context) error {
			atomic.AddInt32(&count, 1)
			return nil
		}); err != nil {
			t.Fatalf("submit %d: %v", i, err)
		}
	}
	p.Close()

	if got := atomic.LoadInt32(&count); got != 12 {
		t.Fatalf("expected 12 tasks executed, got %d", got)
	}
}

func TestPoolBackpressureBlocks(t *testing.T) {
	p := NewPool(1, 1)
	defer p.Close()

	block := make(chan struct{})
	err := p.Submit(context.Background(), func(ctx context.Context) error {
		<-block
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected submit error: %v", err)
	}
	if err := p.Submit(context.Background(), func(context.Context) error { return nil }); err != nil {
		t.Fatalf("unexpected submit error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()

	result := make(chan error, 1)
	go func() {
		result <- p.Submit(ctx, func(context.Context) error { return nil })
	}()

	select {
	case err := <-result:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected deadline exceeded, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("submit did not respect context deadline")
	}

	close(block) // free first worker
}

func TestPoolCloseDrains(t *testing.T) {
	p := NewPool(2, 4)

	var count int32
	ctx := context.Background()
	for i := 0; i < 6; i++ {
		if err := p.Submit(ctx, func(context.Context) error {
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&count, 1)
			return nil
		}); err != nil {
			t.Fatalf("submit %d: %v", i, err)
		}
	}
	p.Close()

	if got := atomic.LoadInt32(&count); got != 6 {
		t.Fatalf("expected all tasks to complete before close returned, got %d", got)
	}
}
