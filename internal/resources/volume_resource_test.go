package resources

import (
	"testing"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
)

// An imported volume has no prior target in state, so the target has to come
// back off the API or the next plan replaces the volume it just imported.
func TestVolumeStateFromReadsTheTarget(t *testing.T) {
	environmentVolume := &agentsv1.Volume{
		Meta:      &agentsv1.EntityMeta{Id: "11111111-1111-1111-1111-111111111111"},
		Name:      "workspace",
		MountPath: "/workspace",
		Size:      "10Gi",
		Target:    &agentsv1.Volume_EnvironmentId{EnvironmentId: "22222222-2222-2222-2222-222222222222"},
	}
	state := volumeStateFrom(environmentVolume)
	if got := state.EnvironmentID.ValueString(); got != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("expected the environment target, got %q", got)
	}
	if !state.McpID.IsNull() {
		t.Fatalf("expected mcp_id to stay null, got %q", state.McpID.ValueString())
	}

	mcpVolume := &agentsv1.Volume{
		Meta:      &agentsv1.EntityMeta{Id: "33333333-3333-3333-3333-333333333333"},
		Name:      "index",
		MountPath: "/var/lib/index",
		Target:    &agentsv1.Volume_McpId{McpId: "44444444-4444-4444-4444-444444444444"},
	}
	state = volumeStateFrom(mcpVolume)
	if got := state.McpID.ValueString(); got != "44444444-4444-4444-4444-444444444444" {
		t.Fatalf("expected the mcp target, got %q", got)
	}
	if !state.EnvironmentID.IsNull() {
		t.Fatalf("expected environment_id to stay null, got %q", state.EnvironmentID.ValueString())
	}
	// No size is what makes a volume ephemeral scratch rather than a disk.
	if !state.Size.IsNull() {
		t.Fatalf("expected size to stay null, got %q", state.Size.ValueString())
	}
}
