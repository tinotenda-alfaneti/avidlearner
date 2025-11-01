//go:build ignore

package challenge

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type logEntry struct {
	method   string
	path     string
	status   int
	duration time.Duration
}

type captureLogger struct {
	entries []logEntry
}

func (c *captureLogger) Log(method, path string, status int, duration time.Duration) {
	c.entries = append(c.entries, logEntry{method: method, path: path, status: status, duration: duration})
}

func TestLoggingMiddlewareCapturesStatusAndDuration(t *testing.T) {
	logger := &captureLogger{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	})
	wrapped := Logging(logger, handler)

	req := httptest.NewRequest(http.MethodGet, "/demo", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected recorder status %d, got %d", http.StatusNoContent, rec.Code)
	}
	if len(logger.entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(logger.entries))
	}
	entry := logger.entries[0]
	if entry.method != http.MethodGet || entry.path != "/demo" {
		t.Fatalf("expected GET /demo, got %s %s", entry.method, entry.path)
	}
	if entry.status != http.StatusNoContent {
		t.Fatalf("expected logged status %d, got %d", http.StatusNoContent, entry.status)
	}
	if entry.duration < 15*time.Millisecond {
		t.Fatalf("expected duration to include handler time, got %v", entry.duration)
	}
}

func TestLoggingMiddlewareDefaultsStatus(t *testing.T) {
	logger := &captureLogger{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	wrapped := Logging(logger, handler)
	req := httptest.NewRequest(http.MethodPost, "/ok", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected recorder status %d, got %d", http.StatusOK, rec.Code)
	}
	if len(logger.entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(logger.entries))
	}
	entry := logger.entries[0]
	if entry.status != http.StatusOK {
		t.Fatalf("expected logged OK status, got %d", entry.status)
	}
	if entry.method != http.MethodPost || entry.path != "/ok" {
		t.Fatalf("expected POST /ok, got %s %s", entry.method, entry.path)
	}
	if entry.duration <= 0 {
		t.Fatalf("expected positive duration, got %v", entry.duration)
	}
}
