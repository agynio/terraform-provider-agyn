package agentapi

import (
	"context"
	"fmt"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
)

func (c *Client) CreateMcp(ctx context.Context, req *agentsv1.CreateMcpRequest) (*agentsv1.Mcp, error) {
	return withConflictRetry(ctx, "create mcp", func() (*agentsv1.Mcp, error) {
		resp, err := c.gateway.CreateMcp(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create mcp: %w", err)
		}
		return resp.Mcp, nil
	})
}

func (c *Client) GetMcp(ctx context.Context, id string) (*agentsv1.Mcp, error) {
	resp, err := c.gateway.GetMcp(ctx, &agentsv1.GetMcpRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get mcp: %w", err)
	}
	return resp.Mcp, nil
}

func (c *Client) UpdateMcp(ctx context.Context, req *agentsv1.UpdateMcpRequest) (*agentsv1.Mcp, error) {
	resp, err := c.gateway.UpdateMcp(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update mcp: %w", err)
	}
	return resp.Mcp, nil
}

func (c *Client) DeleteMcp(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete mcp", func() error {
		_, err := c.gateway.DeleteMcp(ctx, &agentsv1.DeleteMcpRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete mcp: %w", err)
		}
		return nil
	})
}
