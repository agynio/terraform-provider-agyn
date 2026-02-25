package teamapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func (c *Client) GetAttachment(ctx context.Context, id string) (*Attachment, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	baseClient, ok := c.raw.ClientInterface.(*teamclient.Client)
	if !ok {
		return nil, fmt.Errorf("unsupported client type %T", c.raw.ClientInterface)
	}

	serverURL, err := url.Parse(baseClient.Server)
	if err != nil {
		return nil, fmt.Errorf("parse server url: %w", err)
	}

	operationPath := fmt.Sprintf("/attachments/%s", uuidToString(uuidValue))
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	requestURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, fmt.Errorf("build attachment url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create attachment request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := baseClient.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get attachment request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read attachment response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var attachment Attachment
		if err := json.Unmarshal(body, &attachment); err != nil {
			return nil, fmt.Errorf("decode attachment response: %w", err)
		}
		return &attachment, nil
	case http.StatusNotFound:
		return nil, ErrAttachmentNotFound
	default:
		return nil, errorFromResponse("get attachment", resp.StatusCode, body)
	}
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
