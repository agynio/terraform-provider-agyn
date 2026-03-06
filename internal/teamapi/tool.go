package teamapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Tool struct {
	ID          string
	Name        *string
	Description *string
	Type        string
	Config      *json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type ToolCreate struct {
	Type        string
	Name        *string
	Description *string
	Config      *json.RawMessage
}

type ToolUpdate struct {
	Name        *string
	Description *string
	Config      *json.RawMessage
}

func (c *Client) CreateTool(ctx context.Context, input ToolCreate) (*Tool, error) {
	if input.Type == "" {
		return nil, fmt.Errorf("type is required")
	}

	payload := map[string]any{
		"type": input.Type,
	}
	if input.Name != nil {
		payload["name"] = *input.Name
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}
	if input.Config != nil {
		if !json.Valid(*input.Config) {
			return nil, fmt.Errorf("config must be valid JSON")
		}
		payload["config"] = json.RawMessage(*input.Config)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal tool payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < graphConflictRetryCount; attempt++ {
		if err := waitForConflictRetry(ctx, attempt); err != nil {
			return nil, err
		}

		resp, err := c.raw.PostToolsWithBodyWithResponse(ctx, "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("create tool request: %w", err)
		}

		if resp.JSON201 == nil {
			err := errorFromResponse("create tool", responseStatus(resp), resp.Body)
			if isVersionConflict(err) {
				lastErr = err
				continue
			}
			return nil, err
		}

		tool, err := mapTool(resp.JSON201)
		if err != nil {
			return nil, fmt.Errorf("decode tool response: %w", err)
		}
		return tool, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("create tool failed after %d attempts", graphConflictRetryCount)
}

func (c *Client) GetTool(ctx context.Context, id string) (*Tool, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	resp, err := c.raw.GetToolsIdWithResponse(ctx, uuidValue)
	if err != nil {
		return nil, fmt.Errorf("get tool request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("get tool", responseStatus(resp), resp.Body)
	}

	tool, err := mapTool(resp.JSON200)
	if err != nil {
		return nil, fmt.Errorf("decode tool response: %w", err)
	}
	return tool, nil
}

func (c *Client) UpdateTool(ctx context.Context, id string, input ToolUpdate) (*Tool, error) {
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
	if input.Config != nil {
		if !json.Valid(*input.Config) {
			return nil, fmt.Errorf("config must be valid JSON")
		}
		payload["config"] = json.RawMessage(*input.Config)
	}

	if len(payload) == 0 {
		return c.GetTool(ctx, id)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal tool payload: %w", err)
	}

	resp, err := c.raw.PatchToolsIdWithBodyWithResponse(ctx, uuidValue, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("update tool request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("update tool", responseStatus(resp), resp.Body)
	}

	tool, err := mapTool(resp.JSON200)
	if err != nil {
		return nil, fmt.Errorf("decode tool response: %w", err)
	}
	return tool, nil
}

func (c *Client) DeleteTool(ctx context.Context, id string) error {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt < graphConflictRetryCount; attempt++ {
		if err := waitForConflictRetry(ctx, attempt); err != nil {
			return err
		}

		resp, err := c.raw.DeleteToolsIdWithResponse(ctx, uuidValue)
		if err != nil {
			return fmt.Errorf("delete tool request: %w", err)
		}

		if resp.StatusCode() == http.StatusNoContent {
			return nil
		}

		err = errorFromResponse("delete tool", responseStatus(resp), resp.Body)
		if isVersionConflict(err) {
			lastErr = err
			continue
		}
		return err
	}

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("delete tool failed after %d attempts", graphConflictRetryCount)
}

type toolPayload struct {
	ID          string           `json:"id"`
	Name        *string          `json:"name,omitempty"`
	Description *string          `json:"description,omitempty"`
	Type        string           `json:"type"`
	Config      *json.RawMessage `json:"config,omitempty"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   *time.Time       `json:"updatedAt,omitempty"`
}

func mapTool(source any) (*Tool, error) {
	var payload toolPayload
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &Tool{
		ID:          payload.ID,
		Name:        payload.Name,
		Description: payload.Description,
		Type:        payload.Type,
		Config:      payload.Config,
		CreatedAt:   payload.CreatedAt,
		UpdatedAt:   payload.UpdatedAt,
	}, nil
}
