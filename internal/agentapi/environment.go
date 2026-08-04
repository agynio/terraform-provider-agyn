package agentapi

import (
	"context"
	"fmt"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
)

func (c *Client) CreateEnvironment(ctx context.Context, req *agentsv1.CreateEnvironmentRequest) (*agentsv1.Environment, error) {
	return withConflictRetry(ctx, "create environment", func() (*agentsv1.Environment, error) {
		resp, err := c.gateway.CreateEnvironment(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create environment: %w", err)
		}
		return resp.Environment, nil
	})
}

func (c *Client) GetEnvironment(ctx context.Context, id string) (*agentsv1.Environment, error) {
	resp, err := c.gateway.GetEnvironment(ctx, &agentsv1.GetEnvironmentRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get environment: %w", err)
	}
	return resp.Environment, nil
}

func (c *Client) UpdateEnvironment(ctx context.Context, req *agentsv1.UpdateEnvironmentRequest) (*agentsv1.Environment, error) {
	resp, err := c.gateway.UpdateEnvironment(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update environment: %w", err)
	}
	return resp.Environment, nil
}

func (c *Client) DeleteEnvironment(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete environment", func() error {
		if _, err := c.gateway.DeleteEnvironment(ctx, &agentsv1.DeleteEnvironmentRequest{Id: id}); err != nil {
			return fmt.Errorf("delete environment: %w", err)
		}
		return nil
	})
}
