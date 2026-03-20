package agentapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Mcp struct {
	ID          string
	AgentID     string
	Image       string
	Command     string
	Description *string
	Resources   *ComputeResources
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type McpCreate struct {
	AgentID     string
	Image       string
	Command     string
	Description *string
	Resources   *ComputeResources
}

type McpUpdate struct {
	Image       *string
	Command     *string
	Description *string
	Resources   *ComputeResources
}

func (c *Client) CreateMcp(ctx context.Context, input McpCreate) (*Mcp, error) {
	if input.AgentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}
	if input.Image == "" {
		return nil, fmt.Errorf("image is required")
	}
	if input.Command == "" {
		return nil, fmt.Errorf("command is required")
	}
	if _, err := parseUUID(input.AgentID); err != nil {
		return nil, fmt.Errorf("invalid agent_id: %w", err)
	}

	payload := map[string]any{
		"agentId": input.AgentID,
		"image":   input.Image,
		"command": input.Command,
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}
	if input.Resources != nil {
		payload["resources"] = *input.Resources
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal mcp payload: %w", err)
	}

	return withConflictRetry(ctx, "create mcp", func() (*Mcp, error) {
		resp, err := c.raw.PostMcpsWithBodyWithResponse(ctx, "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("create mcp request: %w", err)
		}

		if resp.JSON201 == nil {
			return nil, errorFromResponse("create mcp", responseStatus(resp), resp.Body)
		}

		mcp, err := mapMcp(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decode mcp response: %w", err)
		}
		return mcp, nil
	})
}

func (c *Client) GetMcp(ctx context.Context, id string) (*Mcp, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	resp, err := c.raw.GetMcpsIdWithResponse(ctx, uuidValue)
	if err != nil {
		return nil, fmt.Errorf("get mcp request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("get mcp", responseStatus(resp), resp.Body)
	}

	mcp, err := mapMcp(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode mcp response: %w", err)
	}
	return mcp, nil
}

func (c *Client) UpdateMcp(ctx context.Context, id string, input McpUpdate) (*Mcp, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	payload := make(map[string]any)
	if input.Image != nil {
		payload["image"] = *input.Image
	}
	if input.Command != nil {
		payload["command"] = *input.Command
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}
	if input.Resources != nil {
		payload["resources"] = *input.Resources
	}

	if len(payload) == 0 {
		return c.GetMcp(ctx, id)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal mcp payload: %w", err)
	}

	resp, err := c.raw.PatchMcpsIdWithBodyWithResponse(ctx, uuidValue, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("update mcp request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("update mcp", responseStatus(resp), resp.Body)
	}

	mcp, err := mapMcp(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode mcp response: %w", err)
	}
	return mcp, nil
}

func (c *Client) DeleteMcp(ctx context.Context, id string) error {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return err
	}

	return withConflictRetryNoResult(ctx, "delete mcp", func() error {
		resp, err := c.raw.DeleteMcpsIdWithResponse(ctx, uuidValue)
		if err != nil {
			return fmt.Errorf("delete mcp request: %w", err)
		}

		if resp.StatusCode() == http.StatusNoContent {
			return nil
		}

		return errorFromResponse("delete mcp", responseStatus(resp), resp.Body)
	})
}

type mcpPayload struct {
	ID          string            `json:"id"`
	AgentID     string            `json:"agentId"`
	Image       string            `json:"image"`
	Command     string            `json:"command"`
	Description *string           `json:"description,omitempty"`
	Resources   *ComputeResources `json:"resources,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   *time.Time        `json:"updatedAt,omitempty"`
}

func mapMcp(source any) (*Mcp, error) {
	var payload mcpPayload
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &Mcp{
		ID:          payload.ID,
		AgentID:     payload.AgentID,
		Image:       payload.Image,
		Command:     payload.Command,
		Description: payload.Description,
		Resources:   payload.Resources,
		CreatedAt:   payload.CreatedAt,
		UpdatedAt:   payload.UpdatedAt,
	}, nil
}
