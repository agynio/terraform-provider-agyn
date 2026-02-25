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

func (c *Client) GetAttachment(ctx context.Context, id string, kind string, sourceID string, targetID string) (*Attachment, error) {
	if _, err := parseUUID(id); err != nil {
		return nil, err
	}
	srcUUID, err := parseUUID(sourceID)
	if err != nil {
		return nil, fmt.Errorf("invalid source_id: %w", err)
	}
	tgtUUID, err := parseUUID(targetID)
	if err != nil {
		return nil, fmt.Errorf("invalid target_id: %w", err)
	}

	kindValue := teamclient.GetAttachmentsParamsKind(kind)
	page := 1
	perPage := 100
	params := &teamclient.GetAttachmentsParams{
		Kind:     &kindValue,
		SourceId: &srcUUID,
		TargetId: &tgtUUID,
		Page:     &page,
		PerPage:  &perPage,
	}

	for {
		resp, err := c.raw.GetAttachmentsWithResponse(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("list attachments request: %w", err)
		}

		if resp.JSON200 == nil {
			return nil, errorFromResponse("list attachments", responseStatus(resp), resp.Body)
		}

		list, err := mapAttachmentList(resp.JSON200)
		if err != nil {
			return nil, fmt.Errorf("decode attachments response: %w", err)
		}

		for _, item := range list.Items {
			if item.ID == id {
				return &item, nil
			}
		}

		if list.NextPage == nil {
			break
		}
		if params.Page == nil {
			break
		}
		if *params.Page >= list.TotalPages {
			break
		}
		next := *params.Page + 1
		params.Page = &next
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

type attachmentListPayload struct {
	Items      []Attachment `json:"items"`
	Page       int          `json:"page"`
	TotalPages int          `json:"totalPages"`
	NextPage   *int         `json:"nextPage,omitempty"`
}

func mapAttachment(source any) (*Attachment, error) {
	var payload Attachment
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

func mapAttachmentList(source any) (*attachmentListPayload, error) {
	var payload attachmentListPayload
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
