package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	retryMaxRetries = 3
	retryBaseDelay  = 250 * time.Millisecond
	retryMaxDelay   = 2 * time.Second
)

func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

func DoWithRetry(
	ctx context.Context,
	client *http.Client,
	makeReq func() (*http.Request, error),
	formatAPIError func(status int, body []byte) error,
) (*http.Response, error) {
	if client == nil {
		client = http.DefaultClient
	}

	var lastErr error

	for attempt := 0; attempt <= retryMaxRetries; attempt++ {
		req, err := makeReq()
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("execute request: %w", err)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if formatAPIError != nil {
			lastErr = formatAPIError(resp.StatusCode, body)
		} else {
			lastErr = fmt.Errorf("request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		if lastErr == nil {
			lastErr = fmt.Errorf("request failed with status %d", resp.StatusCode)
		}

		if !isRetryableStatus(resp.StatusCode) || attempt == retryMaxRetries {
			return nil, lastErr
		}

		if err := sleepWithBackoff(ctx, attempt); err != nil {
			return nil, err
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("request failed")
}

func isRetryableStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway,
		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func sleepWithBackoff(ctx context.Context, attempt int) error {
	delay := retryBaseDelay * time.Duration(1<<attempt)
	if delay > retryMaxDelay {
		delay = retryMaxDelay
	}

	jitter := time.Duration(rand.Int63n(int64(delay/2) + 1))
	delay = delay + jitter
	if delay > retryMaxDelay {
		delay = retryMaxDelay
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
