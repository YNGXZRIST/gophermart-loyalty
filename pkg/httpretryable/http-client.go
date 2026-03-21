package httpretryable

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

const defaultTimeout = 10 * time.Second

type RetryableClient struct {
	RetryMax           int
	client             *http.Client
	rateBeforeUnixNano atomic.Int64
}

func NewRetryableClient() *RetryableClient {
	client := &http.Client{}
	return &RetryableClient{client: client}
}
func (c *RetryableClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	n := c.rateBeforeUnixNano.Load()
	if n > 0 {
		rateBefore := time.Unix(0, n)
		if rateBefore.After(time.Now()) {
			timer := time.NewTimer(time.Until(rateBefore))
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, fmt.Errorf("timed out waiting for rate before %s", rateBefore)
			case <-timer.C:
			}
		}
	}

	for i := 0; i < c.RetryMax; i++ {
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}
		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			err := c.processRetryTimeout(resp)
			if err != nil {
				return nil, fmt.Errorf("error processing retryable response: %s", err)
			}
			_ = resp.Body.Close()
			continue
		case http.StatusRequestTimeout,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout:
			_ = resp.Body.Close()
			continue
		default:
			return resp, nil
		}
	}
	return nil, fmt.Errorf("http retry max retries exceeded")
}
func (c *RetryableClient) processRetryTimeout(resp *http.Response) error {
	delay := defaultTimeout
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		r, err := strconv.Atoi(retryAfter)
		if err != nil {
			return fmt.Errorf("http retry max retries parse error: %w", err)
		}
		delay = time.Duration(r) * time.Second
	}
	c.bumpRateBefore(time.Now().Add(delay))
	return nil
}
func (c *RetryableClient) bumpRateBefore(until time.Time) {
	newV := until.UnixNano()
	for {
		oldV := c.rateBeforeUnixNano.Load()
		if oldV >= newV {
			return
		}
		if c.rateBeforeUnixNano.CompareAndSwap(oldV, newV) {
			return
		}
	}
}
