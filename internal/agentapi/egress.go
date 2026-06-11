package agentapi

import (
	"context"
	"fmt"

	egressv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/egress/v1"
)

func (c *Client) CreateEgressRule(ctx context.Context, req *egressv1.CreateEgressRuleRequest) (*egressv1.EgressRule, error) {
	return withConflictRetry(ctx, "create egress rule", func() (*egressv1.EgressRule, error) {
		resp, err := c.egressGateway.CreateEgressRule(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create egress rule: %w", err)
		}
		return resp.EgressRule, nil
	})
}

func (c *Client) GetEgressRule(ctx context.Context, id string) (*egressv1.EgressRule, error) {
	resp, err := c.egressGateway.GetEgressRule(ctx, &egressv1.GetEgressRuleRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get egress rule: %w", err)
	}
	return resp.EgressRule, nil
}

func (c *Client) UpdateEgressRule(ctx context.Context, req *egressv1.UpdateEgressRuleRequest) (*egressv1.EgressRule, error) {
	resp, err := c.egressGateway.UpdateEgressRule(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update egress rule: %w", err)
	}
	return resp.EgressRule, nil
}

func (c *Client) DeleteEgressRule(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete egress rule", func() error {
		_, err := c.egressGateway.DeleteEgressRule(ctx, &egressv1.DeleteEgressRuleRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete egress rule: %w", err)
		}
		return nil
	})
}

func (c *Client) CreateEgressRuleAttachment(ctx context.Context, req *egressv1.CreateEgressRuleAttachmentRequest) (*egressv1.EgressRuleAttachment, error) {
	return withConflictRetry(ctx, "create egress rule attachment", func() (*egressv1.EgressRuleAttachment, error) {
		resp, err := c.egressGateway.CreateEgressRuleAttachment(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create egress rule attachment: %w", err)
		}
		return resp.EgressRuleAttachment, nil
	})
}

func (c *Client) GetEgressRuleAttachmentByRuleAndAgent(ctx context.Context, organizationID string, ruleID string, agentID string) (*egressv1.EgressRuleAttachment, error) {
	resp, err := c.egressGateway.ListEgressRuleAttachments(ctx, &egressv1.ListEgressRuleAttachmentsRequest{OrganizationId: organizationID, RuleId: &ruleID, AgentId: &agentID})
	if err != nil {
		return nil, fmt.Errorf("list egress rule attachments: %w", err)
	}
	attachments := resp.GetEgressRuleAttachments()
	if len(attachments) == 0 {
		return nil, ErrNotFound
	}
	return attachments[0], nil
}

func (c *Client) DeleteEgressRuleAttachment(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete egress rule attachment", func() error {
		_, err := c.egressGateway.DeleteEgressRuleAttachment(ctx, &egressv1.DeleteEgressRuleAttachmentRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete egress rule attachment: %w", err)
		}
		return nil
	})
}
