package agentapi

import (
	"context"
	"fmt"

	organizationsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/organizations/v1"
)

func (c *Client) CreateOrganization(ctx context.Context, req *organizationsv1.CreateOrganizationRequest) (*organizationsv1.Organization, error) {
	return withConflictRetry(ctx, "create organization", func() (*organizationsv1.Organization, error) {
		resp, err := c.organizationsGateway.CreateOrganization(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create organization: %w", err)
		}
		return resp.Organization, nil
	})
}

func (c *Client) GetOrganization(ctx context.Context, id string) (*organizationsv1.Organization, error) {
	resp, err := c.organizationsGateway.GetOrganization(ctx, &organizationsv1.GetOrganizationRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get organization: %w", err)
	}
	return resp.Organization, nil
}

func (c *Client) UpdateOrganization(ctx context.Context, req *organizationsv1.UpdateOrganizationRequest) (*organizationsv1.Organization, error) {
	resp, err := c.organizationsGateway.UpdateOrganization(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update organization: %w", err)
	}
	return resp.Organization, nil
}

func (c *Client) DeleteOrganization(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete organization", func() error {
		_, err := c.organizationsGateway.DeleteOrganization(ctx, &organizationsv1.DeleteOrganizationRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete organization: %w", err)
		}
		return nil
	})
}
