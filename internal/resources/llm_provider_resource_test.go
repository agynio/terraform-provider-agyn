package resources

import (
	"testing"

	llmv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/llm/v1"
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
		if got := toProtoAuthMethod(tt.input); got != tt.expected {
			t.Fatalf("%s: expected %v, got %v", tt.name, tt.expected, got)
		}
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
		if got := fromProtoAuthMethod(tt.input); got != tt.expected {
			t.Fatalf("%s: expected %q, got %q", tt.name, tt.expected, got)
		}
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
		if got := toProtoProtocol(tt.input); got != tt.expected {
			t.Fatalf("%s: expected %v, got %v", tt.name, tt.expected, got)
		}
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
		if got := fromProtoProtocol(tt.input); got != tt.expected {
			t.Fatalf("%s: expected %q, got %q", tt.name, tt.expected, got)
		}
	}
}
