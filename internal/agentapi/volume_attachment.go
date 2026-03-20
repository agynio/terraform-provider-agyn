package agentapi

import (
	"context"
	"fmt"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
)

func (c *Client) CreateVolumeAttachment(ctx context.Context, req *agentsv1.CreateVolumeAttachmentRequest) (*agentsv1.VolumeAttachment, error) {
	return withConflictRetry(ctx, "create volume attachment", func() (*agentsv1.VolumeAttachment, error) {
		resp, err := c.gateway.CreateVolumeAttachment(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create volume attachment: %w", err)
		}
		return resp.VolumeAttachment, nil
	})
}

func (c *Client) GetVolumeAttachment(ctx context.Context, id string) (*agentsv1.VolumeAttachment, error) {
	resp, err := c.gateway.GetVolumeAttachment(ctx, &agentsv1.GetVolumeAttachmentRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get volume attachment: %w", err)
	}
	return resp.VolumeAttachment, nil
}

func (c *Client) DeleteVolumeAttachment(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete volume attachment", func() error {
		_, err := c.gateway.DeleteVolumeAttachment(ctx, &agentsv1.DeleteVolumeAttachmentRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete volume attachment: %w", err)
		}
		return nil
	})
}
