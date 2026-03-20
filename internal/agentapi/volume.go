package agentapi

import (
	"context"
	"fmt"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
)

func (c *Client) CreateVolume(ctx context.Context, req *agentsv1.CreateVolumeRequest) (*agentsv1.Volume, error) {
	return withConflictRetry(ctx, "create volume", func() (*agentsv1.Volume, error) {
		resp, err := c.gateway.CreateVolume(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create volume: %w", err)
		}
		return resp.Volume, nil
	})
}

func (c *Client) GetVolume(ctx context.Context, id string) (*agentsv1.Volume, error) {
	resp, err := c.gateway.GetVolume(ctx, &agentsv1.GetVolumeRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get volume: %w", err)
	}
	return resp.Volume, nil
}

func (c *Client) UpdateVolume(ctx context.Context, req *agentsv1.UpdateVolumeRequest) (*agentsv1.Volume, error) {
	resp, err := c.gateway.UpdateVolume(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update volume: %w", err)
	}
	return resp.Volume, nil
}

func (c *Client) DeleteVolume(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete volume", func() error {
		_, err := c.gateway.DeleteVolume(ctx, &agentsv1.DeleteVolumeRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete volume: %w", err)
		}
		return nil
	})
}
