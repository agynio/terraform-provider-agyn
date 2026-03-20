package agentapi

import (
	"context"
	"fmt"
	"time"
)

const (
	conflictRetryCount = 4
	conflictBaseDelay  = 200 * time.Millisecond
)

func waitForConflictRetry(ctx context.Context, attempt int) error {
	if attempt <= 0 {
		return nil
	}

	delay := conflictBaseDelay * time.Duration(attempt)
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func withConflictRetry[T any](ctx context.Context, op string, fn func() (T, error)) (T, error) {
	var zero T
	var lastErr error
	for attempt := 0; attempt < conflictRetryCount; attempt++ {
		if err := waitForConflictRetry(ctx, attempt); err != nil {
			return zero, err
		}

		result, err := fn()
		if err != nil {
			if IsConflict(err) {
				lastErr = err
				continue
			}
			return zero, err
		}
		return result, nil
	}

	if lastErr != nil {
		return zero, lastErr
	}
	return zero, fmt.Errorf("%s failed after %d attempts", op, conflictRetryCount)
}

func withConflictRetryNoResult(ctx context.Context, op string, fn func() error) error {
	_, err := withConflictRetry(ctx, op, func() (struct{}, error) {
		return struct{}{}, fn()
	})
	return err
}
