package agentapi

import (
	"context"
	"fmt"

	llmv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/llm/v1"
)

func (c *Client) CreateModel(ctx context.Context, req *llmv1.CreateModelRequest) (*llmv1.Model, error) {
	return withConflictRetry(ctx, "create model", func() (*llmv1.Model, error) {
		resp, err := c.llmGateway.CreateModel(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create model: %w", err)
		}
		return resp.Model, nil
	})
}

func (c *Client) GetModel(ctx context.Context, id string) (*llmv1.Model, error) {
	resp, err := c.llmGateway.GetModel(ctx, &llmv1.GetModelRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get model: %w", err)
	}
	return resp.Model, nil
}

func (c *Client) UpdateModel(ctx context.Context, req *llmv1.UpdateModelRequest) (*llmv1.Model, error) {
	resp, err := c.llmGateway.UpdateModel(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update model: %w", err)
	}
	return resp.Model, nil
}

func (c *Client) DeleteModel(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete model", func() error {
		_, err := c.llmGateway.DeleteModel(ctx, &llmv1.DeleteModelRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete model: %w", err)
		}
		return nil
	})
}
