package teamapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MemoryBucket struct {
	ID          string
	Title       *string
	Description *string
	Config      json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type MemoryBucketCreate struct {
	Title       *string
	Description *string
	Config      json.RawMessage
}

type MemoryBucketUpdate struct {
	Title       *string
	Description *string
	Config      *json.RawMessage
}

func (c *Client) CreateMemoryBucket(ctx context.Context, input MemoryBucketCreate) (*MemoryBucket, error) {
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
		return nil, fmt.Errorf("marshal memory bucket payload: %w", err)
	}

	resp, err := c.raw.PostMemoryBucketsWithBodyWithResponse(ctx, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create memory bucket request: %w", err)
	}

	if resp.JSON201 == nil {
		return nil, errorFromResponse("create memory bucket", responseStatus(resp), resp.Body)
	}

	bucket, err := mapMemoryBucket(resp.JSON201)
	if err != nil {
		return nil, fmt.Errorf("decode memory bucket response: %w", err)
	}
	return bucket, nil
}

func (c *Client) GetMemoryBucket(ctx context.Context, id string) (*MemoryBucket, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	resp, err := c.raw.GetMemoryBucketsIdWithResponse(ctx, uuidValue)
	if err != nil {
		return nil, fmt.Errorf("get memory bucket request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("get memory bucket", responseStatus(resp), resp.Body)
	}

	bucket, err := mapMemoryBucket(resp.JSON200)
	if err != nil {
		return nil, fmt.Errorf("decode memory bucket response: %w", err)
	}
	return bucket, nil
}

func (c *Client) UpdateMemoryBucket(ctx context.Context, id string, input MemoryBucketUpdate) (*MemoryBucket, error) {
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
		return c.GetMemoryBucket(ctx, id)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal memory bucket payload: %w", err)
	}

	resp, err := c.raw.PatchMemoryBucketsIdWithBodyWithResponse(ctx, uuidValue, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("update memory bucket request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("update memory bucket", responseStatus(resp), resp.Body)
	}

	bucket, err := mapMemoryBucket(resp.JSON200)
	if err != nil {
		return nil, fmt.Errorf("decode memory bucket response: %w", err)
	}
	return bucket, nil
}

func (c *Client) DeleteMemoryBucket(ctx context.Context, id string) error {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return err
	}

	resp, err := c.raw.DeleteMemoryBucketsIdWithResponse(ctx, uuidValue)
	if err != nil {
		return fmt.Errorf("delete memory bucket request: %w", err)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return errorFromResponse("delete memory bucket", responseStatus(resp), resp.Body)
	}
	return nil
}

type memoryBucketPayload struct {
	ID          string          `json:"id"`
	Title       *string         `json:"title,omitempty"`
	Description *string         `json:"description,omitempty"`
	Config      json.RawMessage `json:"config"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   *time.Time      `json:"updatedAt,omitempty"`
}

func mapMemoryBucket(source any) (*MemoryBucket, error) {
	var payload memoryBucketPayload
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &MemoryBucket{
		ID:          payload.ID,
		Title:       payload.Title,
		Description: payload.Description,
		Config:      payload.Config,
		CreatedAt:   payload.CreatedAt,
		UpdatedAt:   payload.UpdatedAt,
	}, nil
}
