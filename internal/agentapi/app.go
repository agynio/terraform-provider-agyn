package agentapi

import (
	"context"
	"fmt"

	appsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/apps/v1"
)

type CreateAppResult struct {
	App          *appsv1.App
	ServiceToken string
}

func (c *Client) CreateApp(ctx context.Context, req *appsv1.CreateAppRequest) (*CreateAppResult, error) {
	return withConflictRetry(ctx, "create app", func() (*CreateAppResult, error) {
		resp, err := c.appsGateway.CreateApp(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create app: %w", err)
		}
		return &CreateAppResult{App: resp.App, ServiceToken: resp.ServiceToken}, nil
	})
}

func (c *Client) GetApp(ctx context.Context, id string) (*appsv1.App, error) {
	resp, err := c.appsGateway.GetApp(ctx, &appsv1.GetAppRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get app: %w", err)
	}
	return resp.App, nil
}

func (c *Client) DeleteApp(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete app", func() error {
		_, err := c.appsGateway.DeleteApp(ctx, &appsv1.DeleteAppRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete app: %w", err)
		}
		return nil
	})
}
