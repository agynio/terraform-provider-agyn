package agentapi

import (
	"context"
	"fmt"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
)

func (c *Client) CreateHook(ctx context.Context, req *agentsv1.CreateHookRequest) (*agentsv1.Hook, error) {
	return withConflictRetry(ctx, "create hook", func() (*agentsv1.Hook, error) {
		resp, err := c.gateway.CreateHook(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create hook: %w", err)
		}
		return resp.Hook, nil
	})
}

func (c *Client) GetHook(ctx context.Context, id string) (*agentsv1.Hook, error) {
	resp, err := c.gateway.GetHook(ctx, &agentsv1.GetHookRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get hook: %w", err)
	}
	return resp.Hook, nil
}

func (c *Client) UpdateHook(ctx context.Context, req *agentsv1.UpdateHookRequest) (*agentsv1.Hook, error) {
	resp, err := c.gateway.UpdateHook(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update hook: %w", err)
	}
	return resp.Hook, nil
}

func (c *Client) DeleteHook(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete hook", func() error {
		_, err := c.gateway.DeleteHook(ctx, &agentsv1.DeleteHookRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete hook: %w", err)
		}
		return nil
	})
}
