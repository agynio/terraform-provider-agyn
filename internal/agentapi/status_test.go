package agentapi

import "testing"

type fakeStatus struct{ code int }

func (f fakeStatus) StatusCode() int { return f.code }

func TestResponseStatus(t *testing.T) {
	if got := responseStatus(nil); got != 500 {
		t.Fatalf("expected 500 for nil response, got %d", got)
	}

	if got := responseStatus(fakeStatus{code: 204}); got != 204 {
		t.Fatalf("expected 204, got %d", got)
	}

	if got := responseStatus(fakeStatus{}); got != 500 {
		t.Fatalf("expected fallback 500, got %d", got)
	}
}
