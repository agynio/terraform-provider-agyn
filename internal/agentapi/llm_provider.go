package agentapi

import (
	"context"
	"fmt"

	llmv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/llm/v1"
)

func (c *Client) CreateLLMProvider(ctx context.Context, req *llmv1.CreateLLMProviderRequest) (*llmv1.LLMProvider, error) {
	return withConflictRetry(ctx, "create llm provider", func() (*llmv1.LLMProvider, error) {
		resp, err := c.llmGateway.CreateLLMProvider(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create llm provider: %w", err)
		}
		return resp.Provider, nil
	})
}

func (c *Client) GetLLMProvider(ctx context.Context, id string) (*llmv1.LLMProvider, error) {
	resp, err := c.llmGateway.GetLLMProvider(ctx, &llmv1.GetLLMProviderRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get llm provider: %w", err)
	}
	return resp.Provider, nil
}

func (c *Client) UpdateLLMProvider(ctx context.Context, req *llmv1.UpdateLLMProviderRequest) (*llmv1.LLMProvider, error) {
	resp, err := c.llmGateway.UpdateLLMProvider(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update llm provider: %w", err)
	}
	return resp.Provider, nil
}

func (c *Client) DeleteLLMProvider(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete llm provider", func() error {
		_, err := c.llmGateway.DeleteLLMProvider(ctx, &llmv1.DeleteLLMProviderRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete llm provider: %w", err)
		}
		return nil
	})
}
