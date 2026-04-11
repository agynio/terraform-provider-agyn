package resources

import (
	"testing"

	llmv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/llm/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestToProtoAuthMethod(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected llmv1.AuthMethod
	}{
		{
			name:     "bearer",
			input:    "bearer",
			expected: llmv1.AuthMethod_AUTH_METHOD_BEARER,
		},
		{
			name:     "x_api_key",
			input:    "x_api_key",
			expected: llmv1.AuthMethod_AUTH_METHOD_X_API_KEY,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toProtoAuthMethod(tt.input); got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestFromProtoAuthMethod(t *testing.T) {
	tests := []struct {
		name     string
		input    llmv1.AuthMethod
		expected string
	}{
		{
			name:     "bearer",
			input:    llmv1.AuthMethod_AUTH_METHOD_BEARER,
			expected: "bearer",
		},
		{
			name:     "x_api_key",
			input:    llmv1.AuthMethod_AUTH_METHOD_X_API_KEY,
			expected: "x_api_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fromProtoAuthMethod(tt.input); got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestToProtoProtocol(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected llmv1.Protocol
	}{
		{
			name:     "responses",
			input:    "responses",
			expected: llmv1.Protocol_PROTOCOL_RESPONSES,
		},
		{
			name:     "anthropic_messages",
			input:    "anthropic_messages",
			expected: llmv1.Protocol_PROTOCOL_ANTHROPIC_MESSAGES,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toProtoProtocol(tt.input); got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestFromProtoProtocol(t *testing.T) {
	tests := []struct {
		name     string
		input    llmv1.Protocol
		expected string
	}{
		{
			name:     "responses",
			input:    llmv1.Protocol_PROTOCOL_RESPONSES,
			expected: "responses",
		},
		{
			name:     "anthropic_messages",
			input:    llmv1.Protocol_PROTOCOL_ANTHROPIC_MESSAGES,
			expected: "anthropic_messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fromProtoProtocol(tt.input); got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestUpdateProtocolPointer(t *testing.T) {
	protocolPointer := func(value llmv1.Protocol) *llmv1.Protocol {
		return &value
	}

	tests := []struct {
		name     string
		plan     types.String
		prior    types.String
		expected *llmv1.Protocol
	}{
		{
			name:     "unknown plan",
			plan:     types.StringUnknown(),
			prior:    types.StringValue("responses"),
			expected: nil,
		},
		{
			name:     "null plan with null prior",
			plan:     types.StringNull(),
			prior:    types.StringNull(),
			expected: nil,
		},
		{
			name:     "null plan with prior set",
			plan:     types.StringNull(),
			prior:    types.StringValue("responses"),
			expected: protocolPointer(llmv1.Protocol_PROTOCOL_UNSPECIFIED),
		},
		{
			name:     "set plan with prior set",
			plan:     types.StringValue("responses"),
			prior:    types.StringValue("anthropic_messages"),
			expected: protocolPointer(llmv1.Protocol_PROTOCOL_RESPONSES),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updateProtocolPointer(tt.plan, tt.prior)
			if tt.expected == nil {
				if got != nil {
					t.Fatalf("expected nil pointer, got %v", got)
				}
				return
			}
			if got == nil || *got != *tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
