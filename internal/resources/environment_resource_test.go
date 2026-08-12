package resources

import (
	"testing"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEnvironmentAvailabilityFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected agentsv1.EnvironmentAvailability
	}{
		{
			name:     "internal",
			input:    "internal",
			expected: agentsv1.EnvironmentAvailability_ENVIRONMENT_AVAILABILITY_INTERNAL,
		},
		{
			name:     "private",
			input:    "private",
			expected: agentsv1.EnvironmentAvailability_ENVIRONMENT_AVAILABILITY_PRIVATE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := environmentAvailabilityFromString(tt.input); got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestEnvironmentAvailabilityToString(t *testing.T) {
	tests := []struct {
		name     string
		input    agentsv1.EnvironmentAvailability
		expected string
	}{
		{
			name:     "internal",
			input:    agentsv1.EnvironmentAvailability_ENVIRONMENT_AVAILABILITY_INTERNAL,
			expected: "internal",
		},
		{
			name:     "private",
			input:    agentsv1.EnvironmentAvailability_ENVIRONMENT_AVAILABILITY_PRIVATE,
			expected: "private",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := environmentAvailabilityToString(tt.input); got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

// An update that omits availability leaves the API holding the prior value,
// which then contradicts the plan Terraform just applied.
func TestEnvironmentAvailabilityUpdate(t *testing.T) {
	if got := environmentAvailabilityUpdate(types.StringValue("private")); got == nil ||
		*got != agentsv1.EnvironmentAvailability_ENVIRONMENT_AVAILABILITY_PRIVATE {
		t.Fatalf("expected private to be sent, got %v", got)
	}
	if got := environmentAvailabilityUpdate(types.StringUnknown()); got != nil {
		t.Fatalf("expected an unknown value to send nothing, got %v", *got)
	}
}
