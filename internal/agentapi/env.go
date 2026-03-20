package agentapi

import (
	"context"
	"fmt"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
)

func (c *Client) CreateEnv(ctx context.Context, req *agentsv1.CreateEnvRequest) (*agentsv1.Env, error) {
	return withConflictRetry(ctx, "create env", func() (*agentsv1.Env, error) {
		resp, err := c.gateway.CreateEnv(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create env: %w", err)
		}
		return resp.Env, nil
	})
}

func (c *Client) GetEnv(ctx context.Context, id string) (*agentsv1.Env, error) {
	resp, err := c.gateway.GetEnv(ctx, &agentsv1.GetEnvRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get env: %w", err)
	}
	return resp.Env, nil
}

func (c *Client) UpdateEnv(ctx context.Context, req *agentsv1.UpdateEnvRequest) (*agentsv1.Env, error) {
	resp, err := c.gateway.UpdateEnv(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update env: %w", err)
	}
	return resp.Env, nil
}

func (c *Client) DeleteEnv(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete env", func() error {
		_, err := c.gateway.DeleteEnv(ctx, &agentsv1.DeleteEnvRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete env: %w", err)
		}
		return nil
	})
}
