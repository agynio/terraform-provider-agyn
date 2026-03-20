package agentapi

import "net/http"

type statusProvider interface {
	StatusCode() int
}

func responseStatus(resp statusProvider) int {
	if resp == nil {
		return http.StatusInternalServerError
	}
	if code := resp.StatusCode(); code != 0 {
		return code
	}
	return http.StatusInternalServerError
}
