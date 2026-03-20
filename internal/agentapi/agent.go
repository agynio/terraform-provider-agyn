package agentapi

import (
	"context"
	"fmt"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
)

func (c *Client) CreateAgent(ctx context.Context, req *agentsv1.CreateAgentRequest) (*agentsv1.Agent, error) {
	return withConflictRetry(ctx, "create agent", func() (*agentsv1.Agent, error) {
		resp, err := c.gateway.CreateAgent(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create agent: %w", err)
		}
		return resp.Agent, nil
	})
}

func (c *Client) GetAgent(ctx context.Context, id string) (*agentsv1.Agent, error) {
	resp, err := c.gateway.GetAgent(ctx, &agentsv1.GetAgentRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get agent: %w", err)
	}
	return resp.Agent, nil
}

func (c *Client) UpdateAgent(ctx context.Context, req *agentsv1.UpdateAgentRequest) (*agentsv1.Agent, error) {
	resp, err := c.gateway.UpdateAgent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update agent: %w", err)
	}
	return resp.Agent, nil
}

func (c *Client) DeleteAgent(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete agent", func() error {
		_, err := c.gateway.DeleteAgent(ctx, &agentsv1.DeleteAgentRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete agent: %w", err)
		}
		return nil
	})
}
