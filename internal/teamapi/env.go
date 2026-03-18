package teamapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Env struct {
	ID          string
	Name        string
	Description *string
	AgentID     *string
	McpID       *string
	HookID      *string
	Value       *string
	SecretID    *string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type EnvCreate struct {
	Name        string
	Description *string
	AgentID     *string
	McpID       *string
	HookID      *string
	Value       *string
	SecretID    *string
}

type EnvUpdate struct {
	Name        *string
	Description *string
	Value       *string
	SecretID    *string
}

func (c *Client) CreateEnv(ctx context.Context, input EnvCreate) (*Env, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	ownerCount := 0
	if input.AgentID != nil {
		ownerCount++
		if _, err := parseUUID(*input.AgentID); err != nil {
			return nil, fmt.Errorf("invalid agent_id: %w", err)
		}
	}
	if input.McpID != nil {
		ownerCount++
		if _, err := parseUUID(*input.McpID); err != nil {
			return nil, fmt.Errorf("invalid mcp_id: %w", err)
		}
	}
	if input.HookID != nil {
		ownerCount++
		if _, err := parseUUID(*input.HookID); err != nil {
			return nil, fmt.Errorf("invalid hook_id: %w", err)
		}
	}
	if ownerCount != 1 {
		return nil, fmt.Errorf("exactly one of agent_id, mcp_id, or hook_id must be set")
	}

	valueCount := 0
	if input.Value != nil {
		valueCount++
	}
	if input.SecretID != nil {
		valueCount++
		if _, err := parseUUID(*input.SecretID); err != nil {
			return nil, fmt.Errorf("invalid secret_id: %w", err)
		}
	}
	if valueCount != 1 {
		return nil, fmt.Errorf("exactly one of value or secret_id must be set")
	}

	payload := map[string]any{
		"name": input.Name,
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}
	if input.AgentID != nil {
		payload["agentId"] = *input.AgentID
	}
	if input.McpID != nil {
		payload["mcpId"] = *input.McpID
	}
	if input.HookID != nil {
		payload["hookId"] = *input.HookID
	}
	if input.Value != nil {
		payload["value"] = *input.Value
	}
	if input.SecretID != nil {
		payload["secretId"] = *input.SecretID
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal env payload: %w", err)
	}

	return withConflictRetry(ctx, "create env", func() (*Env, error) {
		resp, err := c.raw.PostEnvsWithBodyWithResponse(ctx, "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("create env request: %w", err)
		}

		if resp.JSON201 == nil {
			return nil, errorFromResponse("create env", responseStatus(resp), resp.Body)
		}

		env, err := mapEnv(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decode env response: %w", err)
		}
		return env, nil
	})
}

func (c *Client) GetEnv(ctx context.Context, id string) (*Env, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	resp, err := c.raw.GetEnvsIdWithResponse(ctx, uuidValue)
	if err != nil {
		return nil, fmt.Errorf("get env request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("get env", responseStatus(resp), resp.Body)
	}

	env, err := mapEnv(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode env response: %w", err)
	}
	return env, nil
}

func (c *Client) UpdateEnv(ctx context.Context, id string, input EnvUpdate) (*Env, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	payload := make(map[string]any)
	if input.Name != nil {
		payload["name"] = *input.Name
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}
	if input.Value != nil {
		payload["value"] = *input.Value
	}
	if input.SecretID != nil {
		if _, err := parseUUID(*input.SecretID); err != nil {
			return nil, fmt.Errorf("invalid secret_id: %w", err)
		}
		payload["secretId"] = *input.SecretID
	}

	if len(payload) == 0 {
		return c.GetEnv(ctx, id)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal env payload: %w", err)
	}

	resp, err := c.raw.PatchEnvsIdWithBodyWithResponse(ctx, uuidValue, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("update env request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("update env", responseStatus(resp), resp.Body)
	}

	env, err := mapEnv(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode env response: %w", err)
	}
	return env, nil
}

func (c *Client) DeleteEnv(ctx context.Context, id string) error {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return err
	}

	return withConflictRetryNoResult(ctx, "delete env", func() error {
		resp, err := c.raw.DeleteEnvsIdWithResponse(ctx, uuidValue)
		if err != nil {
			return fmt.Errorf("delete env request: %w", err)
		}

		if resp.StatusCode() == http.StatusNoContent {
			return nil
		}

		return errorFromResponse("delete env", responseStatus(resp), resp.Body)
	})
}

type envPayload struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	AgentID     *string    `json:"agentId,omitempty"`
	McpID       *string    `json:"mcpId,omitempty"`
	HookID      *string    `json:"hookId,omitempty"`
	Value       *string    `json:"value,omitempty"`
	SecretID    *string    `json:"secretId,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

func mapEnv(source any) (*Env, error) {
	var payload envPayload
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &Env{
		ID:          payload.ID,
		Name:        payload.Name,
		Description: payload.Description,
		AgentID:     payload.AgentID,
		McpID:       payload.McpID,
		HookID:      payload.HookID,
		Value:       payload.Value,
		SecretID:    payload.SecretID,
		CreatedAt:   payload.CreatedAt,
		UpdatedAt:   payload.UpdatedAt,
	}, nil
}
