package agentapi

import (
	"context"
	"fmt"
	"time"

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

// WaitForVersion blocks until the catalog has discovered at least one version
// of the image, so a resource naming a tag of it can resolve. Discovery is
// asynchronous, and an image with no versions yet resolves to nothing.
func (c *Client) WaitForVersion(ctx context.Context, id string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		resp, err := c.imagesGateway.ListVersions(ctx, &imagesv1.ListVersionsRequest{ImageId: id})
		if err != nil {
			return fmt.Errorf("list versions: %w", err)
		}
		if len(resp.Versions) > 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("no version of image %s was discovered within %s", id, timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
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
