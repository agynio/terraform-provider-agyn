package teamapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Problem struct {
	Title    string  `json:"title"`
	Detail   *string `json:"detail,omitempty"`
	Status   int     `json:"status"`
	Type     *string `json:"type,omitempty"`
	Instance *string `json:"instance,omitempty"`
}

type APIError struct {
	Status int
	Title  string
	Detail string
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s (%d): %s", e.Title, e.Status, e.Detail)
	}
	return fmt.Sprintf("%s (%d)", e.Title, e.Status)
}

func errorFromResponse(op string, status int, body []byte) error {
	detail := http.StatusText(status)
	title := detail
	if title == "" {
		title = "API error"
	}
	if len(body) > 0 {
		if prob := parseProblem(body); prob != nil {
			title = prob.Title
			if prob.Detail != nil {
				detail = *prob.Detail
			}
		} else {
			detail = string(body)
		}
	}
	return fmt.Errorf("%s: %w", op, &APIError{Status: status, Title: title, Detail: detail})
}

func parseProblem(body []byte) *Problem {
	var prob Problem
	if err := json.Unmarshal(body, &prob); err != nil {
		return nil
	}
	if prob.Title == "" && prob.Status == 0 {
		return nil
	}
	return &prob
}
