package agentapi

import (
	"context"
	"fmt"

	appsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/apps/v1"
)

func (c *Client) InstallApp(ctx context.Context, req *appsv1.InstallAppRequest) (*appsv1.Installation, error) {
	return withConflictRetry(ctx, "install app", func() (*appsv1.Installation, error) {
		resp, err := c.appsGateway.InstallApp(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("install app: %w", err)
		}
		return resp.Installation, nil
	})
}

func (c *Client) GetInstallation(ctx context.Context, id string) (*appsv1.Installation, error) {
	resp, err := c.appsGateway.GetInstallation(ctx, &appsv1.GetInstallationRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get installation: %w", err)
	}
	return resp.Installation, nil
}

func (c *Client) UpdateInstallation(ctx context.Context, req *appsv1.UpdateInstallationRequest) (*appsv1.Installation, error) {
	resp, err := c.appsGateway.UpdateInstallation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update installation: %w", err)
	}
	return resp.Installation, nil
}

func (c *Client) UninstallApp(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "uninstall app", func() error {
		_, err := c.appsGateway.UninstallApp(ctx, &appsv1.UninstallAppRequest{Id: id})
		if err != nil {
			return fmt.Errorf("uninstall app: %w", err)
		}
		return nil
	})
}
