package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"lesson-plan/backend/internal/config"
	"lesson-plan/backend/internal/middleware"
	"lesson-plan/backend/internal/observability"
)

const (
	agentRequestRetryMax       = 2
	agentRequestRetryBaseDelay = 250 * time.Millisecond
)

func newAgentHTTPClient(cfg *config.AgentConfig) *http.Client {
	timeout := cfg.TimeoutDuration()
	if timeout <= 0 {
		timeout = 120 * time.Second
	}

	return &http.Client{
		Timeout: timeout,
	}
}

func retryableStatusCode(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= 500
}

func retryableError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout() || netErr.Temporary()
	}

	return true
}

func sleepWithContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func doAgentRequestWithRetry(
	ctx context.Context,
	httpClient *http.Client,
	method string,
	url string,
	body []byte,
	headers map[string]string,
	operation string,
) (int, []byte, error) {
	for attempt := 0; attempt <= agentRequestRetryMax; attempt++ {
		requestBody := bytes.NewReader(body)
		req, err := http.NewRequestWithContext(ctx, method, url, requestBody)
		if err != nil {
			return 0, nil, err
		}

		for key, value := range headers {
			req.Header.Set(key, value)
		}

		traceID := middleware.TraceIDFromContext(ctx)
		if traceID != "" {
			req.Header.Set(middleware.TraceIDHeader, traceID)
			req.Header.Set(middleware.RequestIDHeader, traceID)
		}

		start := time.Now()
		resp, err := httpClient.Do(req)
		latency := time.Since(start)

		if err != nil {
			observability.RecordDownstream("agent", operation, 0, latency)
			if attempt < agentRequestRetryMax && retryableError(err) {
				backoff := agentRequestRetryBaseDelay * time.Duration(1<<attempt)
				if !sleepWithContext(ctx, backoff) {
					return 0, nil, ctx.Err()
				}
				continue
			}
			return 0, nil, err
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			observability.RecordDownstream("agent", operation, resp.StatusCode, latency)
			if attempt < agentRequestRetryMax {
				backoff := agentRequestRetryBaseDelay * time.Duration(1<<attempt)
				if !sleepWithContext(ctx, backoff) {
					return 0, nil, ctx.Err()
				}
				continue
			}
			return resp.StatusCode, nil, readErr
		}

		observability.RecordDownstream("agent", operation, resp.StatusCode, latency)
		if attempt < agentRequestRetryMax && retryableStatusCode(resp.StatusCode) {
			backoff := agentRequestRetryBaseDelay * time.Duration(1<<attempt)
			if !sleepWithContext(ctx, backoff) {
				return 0, nil, ctx.Err()
			}
			continue
		}

		return resp.StatusCode, respBody, nil
	}

	return 0, nil, context.DeadlineExceeded
}
