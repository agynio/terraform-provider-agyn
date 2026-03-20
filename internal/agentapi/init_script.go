package agentapi

import (
	"context"
	"fmt"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
)

func (c *Client) CreateInitScript(ctx context.Context, req *agentsv1.CreateInitScriptRequest) (*agentsv1.InitScript, error) {
	return withConflictRetry(ctx, "create init script", func() (*agentsv1.InitScript, error) {
		resp, err := c.gateway.CreateInitScript(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create init script: %w", err)
		}
		return resp.InitScript, nil
	})
}

func (c *Client) GetInitScript(ctx context.Context, id string) (*agentsv1.InitScript, error) {
	resp, err := c.gateway.GetInitScript(ctx, &agentsv1.GetInitScriptRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get init script: %w", err)
	}
	return resp.InitScript, nil
}

func (c *Client) UpdateInitScript(ctx context.Context, req *agentsv1.UpdateInitScriptRequest) (*agentsv1.InitScript, error) {
	resp, err := c.gateway.UpdateInitScript(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update init script: %w", err)
	}
	return resp.InitScript, nil
}

func (c *Client) DeleteInitScript(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete init script", func() error {
		_, err := c.gateway.DeleteInitScript(ctx, &agentsv1.DeleteInitScriptRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete init script: %w", err)
		}
		return nil
	})
}
