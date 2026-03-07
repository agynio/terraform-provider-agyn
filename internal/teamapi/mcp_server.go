package teamapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MCPServer struct {
	ID          string
	Title       *string
	Description *string
	Config      json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type MCPServerCreate struct {
	Title       *string
	Description *string
	Config      json.RawMessage
}

type MCPServerUpdate struct {
	Title       *string
	Description *string
	Config      *json.RawMessage
}

func (c *Client) CreateMCPServer(ctx context.Context, input MCPServerCreate) (*MCPServer, error) {
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
		return nil, fmt.Errorf("marshal MCP server payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < graphConflictRetryCount; attempt++ {
		if err := waitForConflictRetry(ctx, attempt); err != nil {
			return nil, err
		}

		resp, err := c.raw.PostMcpServersWithBodyWithResponse(ctx, "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("create MCP server request: %w", err)
		}

		if resp.JSON201 == nil {
			err := errorFromResponse("create MCP server", responseStatus(resp), resp.Body)
			if isVersionConflict(err) {
				lastErr = err
				continue
			}
			return nil, err
		}

		server, err := mapMCPServer(resp.JSON201)
		if err != nil {
			return nil, fmt.Errorf("decode MCP server response: %w", err)
		}
		return server, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("create MCP server failed after %d attempts", graphConflictRetryCount)
}

func (c *Client) GetMCPServer(ctx context.Context, id string) (*MCPServer, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	resp, err := c.raw.GetMcpServersIdWithResponse(ctx, uuidValue)
	if err != nil {
		return nil, fmt.Errorf("get MCP server request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("get MCP server", responseStatus(resp), resp.Body)
	}

	server, err := mapMCPServer(resp.JSON200)
	if err != nil {
		return nil, fmt.Errorf("decode MCP server response: %w", err)
	}
	return server, nil
}

func (c *Client) UpdateMCPServer(ctx context.Context, id string, input MCPServerUpdate) (*MCPServer, error) {
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
		return c.GetMCPServer(ctx, id)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal MCP server payload: %w", err)
	}

	resp, err := c.raw.PatchMcpServersIdWithBodyWithResponse(ctx, uuidValue, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("update MCP server request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("update MCP server", responseStatus(resp), resp.Body)
	}

	server, err := mapMCPServer(resp.JSON200)
	if err != nil {
		return nil, fmt.Errorf("decode MCP server response: %w", err)
	}
	return server, nil
}

func (c *Client) DeleteMCPServer(ctx context.Context, id string) error {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt < graphConflictRetryCount; attempt++ {
		if err := waitForConflictRetry(ctx, attempt); err != nil {
			return err
		}

		resp, err := c.raw.DeleteMcpServersIdWithResponse(ctx, uuidValue)
		if err != nil {
			return fmt.Errorf("delete MCP server request: %w", err)
		}

		if resp.StatusCode() == http.StatusNoContent {
			return nil
		}

		err = errorFromResponse("delete MCP server", responseStatus(resp), resp.Body)
		if isVersionConflict(err) {
			lastErr = err
			continue
		}
		return err
	}

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("delete MCP server failed after %d attempts", graphConflictRetryCount)
}

type mcpServerPayload struct {
	ID          string          `json:"id"`
	Title       *string         `json:"title,omitempty"`
	Description *string         `json:"description,omitempty"`
	Config      json.RawMessage `json:"config"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   *time.Time      `json:"updatedAt,omitempty"`
}

func mapMCPServer(source any) (*MCPServer, error) {
	var payload mcpServerPayload
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &MCPServer{
		ID:          payload.ID,
		Title:       payload.Title,
		Description: payload.Description,
		Config:      payload.Config,
		CreatedAt:   payload.CreatedAt,
		UpdatedAt:   payload.UpdatedAt,
	}, nil
}
