package agentapi

import (
	"context"
	"fmt"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
)

func (c *Client) CreateSkill(ctx context.Context, req *agentsv1.CreateSkillRequest) (*agentsv1.Skill, error) {
	return withConflictRetry(ctx, "create skill", func() (*agentsv1.Skill, error) {
		resp, err := c.gateway.CreateSkill(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create skill: %w", err)
		}
		return resp.Skill, nil
	})
}

func (c *Client) GetSkill(ctx context.Context, id string) (*agentsv1.Skill, error) {
	resp, err := c.gateway.GetSkill(ctx, &agentsv1.GetSkillRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get skill: %w", err)
	}
	return resp.Skill, nil
}

func (c *Client) UpdateSkill(ctx context.Context, req *agentsv1.UpdateSkillRequest) (*agentsv1.Skill, error) {
	resp, err := c.gateway.UpdateSkill(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update skill: %w", err)
	}
	return resp.Skill, nil
}

func (c *Client) DeleteSkill(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete skill", func() error {
		_, err := c.gateway.DeleteSkill(ctx, &agentsv1.DeleteSkillRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete skill: %w", err)
		}
		return nil
	})
}
