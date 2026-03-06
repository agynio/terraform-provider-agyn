package teamapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
)

const (
	graphConflictRetryCount = 4
	graphConflictBaseDelay  = 200 * time.Millisecond
)

func isVersionConflict(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	if apiErr.Status != http.StatusConflict {
		return false
	}
	return strings.Contains(apiErr.Detail, "VERSION_CONFLICT")
}

func waitForConflictRetry(ctx context.Context, attempt int) error {
	if attempt <= 0 {
		return nil
	}

	delay := graphConflictBaseDelay * time.Duration(attempt)
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
