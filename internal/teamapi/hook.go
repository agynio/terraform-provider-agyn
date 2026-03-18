package teamapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Hook struct {
	ID          string
	AgentID     string
	Event       string
	Function    string
	Image       string
	Description *string
	Resources   *ComputeResources
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type HookCreate struct {
	AgentID     string
	Event       string
	Function    string
	Image       string
	Description *string
	Resources   *ComputeResources
}

type HookUpdate struct {
	Event       *string
	Function    *string
	Image       *string
	Description *string
	Resources   *ComputeResources
}

func (c *Client) CreateHook(ctx context.Context, input HookCreate) (*Hook, error) {
	if input.AgentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}
	if input.Event == "" {
		return nil, fmt.Errorf("event is required")
	}
	if input.Function == "" {
		return nil, fmt.Errorf("function is required")
	}
	if input.Image == "" {
		return nil, fmt.Errorf("image is required")
	}
	if _, err := parseUUID(input.AgentID); err != nil {
		return nil, fmt.Errorf("invalid agent_id: %w", err)
	}

	payload := map[string]any{
		"agentId":  input.AgentID,
		"event":    input.Event,
		"function": input.Function,
		"image":    input.Image,
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}
	if input.Resources != nil {
		payload["resources"] = *input.Resources
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal hook payload: %w", err)
	}

	return withConflictRetry(ctx, "create hook", func() (*Hook, error) {
		resp, err := c.raw.PostHooksWithBodyWithResponse(ctx, "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("create hook request: %w", err)
		}

		if resp.JSON201 == nil {
			return nil, errorFromResponse("create hook", responseStatus(resp), resp.Body)
		}

		hook, err := mapHook(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decode hook response: %w", err)
		}
		return hook, nil
	})
}

func (c *Client) GetHook(ctx context.Context, id string) (*Hook, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	resp, err := c.raw.GetHooksIdWithResponse(ctx, uuidValue)
	if err != nil {
		return nil, fmt.Errorf("get hook request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("get hook", responseStatus(resp), resp.Body)
	}

	hook, err := mapHook(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode hook response: %w", err)
	}
	return hook, nil
}

func (c *Client) UpdateHook(ctx context.Context, id string, input HookUpdate) (*Hook, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	payload := make(map[string]any)
	if input.Event != nil {
		payload["event"] = *input.Event
	}
	if input.Function != nil {
		payload["function"] = *input.Function
	}
	if input.Image != nil {
		payload["image"] = *input.Image
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}
	if input.Resources != nil {
		payload["resources"] = *input.Resources
	}

	if len(payload) == 0 {
		return c.GetHook(ctx, id)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal hook payload: %w", err)
	}

	resp, err := c.raw.PatchHooksIdWithBodyWithResponse(ctx, uuidValue, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("update hook request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("update hook", responseStatus(resp), resp.Body)
	}

	hook, err := mapHook(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode hook response: %w", err)
	}
	return hook, nil
}

func (c *Client) DeleteHook(ctx context.Context, id string) error {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return err
	}

	return withConflictRetryNoResult(ctx, "delete hook", func() error {
		resp, err := c.raw.DeleteHooksIdWithResponse(ctx, uuidValue)
		if err != nil {
			return fmt.Errorf("delete hook request: %w", err)
		}

		if resp.StatusCode() == http.StatusNoContent {
			return nil
		}

		return errorFromResponse("delete hook", responseStatus(resp), resp.Body)
	})
}

type hookPayload struct {
	ID          string            `json:"id"`
	AgentID     string            `json:"agentId"`
	Event       string            `json:"event"`
	Function    string            `json:"function"`
	Image       string            `json:"image"`
	Description *string           `json:"description,omitempty"`
	Resources   *ComputeResources `json:"resources,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   *time.Time        `json:"updatedAt,omitempty"`
}

func mapHook(source any) (*Hook, error) {
	var payload hookPayload
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &Hook{
		ID:          payload.ID,
		AgentID:     payload.AgentID,
		Event:       payload.Event,
		Function:    payload.Function,
		Image:       payload.Image,
		Description: payload.Description,
		Resources:   payload.Resources,
		CreatedAt:   payload.CreatedAt,
		UpdatedAt:   payload.UpdatedAt,
	}, nil
}
