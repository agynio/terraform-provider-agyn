package agentapi

import (
	"context"
	"fmt"

	usersv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/users/v1"
)

func (c *Client) CreateUser(ctx context.Context, req *usersv1.CreateUserRequest) (*usersv1.User, error) {
	return withConflictRetry(ctx, "create user", func() (*usersv1.User, error) {
		resp, err := c.usersGateway.CreateUser(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}
		return resp.User, nil
	})
}

func (c *Client) GetUser(ctx context.Context, identityID string) (*usersv1.User, usersv1.ClusterRole, error) {
	resp, err := c.usersGateway.GetUser(ctx, &usersv1.GetUserRequest{IdentityId: identityID})
	if err != nil {
		return nil, usersv1.ClusterRole_CLUSTER_ROLE_UNSPECIFIED, fmt.Errorf("get user: %w", err)
	}
	return resp.User, resp.ClusterRole, nil
}

func (c *Client) UpdateUser(ctx context.Context, req *usersv1.UpdateUserRequest) (*usersv1.User, error) {
	resp, err := c.usersGateway.UpdateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return resp.User, nil
}

func (c *Client) DeleteUser(ctx context.Context, identityID string) error {
	return withConflictRetryNoResult(ctx, "delete user", func() error {
		_, err := c.usersGateway.DeleteUser(ctx, &usersv1.DeleteUserRequest{IdentityId: identityID})
		if err != nil {
			return fmt.Errorf("delete user: %w", err)
		}
		return nil
	})
}
