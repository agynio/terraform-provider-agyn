package teamapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WorkspaceConfiguration struct {
	ID          string
	Title       *string
	Description *string
	Config      WorkspaceConfigurationConfig
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type WorkspaceConfigurationCreate struct {
	Title       *string
	Description *string
	Config      WorkspaceConfigurationConfig
}

type WorkspaceConfigurationUpdate struct {
	Title       *string
	Description *string
	Config      *WorkspaceConfigurationConfig
}

func (c *Client) CreateWorkspaceConfiguration(ctx context.Context, input WorkspaceConfigurationCreate) (*WorkspaceConfiguration, error) {
	payload := map[string]any{
		"config": input.Config,
	}
	if input.Title != nil {
		payload["title"] = *input.Title
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal workspace configuration payload: %w", err)
	}

	return withConflictRetry(ctx, "create workspace configuration", func() (*WorkspaceConfiguration, error) {
		resp, err := c.raw.PostWorkspaceConfigurationsWithBodyWithResponse(ctx, "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("create workspace configuration request: %w", err)
		}

		if resp.JSON201 == nil {
			return nil, errorFromResponse("create workspace configuration", responseStatus(resp), resp.Body)
		}

		cfg, err := mapWorkspaceConfiguration(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decode workspace configuration response: %w", err)
		}
		return cfg, nil
	})
}

func (c *Client) GetWorkspaceConfiguration(ctx context.Context, id string) (*WorkspaceConfiguration, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	resp, err := c.raw.GetWorkspaceConfigurationsIdWithResponse(ctx, uuidValue)
	if err != nil {
		return nil, fmt.Errorf("get workspace configuration request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("get workspace configuration", responseStatus(resp), resp.Body)
	}

	cfg, err := mapWorkspaceConfiguration(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode workspace configuration response: %w", err)
	}
	return cfg, nil
}

func (c *Client) UpdateWorkspaceConfiguration(ctx context.Context, id string, input WorkspaceConfigurationUpdate) (*WorkspaceConfiguration, error) {
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
		payload["config"] = *input.Config
	}

	if len(payload) == 0 {
		return c.GetWorkspaceConfiguration(ctx, id)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal workspace configuration payload: %w", err)
	}

	resp, err := c.raw.PatchWorkspaceConfigurationsIdWithBodyWithResponse(ctx, uuidValue, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("update workspace configuration request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("update workspace configuration", responseStatus(resp), resp.Body)
	}

	cfg, err := mapWorkspaceConfiguration(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode workspace configuration response: %w", err)
	}
	return cfg, nil
}

func (c *Client) DeleteWorkspaceConfiguration(ctx context.Context, id string) error {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return err
	}

	return withConflictRetryNoResult(ctx, "delete workspace configuration", func() error {
		resp, err := c.raw.DeleteWorkspaceConfigurationsIdWithResponse(ctx, uuidValue)
		if err != nil {
			return fmt.Errorf("delete workspace configuration request: %w", err)
		}

		if resp.StatusCode() == http.StatusNoContent {
			return nil
		}

		return errorFromResponse("delete workspace configuration", responseStatus(resp), resp.Body)
	})
}

type workspaceConfigurationPayload struct {
	ID          string          `json:"id"`
	Title       *string         `json:"title,omitempty"`
	Description *string         `json:"description,omitempty"`
	Config      WorkspaceConfigurationConfig `json:"config"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   *time.Time      `json:"updatedAt,omitempty"`
}

func mapWorkspaceConfiguration(source any) (*WorkspaceConfiguration, error) {
	var payload workspaceConfigurationPayload
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &WorkspaceConfiguration{
		ID:          payload.ID,
		Title:       payload.Title,
		Description: payload.Description,
		Config:      payload.Config,
		CreatedAt:   payload.CreatedAt,
		UpdatedAt:   payload.UpdatedAt,
	}, nil
}
