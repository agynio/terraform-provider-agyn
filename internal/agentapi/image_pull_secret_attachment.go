package agentapi

import (
	"context"
	"fmt"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
)

func (c *Client) CreateImagePullSecretAttachment(ctx context.Context, req *agentsv1.CreateImagePullSecretAttachmentRequest) (*agentsv1.ImagePullSecretAttachment, error) {
	return withConflictRetry(ctx, "create image pull secret attachment", func() (*agentsv1.ImagePullSecretAttachment, error) {
		resp, err := c.gateway.CreateImagePullSecretAttachment(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create image pull secret attachment: %w", err)
		}
		return resp.ImagePullSecretAttachment, nil
	})
}

func (c *Client) GetImagePullSecretAttachment(ctx context.Context, id string) (*agentsv1.ImagePullSecretAttachment, error) {
	resp, err := c.gateway.GetImagePullSecretAttachment(ctx, &agentsv1.GetImagePullSecretAttachmentRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get image pull secret attachment: %w", err)
	}
	return resp.ImagePullSecretAttachment, nil
}

func (c *Client) DeleteImagePullSecretAttachment(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete image pull secret attachment", func() error {
		_, err := c.gateway.DeleteImagePullSecretAttachment(ctx, &agentsv1.DeleteImagePullSecretAttachmentRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete image pull secret attachment: %w", err)
		}
		return nil
	})
}
