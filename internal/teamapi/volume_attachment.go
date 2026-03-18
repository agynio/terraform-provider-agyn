package teamapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type VolumeAttachment struct {
	ID        string
	VolumeID  string
	AgentID   *string
	McpID     *string
	HookID    *string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type VolumeAttachmentCreate struct {
	VolumeID string
	AgentID  *string
	McpID    *string
	HookID   *string
}

func (c *Client) CreateVolumeAttachment(ctx context.Context, input VolumeAttachmentCreate) (*VolumeAttachment, error) {
	if input.VolumeID == "" {
		return nil, fmt.Errorf("volume_id is required")
	}
	if _, err := parseUUID(input.VolumeID); err != nil {
		return nil, fmt.Errorf("invalid volume_id: %w", err)
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
		"volumeId": input.VolumeID,
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
		return nil, fmt.Errorf("marshal volume attachment payload: %w", err)
	}

	return withConflictRetry(ctx, "create volume attachment", func() (*VolumeAttachment, error) {
		resp, err := c.raw.PostVolumeAttachmentsWithBodyWithResponse(ctx, "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("create volume attachment request: %w", err)
		}

		if resp.JSON201 == nil {
			return nil, errorFromResponse("create volume attachment", responseStatus(resp), resp.Body)
		}

		attachment, err := mapVolumeAttachment(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decode volume attachment response: %w", err)
		}
		return attachment, nil
	})
}

func (c *Client) GetVolumeAttachment(ctx context.Context, id string) (*VolumeAttachment, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	resp, err := c.raw.GetVolumeAttachmentsIdWithResponse(ctx, uuidValue)
	if err != nil {
		return nil, fmt.Errorf("get volume attachment request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("get volume attachment", responseStatus(resp), resp.Body)
	}

	attachment, err := mapVolumeAttachment(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode volume attachment response: %w", err)
	}
	return attachment, nil
}

func (c *Client) DeleteVolumeAttachment(ctx context.Context, id string) error {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return err
	}

	return withConflictRetryNoResult(ctx, "delete volume attachment", func() error {
		resp, err := c.raw.DeleteVolumeAttachmentsIdWithResponse(ctx, uuidValue)
		if err != nil {
			return fmt.Errorf("delete volume attachment request: %w", err)
		}

		if resp.StatusCode() == http.StatusNoContent {
			return nil
		}

		return errorFromResponse("delete volume attachment", responseStatus(resp), resp.Body)
	})
}

type volumeAttachmentPayload struct {
	ID        string     `json:"id"`
	VolumeID  string     `json:"volumeId"`
	AgentID   *string    `json:"agentId,omitempty"`
	McpID     *string    `json:"mcpId,omitempty"`
	HookID    *string    `json:"hookId,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

func mapVolumeAttachment(source any) (*VolumeAttachment, error) {
	var payload volumeAttachmentPayload
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &VolumeAttachment{
		ID:        payload.ID,
		VolumeID:  payload.VolumeID,
		AgentID:   payload.AgentID,
		McpID:     payload.McpID,
		HookID:    payload.HookID,
		CreatedAt: payload.CreatedAt,
		UpdatedAt: payload.UpdatedAt,
	}, nil
}
