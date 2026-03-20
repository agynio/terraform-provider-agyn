package agentapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Volume struct {
	ID          string
	Persistent  bool
	MountPath   string
	Size        *string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type VolumeCreate struct {
	Persistent  bool
	MountPath   string
	Size        *string
	Description *string
}

type VolumeUpdate struct {
	Persistent  *bool
	MountPath   *string
	Size        *string
	Description *string
}

func (c *Client) CreateVolume(ctx context.Context, input VolumeCreate) (*Volume, error) {
	if input.MountPath == "" {
		return nil, fmt.Errorf("mount_path is required")
	}
	if input.Persistent && (input.Size == nil || *input.Size == "") {
		return nil, fmt.Errorf("size is required when persistent is true")
	}

	payload := map[string]any{
		"persistent": input.Persistent,
		"mountPath":  input.MountPath,
	}
	if input.Size != nil {
		payload["size"] = *input.Size
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal volume payload: %w", err)
	}

	return withConflictRetry(ctx, "create volume", func() (*Volume, error) {
		resp, err := c.raw.PostVolumesWithBodyWithResponse(ctx, "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("create volume request: %w", err)
		}

		if resp.JSON201 == nil {
			return nil, errorFromResponse("create volume", responseStatus(resp), resp.Body)
		}

		volume, err := mapVolume(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decode volume response: %w", err)
		}
		return volume, nil
	})
}

func (c *Client) GetVolume(ctx context.Context, id string) (*Volume, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	resp, err := c.raw.GetVolumesIdWithResponse(ctx, uuidValue)
	if err != nil {
		return nil, fmt.Errorf("get volume request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("get volume", responseStatus(resp), resp.Body)
	}

	volume, err := mapVolume(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode volume response: %w", err)
	}
	return volume, nil
}

func (c *Client) UpdateVolume(ctx context.Context, id string, input VolumeUpdate) (*Volume, error) {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	payload := make(map[string]any)
	if input.Persistent != nil {
		payload["persistent"] = *input.Persistent
	}
	if input.MountPath != nil {
		if *input.MountPath == "" {
			return nil, fmt.Errorf("mount_path cannot be empty")
		}
		payload["mountPath"] = *input.MountPath
	}
	if input.Size != nil {
		payload["size"] = *input.Size
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}

	if len(payload) == 0 {
		return c.GetVolume(ctx, id)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal volume payload: %w", err)
	}

	resp, err := c.raw.PatchVolumesIdWithBodyWithResponse(ctx, uuidValue, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("update volume request: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, errorFromResponse("update volume", responseStatus(resp), resp.Body)
	}

	volume, err := mapVolume(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode volume response: %w", err)
	}
	return volume, nil
}

func (c *Client) DeleteVolume(ctx context.Context, id string) error {
	uuidValue, err := parseUUID(id)
	if err != nil {
		return err
	}

	return withConflictRetryNoResult(ctx, "delete volume", func() error {
		resp, err := c.raw.DeleteVolumesIdWithResponse(ctx, uuidValue)
		if err != nil {
			return fmt.Errorf("delete volume request: %w", err)
		}

		if resp.StatusCode() == http.StatusNoContent {
			return nil
		}

		return errorFromResponse("delete volume", responseStatus(resp), resp.Body)
	})
}

type volumePayload struct {
	ID          string     `json:"id"`
	Persistent  bool       `json:"persistent"`
	MountPath   string     `json:"mountPath"`
	Size        *string    `json:"size,omitempty"`
	Description *string    `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

func mapVolume(source any) (*Volume, error) {
	var payload volumePayload
	if err := decodePayload(source, &payload); err != nil {
		return nil, err
	}
	return &Volume{
		ID:          payload.ID,
		Persistent:  payload.Persistent,
		MountPath:   payload.MountPath,
		Size:        payload.Size,
		Description: payload.Description,
		CreatedAt:   payload.CreatedAt,
		UpdatedAt:   payload.UpdatedAt,
	}, nil
}
