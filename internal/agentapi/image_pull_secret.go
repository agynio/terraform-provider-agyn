package agentapi

import (
	"context"
	"fmt"

	secretsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/secrets/v1"
)

func (c *Client) CreateImagePullSecret(ctx context.Context, req *secretsv1.CreateImagePullSecretRequest) (*secretsv1.ImagePullSecret, error) {
	return withConflictRetry(ctx, "create image pull secret", func() (*secretsv1.ImagePullSecret, error) {
		resp, err := c.secretsGateway.CreateImagePullSecret(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create image pull secret: %w", err)
		}
		return resp.ImagePullSecret, nil
	})
}

func (c *Client) GetImagePullSecret(ctx context.Context, id string) (*secretsv1.ImagePullSecret, error) {
	resp, err := c.secretsGateway.GetImagePullSecret(ctx, &secretsv1.GetImagePullSecretRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get image pull secret: %w", err)
	}
	return resp.ImagePullSecret, nil
}

func (c *Client) UpdateImagePullSecret(ctx context.Context, req *secretsv1.UpdateImagePullSecretRequest) (*secretsv1.ImagePullSecret, error) {
	resp, err := c.secretsGateway.UpdateImagePullSecret(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update image pull secret: %w", err)
	}
	return resp.ImagePullSecret, nil
}

func (c *Client) DeleteImagePullSecret(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete image pull secret", func() error {
		_, err := c.secretsGateway.DeleteImagePullSecret(ctx, &secretsv1.DeleteImagePullSecretRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete image pull secret: %w", err)
		}
		return nil
	})
}
