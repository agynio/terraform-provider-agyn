package agentapi

import (
	"context"
	"fmt"

	imagesv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/images/v1"
)

func (c *Client) CreateImage(ctx context.Context, req *imagesv1.CreateImageRequest) (*imagesv1.Image, error) {
	return withConflictRetry(ctx, "create image", func() (*imagesv1.Image, error) {
		resp, err := c.imagesGateway.CreateImage(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create image: %w", err)
		}
		return resp.Image, nil
	})
}

func (c *Client) GetImage(ctx context.Context, id string) (*imagesv1.Image, error) {
	resp, err := c.imagesGateway.GetImage(ctx, &imagesv1.GetImageRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get image: %w", err)
	}
	return resp.Image, nil
}

func (c *Client) UpdateImage(ctx context.Context, req *imagesv1.UpdateImageRequest) (*imagesv1.Image, error) {
	resp, err := c.imagesGateway.UpdateImage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update image: %w", err)
	}
	return resp.Image, nil
}

func (c *Client) DeleteImage(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete image", func() error {
		if _, err := c.imagesGateway.DeleteImage(ctx, &imagesv1.DeleteImageRequest{Id: id}); err != nil {
			return fmt.Errorf("delete image: %w", err)
		}
		return nil
	})
}
