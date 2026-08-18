package agentapi

import (
	"context"
	"fmt"

	runnersv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/runners/v1"
)

type RegisterRunnerResult struct {
	Runner       *runnersv1.Runner
	ServiceToken string
}

func (c *Client) RegisterRunner(ctx context.Context, req *runnersv1.RegisterRunnerRequest) (*RegisterRunnerResult, error) {
	return withConflictRetry(ctx, "register runner", func() (*RegisterRunnerResult, error) {
		resp, err := c.runnersGateway.RegisterRunner(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("register runner: %w", err)
		}
		return &RegisterRunnerResult{Runner: resp.Runner, ServiceToken: resp.ServiceToken}, nil
	})
}

func (c *Client) GetRunner(ctx context.Context, id string) (*runnersv1.Runner, error) {
	resp, err := c.runnersGateway.GetRunner(ctx, &runnersv1.GetRunnerRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get runner: %w", err)
	}
	return resp.Runner, nil
}

func (c *Client) UpdateRunner(ctx context.Context, req *runnersv1.UpdateRunnerRequest) (*runnersv1.Runner, error) {
	resp, err := c.runnersGateway.UpdateRunner(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update runner: %w", err)
	}
	return resp.Runner, nil
}

func (c *Client) DeleteRunner(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete runner", func() error {
		_, err := c.runnersGateway.DeleteRunner(ctx, &runnersv1.DeleteRunnerRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete runner: %w", err)
		}
		return nil
	})
}

// ListRunners walks every page so a lookup by name sees every runner the
// organization can place workloads on.
func (c *Client) ListRunners(ctx context.Context, organizationID string) ([]*runnersv1.Runner, error) {
	var runners []*runnersv1.Runner
	var pageToken string
	for {
		req := &runnersv1.ListRunnersRequest{PageSize: runnerListPageSize, PageToken: pageToken}
		if organizationID != "" {
			req.OrganizationId = &organizationID
		}
		resp, err := c.runnersGateway.ListRunners(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("list runners: %w", err)
		}
		runners = append(runners, resp.Runners...)
		pageToken = resp.NextPageToken
		if pageToken == "" {
			return runners, nil
		}
	}
}

const runnerListPageSize = 100
