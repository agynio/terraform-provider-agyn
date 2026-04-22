package agentapi

import (
	"context"
	"fmt"

	secretsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/secrets/v1"
)

func (c *Client) CreateSecret(ctx context.Context, req *secretsv1.CreateSecretRequest) (*secretsv1.Secret, error) {
	return withConflictRetry(ctx, "create secret", func() (*secretsv1.Secret, error) {
		resp, err := c.secretsGateway.CreateSecret(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create secret: %w", err)
		}
		return resp.Secret, nil
	})
}

func (c *Client) GetSecret(ctx context.Context, id string) (*secretsv1.Secret, error) {
	resp, err := c.secretsGateway.GetSecret(ctx, &secretsv1.GetSecretRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get secret: %w", err)
	}
	return resp.Secret, nil
}

func (c *Client) UpdateSecret(ctx context.Context, req *secretsv1.UpdateSecretRequest) (*secretsv1.Secret, error) {
	resp, err := c.secretsGateway.UpdateSecret(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update secret: %w", err)
	}
	return resp.Secret, nil
}

func (c *Client) DeleteSecret(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete secret", func() error {
		_, err := c.secretsGateway.DeleteSecret(ctx, &secretsv1.DeleteSecretRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete secret: %w", err)
		}
		return nil
	})
}
