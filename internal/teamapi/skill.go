package teamapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Skill struct {
	ID          string
	AgentID     string
	Name        string
	Body        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type SkillCreate struct {
	AgentID     string
	Name        string
	Body        string
	Description *string
}

type SkillUpdate struct {
	Name        *string
	Body        *string
	Description *string
}

func (c *Client) CreateSkill(ctx context.Context, input SkillCreate) (*Skill, error) {
	if input.AgentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if input.Body == "" {
		return nil, fmt.Errorf("body is required")
	}
	if _, err := parseUUID(input.AgentID); err != nil {
		return nil, fmt.Errorf("invalid agent_id: %w", err)
	}

	payload := map[string]any{
		"agentId": input.AgentID,
		"name":    input.Name,
		"body":    input.Body,
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal skill payload: %w", err)
	}

	return withConflictRetry(ctx, "create skill", func() (*Skill, error) {
		resp, err := c.raw.PostSkillsWithBodyWithResponse(ctx, "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("create skill request: %w", err)
		}

		if resp.JSON201 == nil {
			return nil, errorFromResponse("create skill", responseStatus(resp), resp.Body)
		}

		skill, err := mapSkill(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decode skill response: %w", err)
		}
		return skill, nil
	})
}

func (c *Client) GetSkill(ctx context.Context, id string) (*Skill, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	resp, err := c.raw.GetSkillsIdWithResponse(ctx, uuidValue)
	if err != nil {
		return nil, fmt.Errorf("get skill request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("get skill", responseStatus(resp), resp.Body)
	}

	skill, err := mapSkill(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode skill response: %w", err)
	}
	return skill, nil
}

func (c *Client) UpdateSkill(ctx context.Context, id string, input SkillUpdate) (*Skill, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	payload := make(map[string]any)
	if input.Name != nil {
		payload["name"] = *input.Name
	}
	if input.Body != nil {
		payload["body"] = *input.Body
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}

	if len(payload) == 0 {
		return c.GetSkill(ctx, id)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal skill payload: %w", err)
	}

	resp, err := c.raw.PatchSkillsIdWithBodyWithResponse(ctx, uuidValue, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("update skill request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("update skill", responseStatus(resp), resp.Body)
	}

	skill, err := mapSkill(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode skill response: %w", err)
	}
	return skill, nil
}

func (c *Client) DeleteSkill(ctx context.Context, id string) error {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return err
	}

	return withConflictRetryNoResult(ctx, "delete skill", func() error {
		resp, err := c.raw.DeleteSkillsIdWithResponse(ctx, uuidValue)
		if err != nil {
			return fmt.Errorf("delete skill request: %w", err)
		}

		if resp.StatusCode() == http.StatusNoContent {
			return nil
		}

		return errorFromResponse("delete skill", responseStatus(resp), resp.Body)
	})
}

type skillPayload struct {
	ID          string     `json:"id"`
	AgentID     string     `json:"agentId"`
	Name        string     `json:"name"`
	Body        string     `json:"body"`
	Description *string    `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

func mapSkill(source any) (*Skill, error) {
	var payload skillPayload
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &Skill{
		ID:          payload.ID,
		AgentID:     payload.AgentID,
		Name:        payload.Name,
		Body:        payload.Body,
		Description: payload.Description,
		CreatedAt:   payload.CreatedAt,
		UpdatedAt:   payload.UpdatedAt,
	}, nil
}
