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
	ID            string
	Name          string
	Role          string
	Model         string
	Image         string
	Description   *string
	Configuration *string
	Resources     *ComputeResources
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

type AgentCreate struct {
	Name          string
	Role          string
	Model         string
	Image         string
	Description   *string
	Configuration *string
	Resources     *ComputeResources
}

type AgentUpdate struct {
	Name          *string
	Role          *string
	Model         *string
	Image         *string
	Description   *string
	Configuration *string
	Resources     *ComputeResources
}

func (c *Client) CreateAgent(ctx context.Context, input AgentCreate) (*Agent, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if input.Role == "" {
		return nil, fmt.Errorf("role is required")
	}
	if input.Model == "" {
		return nil, fmt.Errorf("model is required")
	}
	if input.Image == "" {
		return nil, fmt.Errorf("image is required")
	}
	if _, err := parseUUID(input.Model); err != nil {
		return nil, fmt.Errorf("invalid model: %w", err)
	}

	payload := map[string]any{
		"name":  input.Name,
		"role":  input.Role,
		"model": input.Model,
		"image": input.Image,
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}
	if input.Configuration != nil {
		payload["configuration"] = *input.Configuration
	}
	if input.Resources != nil {
		payload["resources"] = *input.Resources
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

		agent, err := mapAgent(resp.Body)
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

	agent, err := mapAgent(resp.Body)
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
	if input.Name != nil {
		payload["name"] = *input.Name
	}
	if input.Role != nil {
		payload["role"] = *input.Role
	}
	if input.Model != nil {
		if *input.Model == "" {
			return nil, fmt.Errorf("model cannot be empty")
		}
		if _, err := parseUUID(*input.Model); err != nil {
			return nil, fmt.Errorf("invalid model: %w", err)
		}
		payload["model"] = *input.Model
	}
	if input.Image != nil {
		payload["image"] = *input.Image
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}
	if input.Configuration != nil {
		payload["configuration"] = *input.Configuration
	}
	if input.Resources != nil {
		payload["resources"] = *input.Resources
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

	agent, err := mapAgent(resp.Body)
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
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Role          string            `json:"role"`
	Model         string            `json:"model"`
	Image         string            `json:"image"`
	Description   *string           `json:"description,omitempty"`
	Configuration *string           `json:"configuration,omitempty"`
	Resources     *ComputeResources `json:"resources,omitempty"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     *time.Time        `json:"updatedAt,omitempty"`
}

func mapAgent(source any) (*Agent, error) {
	var payload agentPayload
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &Agent{
		ID:            payload.ID,
		Name:          payload.Name,
		Role:          payload.Role,
		Model:         payload.Model,
		Image:         payload.Image,
		Description:   payload.Description,
		Configuration: payload.Configuration,
		Resources:     payload.Resources,
		CreatedAt:     payload.CreatedAt,
		UpdatedAt:     payload.UpdatedAt,
	}, nil
}
