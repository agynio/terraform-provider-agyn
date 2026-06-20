package agentapi

import (
	"context"
	"fmt"

	networksv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/networks/v1"
)

type CreatedTunnelCredential struct {
	TunnelCredential *networksv1.TunnelCredential
	EnrollmentJWT    string
}

func (c *Client) CreateNetwork(ctx context.Context, req *networksv1.CreateNetworkRequest) (*networksv1.Network, error) {
	return withConflictRetry(ctx, "create network", func() (*networksv1.Network, error) {
		resp, err := c.networksGateway.CreateNetwork(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create network: %w", err)
		}
		return resp.Network, nil
	})
}

func (c *Client) GetNetwork(ctx context.Context, id string) (*networksv1.Network, error) {
	resp, err := c.networksGateway.GetNetwork(ctx, &networksv1.GetNetworkRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get network: %w", err)
	}
	return resp.Network, nil
}

func (c *Client) UpdateNetwork(ctx context.Context, req *networksv1.UpdateNetworkRequest) (*networksv1.Network, error) {
	resp, err := c.networksGateway.UpdateNetwork(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update network: %w", err)
	}
	return resp.Network, nil
}

func (c *Client) DeleteNetwork(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete network", func() error {
		_, err := c.networksGateway.DeleteNetwork(ctx, &networksv1.DeleteNetworkRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete network: %w", err)
		}
		return nil
	})
}

func (c *Client) CreateTunnelCredential(ctx context.Context, req *networksv1.CreateTunnelCredentialRequest) (*CreatedTunnelCredential, error) {
	return withConflictRetry(ctx, "create tunnel credential", func() (*CreatedTunnelCredential, error) {
		resp, err := c.networksGateway.CreateTunnelCredential(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create tunnel credential: %w", err)
		}
		return &CreatedTunnelCredential{TunnelCredential: resp.TunnelCredential, EnrollmentJWT: resp.EnrollmentJwt}, nil
	})
}

func (c *Client) GetTunnelCredential(ctx context.Context, id string) (*networksv1.TunnelCredential, error) {
	resp, err := c.networksGateway.GetTunnelCredential(ctx, &networksv1.GetTunnelCredentialRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get tunnel credential: %w", err)
	}
	return resp.TunnelCredential, nil
}

func (c *Client) DeleteTunnelCredential(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete tunnel credential", func() error {
		_, err := c.networksGateway.DeleteTunnelCredential(ctx, &networksv1.DeleteTunnelCredentialRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete tunnel credential: %w", err)
		}
		return nil
	})
}

func (c *Client) CreatePrivateResource(ctx context.Context, req *networksv1.CreatePrivateResourceRequest) (*networksv1.PrivateResource, error) {
	return withConflictRetry(ctx, "create private resource", func() (*networksv1.PrivateResource, error) {
		resp, err := c.networksGateway.CreatePrivateResource(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create private resource: %w", err)
		}
		return resp.PrivateResource, nil
	})
}

func (c *Client) GetPrivateResource(ctx context.Context, id string) (*networksv1.PrivateResource, error) {
	resp, err := c.networksGateway.GetPrivateResource(ctx, &networksv1.GetPrivateResourceRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get private resource: %w", err)
	}
	return resp.PrivateResource, nil
}

func (c *Client) UpdatePrivateResource(ctx context.Context, req *networksv1.UpdatePrivateResourceRequest) (*networksv1.PrivateResource, error) {
	resp, err := c.networksGateway.UpdatePrivateResource(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update private resource: %w", err)
	}
	return resp.PrivateResource, nil
}

func (c *Client) DeletePrivateResource(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete private resource", func() error {
		_, err := c.networksGateway.DeletePrivateResource(ctx, &networksv1.DeletePrivateResourceRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete private resource: %w", err)
		}
		return nil
	})
}

func (c *Client) CreatePrivateResourceAccess(ctx context.Context, req *networksv1.CreatePrivateResourceAccessRequest) (*networksv1.PrivateResourceAccess, error) {
	return withConflictRetry(ctx, "create private resource access", func() (*networksv1.PrivateResourceAccess, error) {
		resp, err := c.networksGateway.CreatePrivateResourceAccess(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create private resource access: %w", err)
		}
		return resp.PrivateResourceAccess, nil
	})
}

func (c *Client) GetPrivateResourceAccessByResourceAndPrincipal(ctx context.Context, privateResourceID string, principalType networksv1.PrivateResourceAccessPrincipalType, principalID string) (*networksv1.PrivateResourceAccess, error) {
	pageToken := ""
	for {
		resp, err := c.networksGateway.ListPrivateResourceAccess(ctx, &networksv1.ListPrivateResourceAccessRequest{PrivateResourceId: &privateResourceID, PrincipalType: principalType.Enum(), PrincipalId: &principalID, PageSize: lookupPageSize, PageToken: pageToken})
		if err != nil {
			return nil, fmt.Errorf("list private resource access: %w", err)
		}
		for _, access := range resp.GetPrivateResourceAccess() {
			if access.GetPrivateResourceId() == privateResourceID && access.GetPrincipalType() == principalType && access.GetPrincipalId() == principalID {
				return access, nil
			}
		}
		pageToken = resp.GetNextPageToken()
		if pageToken == "" {
			return nil, ErrNotFound
		}
	}
}

func (c *Client) DeletePrivateResourceAccess(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete private resource access", func() error {
		_, err := c.networksGateway.DeletePrivateResourceAccess(ctx, &networksv1.DeletePrivateResourceAccessRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete private resource access: %w", err)
		}
		return nil
	})
}
