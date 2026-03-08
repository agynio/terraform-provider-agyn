package teamapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Agent struct {
	ID          string
	Title       *string
	Description *string
	Config      json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type AgentCreate struct {
	Title       *string
	Description *string
	Config      json.RawMessage
}

type AgentUpdate struct {
	Title       *string
	Description *string
	Config      *json.RawMessage
}

func (c *Client) CreateAgent(ctx context.Context, input AgentCreate) (*Agent, error) {
	if len(input.Config) == 0 {
		return nil, fmt.Errorf("config is required")
	}
	if !json.Valid(input.Config) {
		return nil, fmt.Errorf("config must be valid JSON")
	}

	payload := map[string]any{
		"config": json.RawMessage(input.Config),
	}
	if input.Title != nil {
		payload["title"] = *input.Title
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal agent payload: %w", err)
	}

	return withConflictRetry(ctx, "create agent", func() (*Agent, error) {
		resp, err := c.raw.PostAgentsWithBodyWithResponse(ctx, "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("create agent request: %w", err)
		}

		if resp.JSON201 == nil {
			return nil, errorFromResponse("create agent", responseStatus(resp), resp.Body)
		}

		agent, err := mapAgent(resp.JSON201)
		if err != nil {
			return nil, fmt.Errorf("decode agent response: %w", err)
		}
		return agent, nil
	})
}

func (c *Client) GetAgent(ctx context.Context, id string) (*Agent, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	resp, err := c.raw.GetAgentsIdWithResponse(ctx, uuidValue)
	if err != nil {
		return nil, fmt.Errorf("get agent request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("get agent", responseStatus(resp), resp.Body)
	}

	agent, err := mapAgent(resp.JSON200)
	if err != nil {
		return nil, fmt.Errorf("decode agent response: %w", err)
	}
	return agent, nil
}

func (c *Client) UpdateAgent(ctx context.Context, id string, input AgentUpdate) (*Agent, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	payload := make(map[string]any)
	if input.Title != nil {
		payload["title"] = *input.Title
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}
	if input.Config != nil {
		if len(*input.Config) == 0 {
			return nil, fmt.Errorf("config must be valid JSON")
		}
		if !json.Valid(*input.Config) {
			return nil, fmt.Errorf("config must be valid JSON")
		}
		payload["config"] = json.RawMessage(*input.Config)
	}

	if len(payload) == 0 {
		return c.GetAgent(ctx, id)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal agent payload: %w", err)
	}

	resp, err := c.raw.PatchAgentsIdWithBodyWithResponse(ctx, uuidValue, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("update agent request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("update agent", responseStatus(resp), resp.Body)
	}

	agent, err := mapAgent(resp.JSON200)
	if err != nil {
		return nil, fmt.Errorf("decode agent response: %w", err)
	}
	return agent, nil
}

func (c *Client) DeleteAgent(ctx context.Context, id string) error {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return err
	}

	return withConflictRetryNoResult(ctx, "delete agent", func() error {
		resp, err := c.raw.DeleteAgentsIdWithResponse(ctx, uuidValue)
		if err != nil {
			return fmt.Errorf("delete agent request: %w", err)
		}

		if resp.StatusCode() == http.StatusNoContent {
			return nil
		}

		return errorFromResponse("delete agent", responseStatus(resp), resp.Body)
	})
}

type agentPayload struct {
	ID          string          `json:"id"`
	Title       *string         `json:"title,omitempty"`
	Description *string         `json:"description,omitempty"`
	Config      json.RawMessage `json:"config"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   *time.Time      `json:"updatedAt,omitempty"`
}

func mapAgent(source any) (*Agent, error) {
	var payload agentPayload
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &Agent{
		ID:          payload.ID,
		Title:       payload.Title,
		Description: payload.Description,
		Config:      payload.Config,
		CreatedAt:   payload.CreatedAt,
		UpdatedAt:   payload.UpdatedAt,
	}, nil
}
