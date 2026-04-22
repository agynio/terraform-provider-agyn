package agentapi

import (
	"context"
	"fmt"

	secretsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/secrets/v1"
)

func (c *Client) CreateSecretProvider(ctx context.Context, req *secretsv1.CreateSecretProviderRequest) (*secretsv1.SecretProvider, error) {
	return withConflictRetry(ctx, "create secret provider", func() (*secretsv1.SecretProvider, error) {
		resp, err := c.secretsGateway.CreateSecretProvider(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create secret provider: %w", err)
		}
		return resp.SecretProvider, nil
	})
}

func (c *Client) GetSecretProvider(ctx context.Context, id string) (*secretsv1.SecretProvider, error) {
	resp, err := c.secretsGateway.GetSecretProvider(ctx, &secretsv1.GetSecretProviderRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get secret provider: %w", err)
	}
	return resp.SecretProvider, nil
}

func (c *Client) UpdateSecretProvider(ctx context.Context, req *secretsv1.UpdateSecretProviderRequest) (*secretsv1.SecretProvider, error) {
	resp, err := c.secretsGateway.UpdateSecretProvider(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update secret provider: %w", err)
	}
	return resp.SecretProvider, nil
}

func (c *Client) DeleteSecretProvider(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete secret provider", func() error {
		_, err := c.secretsGateway.DeleteSecretProvider(ctx, &secretsv1.DeleteSecretProviderRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete secret provider: %w", err)
		}
		return nil
	})
}
