package agentapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type InitScript struct {
	ID          string
	Script      string
	Description *string
	AgentID     *string
	McpID       *string
	HookID      *string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type InitScriptCreate struct {
	Script      string
	Description *string
	AgentID     *string
	McpID       *string
	HookID      *string
}

type InitScriptUpdate struct {
	Script      *string
	Description *string
}

func (c *Client) CreateInitScript(ctx context.Context, input InitScriptCreate) (*InitScript, error) {
	if input.Script == "" {
		return nil, fmt.Errorf("script is required")
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

	payload := map[string]any{
		"script": input.Script,
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

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal init script payload: %w", err)
	}

	return withConflictRetry(ctx, "create init script", func() (*InitScript, error) {
		resp, err := c.raw.PostInitScriptsWithBodyWithResponse(ctx, "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("create init script request: %w", err)
		}

		if resp.JSON201 == nil {
			return nil, errorFromResponse("create init script", responseStatus(resp), resp.Body)
		}

		script, err := mapInitScript(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decode init script response: %w", err)
		}
		return script, nil
	})
}

func (c *Client) GetInitScript(ctx context.Context, id string) (*InitScript, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	resp, err := c.raw.GetInitScriptsIdWithResponse(ctx, uuidValue)
	if err != nil {
		return nil, fmt.Errorf("get init script request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("get init script", responseStatus(resp), resp.Body)
	}

	script, err := mapInitScript(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode init script response: %w", err)
	}
	return script, nil
}

func (c *Client) UpdateInitScript(ctx context.Context, id string, input InitScriptUpdate) (*InitScript, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	payload := make(map[string]any)
	if input.Script != nil {
		payload["script"] = *input.Script
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}

	if len(payload) == 0 {
		return c.GetInitScript(ctx, id)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal init script payload: %w", err)
	}

	resp, err := c.raw.PatchInitScriptsIdWithBodyWithResponse(ctx, uuidValue, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("update init script request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("update init script", responseStatus(resp), resp.Body)
	}

	script, err := mapInitScript(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode init script response: %w", err)
	}
	return script, nil
}

func (c *Client) DeleteInitScript(ctx context.Context, id string) error {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return err
	}

	return withConflictRetryNoResult(ctx, "delete init script", func() error {
		resp, err := c.raw.DeleteInitScriptsIdWithResponse(ctx, uuidValue)
		if err != nil {
			return fmt.Errorf("delete init script request: %w", err)
		}

		if resp.StatusCode() == http.StatusNoContent {
			return nil
		}

		return errorFromResponse("delete init script", responseStatus(resp), resp.Body)
	})
}

type initScriptPayload struct {
	ID          string     `json:"id"`
	Script      string     `json:"script"`
	Description *string    `json:"description,omitempty"`
	AgentID     *string    `json:"agentId,omitempty"`
	McpID       *string    `json:"mcpId,omitempty"`
	HookID      *string    `json:"hookId,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

func mapInitScript(source any) (*InitScript, error) {
	var payload initScriptPayload
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &InitScript{
		ID:          payload.ID,
		Script:      payload.Script,
		Description: payload.Description,
		AgentID:     payload.AgentID,
		McpID:       payload.McpID,
		HookID:      payload.HookID,
		CreatedAt:   payload.CreatedAt,
		UpdatedAt:   payload.UpdatedAt,
	}, nil
}
