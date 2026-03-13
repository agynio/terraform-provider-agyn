package teamapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/agynio/terraform-provider-agyn/internal/teamclient"
)

type Attachment struct {
	ID         string     `json:"id"`
	Kind       string     `json:"kind"`
	SourceID   string     `json:"sourceId"`
	SourceType string     `json:"sourceType"`
	TargetID   string     `json:"targetId"`
	TargetType string     `json:"targetType"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
}

type AttachmentCreate struct {
	Kind     string
	SourceID string
	TargetID string
}

var ErrAttachmentNotFound = errors.New("attachment not found")

func (c *Client) CreateAttachment(ctx context.Context, input AttachmentCreate) (*Attachment, error) {
	if input.Kind == "" {
		return nil, fmt.Errorf("kind is required")
	}
	if _, err := parseUUID(input.SourceID); err != nil {
		return nil, fmt.Errorf("invalid source_id: %w", err)
	}
	if _, err := parseUUID(input.TargetID); err != nil {
		return nil, fmt.Errorf("invalid target_id: %w", err)
	}

	payload := map[string]any{
		"kind":     input.Kind,
		"sourceId": input.SourceID,
		"targetId": input.TargetID,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal attachment payload: %w", err)
	}

	resp, err := c.raw.PostAttachmentsWithBodyWithResponse(ctx, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create attachment request: %w", err)
	}

	if resp.JSON201 == nil {
		return nil, errorFromResponse("create attachment", responseStatus(resp), resp.Body)
	}

	attachment, err := mapAttachment(resp.JSON201)
	if err != nil {
		return nil, fmt.Errorf("decode attachment response: %w", err)
	}
	return attachment, nil
}

func (c *Client) FindAttachmentByID(ctx context.Context, id, kind, sourceID, targetID string) (*Attachment, error) {
	if kind == "" {
		return nil, fmt.Errorf("kind is required")
	}

	attachmentID, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	sourceUUID, err := parseUUID(sourceID)
	if err != nil {
		return nil, fmt.Errorf("invalid source_id: %w", err)
	}
	targetUUID, err := parseUUID(targetID)
	if err != nil {
		return nil, fmt.Errorf("invalid target_id: %w", err)
	}

	paramsKind := teamclient.GetAttachmentsParamsKind(kind)
	params := teamclient.GetAttachmentsParams{
		Kind:     &paramsKind,
		SourceId: &sourceUUID,
		TargetId: &targetUUID,
	}

	resp, err := c.raw.GetAttachmentsWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("get attachments request: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, errorFromResponse("get attachments", responseStatus(resp), resp.Body)
	}

	for _, item := range resp.JSON200.Items {
		if item.Id != attachmentID {
			continue
		}

		return &Attachment{
			ID:         uuidToString(item.Id),
			Kind:       string(item.Kind),
			SourceID:   uuidToString(item.SourceId),
			SourceType: string(item.SourceType),
			TargetID:   uuidToString(item.TargetId),
			TargetType: string(item.TargetType),
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		}, nil
	}

	return nil, ErrAttachmentNotFound
}

func (c *Client) DeleteAttachment(ctx context.Context, id string) error {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return err
	}

	resp, err := c.raw.DeleteAttachmentsIdWithResponse(ctx, uuidValue)
	if err != nil {
		return fmt.Errorf("delete attachment request: %w", err)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return errorFromResponse("delete attachment", responseStatus(resp), resp.Body)
	}
	return nil
}

func mapAttachment(source any) (*Attachment, error) {
	var payload Attachment
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
