package resources

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-agyn/internal/teamapi"
)

func TestAgentUpgradeStateV0(t *testing.T) {
	config := teamapi.AgentConfig{
		Name:                      stringPtr("agent"),
		Role:                      stringPtr("assistant"),
		Model:                     stringPtr("gpt-4o"),
		SystemPrompt:              stringPtr("hello"),
		DebounceMs:                int64Ptr(150),
		WhenBusy:                  stringPtr("wait"),
		ProcessBuffer:             stringPtr("allTogether"),
		SendFinalResponseToThread: boolPtr(true),
		RestrictOutput:            boolPtr(true),
		RestrictionMessage:        stringPtr("restricted"),
		RestrictionMaxInjections:  int64Ptr(2),
		SummarizationKeepTokens:   int64Ptr(100),
		SummarizationMaxTokens:    int64Ptr(200),
	}
	configJSON := mustMarshal(t, config)
	prior := agentModelV0{
		ID:          types.StringValue("agent-id"),
		Title:       types.StringValue("Agent"),
		Description: types.StringValue("Agent description"),
		Config:      types.StringValue(configJSON),
	}

	state, diags := runUpgrade(t, &agentResource{}, prior)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	var upgraded agentModel
	ctx := context.Background()
	if diags := state.Get(ctx, &upgraded); diags.HasError() {
		t.Fatalf("failed to get upgraded state: %v", diags)
	}
	if !upgraded.Config.IsNull() {
		t.Fatalf("expected config to be null")
	}
	assertStringValue(t, upgraded.Name, "agent")
	assertStringValue(t, upgraded.Role, "assistant")
	assertStringValue(t, upgraded.Model, "gpt-4o")
	assertStringValue(t, upgraded.SystemPrompt, "hello")
	assertInt64Value(t, upgraded.DebounceMs, 150)
	assertStringValue(t, upgraded.WhenBusy, "wait")
	assertStringValue(t, upgraded.ProcessBuffer, "allTogether")
	assertBoolValue(t, upgraded.SendFinalResponseToThread, true)
	assertBoolValue(t, upgraded.RestrictOutput, true)
	assertStringValue(t, upgraded.RestrictionMessage, "restricted")
	assertInt64Value(t, upgraded.RestrictionMaxInjections, 2)
	assertInt64Value(t, upgraded.SummarizationKeepTokens, 100)
	assertInt64Value(t, upgraded.SummarizationMaxTokens, 200)
}

func TestMCPServerUpgradeStateV0(t *testing.T) {
	config := teamapi.MCPServerConfig{
		Namespace:           stringPtr("mcp"),
		Command:             stringPtr("mcp start --stdio"),
		Workdir:             stringPtr("/srv"),
		RequestTimeoutMs:    int64Ptr(5000),
		StartupTimeoutMs:    int64Ptr(1000),
		HeartbeatIntervalMs: int64Ptr(2000),
		StaleTimeoutMs:      int64Ptr(10000),
		Restart: &teamapi.RestartPolicy{
			MaxAttempts: int64Ptr(3),
			BackoffMs:   int64Ptr(250),
		},
		Env: []teamapi.EnvVar{
			{
				Name: "TOKEN",
				ValueRef: &teamapi.EnvValueRef{
					Kind:  "vault",
					Mount: stringPtr("secret"),
					Path:  stringPtr("path"),
					Key:   stringPtr("token"),
				},
			},
		},
	}
	configJSON := mustMarshal(t, config)
	prior := mcpServerModelV0{
		ID:          types.StringValue("mcp-id"),
		Title:       types.StringValue("MCP"),
		Description: types.StringValue("MCP description"),
		Config:      types.StringValue(configJSON),
	}

	state, diags := runUpgrade(t, &mcpServerResource{}, prior)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	var upgraded mcpServerModel
	ctx := context.Background()
	if diags := state.Get(ctx, &upgraded); diags.HasError() {
		t.Fatalf("failed to get upgraded state: %v", diags)
	}
	if !upgraded.Config.IsNull() {
		t.Fatalf("expected config to be null")
	}
	assertStringValue(t, upgraded.Namespace, "mcp")
	assertStringValue(t, upgraded.Command, "mcp start --stdio")
	assertStringValue(t, upgraded.Workdir, "/srv")
	assertInt64Value(t, upgraded.RequestTimeoutMs, 5000)
	assertInt64Value(t, upgraded.StartupTimeoutMs, 1000)
	assertInt64Value(t, upgraded.HeartbeatIntervalMs, 2000)
	assertInt64Value(t, upgraded.StaleTimeoutMs, 10000)
	if upgraded.Restart == nil {
		t.Fatalf("expected restart to be set")
	}
	assertInt64Value(t, upgraded.Restart.MaxAttempts, 3)
	assertInt64Value(t, upgraded.Restart.BackoffMs, 250)
	if len(upgraded.Env) != 1 {
		t.Fatalf("expected 1 env entry, got %d", len(upgraded.Env))
	}
	if upgraded.Env[0].ValueRef == nil {
		t.Fatalf("expected env value_ref")
	}
	assertStringValue(t, upgraded.Env[0].ValueRef.Kind, "vault")
	assertStringValue(t, upgraded.Env[0].ValueRef.Mount, "secret")
	assertStringValue(t, upgraded.Env[0].ValueRef.Path, "path")
	assertStringValue(t, upgraded.Env[0].ValueRef.Key, "token")
}

func TestWorkspaceConfigurationUpgradeStateV0(t *testing.T) {
	cpuLimit := json.RawMessage("2")
	memoryLimit := json.RawMessage(`"4Gi"`)
	nix := json.RawMessage(`{"packages":["go"]}`)
	config := teamapi.WorkspaceConfigurationConfig{
		Image:         stringPtr("ubuntu:22.04"),
		InitialScript: stringPtr("echo hello"),
		CpuLimit:      &cpuLimit,
		MemoryLimit:   &memoryLimit,
		Platform:      stringPtr("auto"),
		EnableDinD:    boolPtr(false),
		TtlSeconds:    int64Ptr(600),
		Volumes: &teamapi.WorkspaceVolumes{
			Enabled:   boolPtr(true),
			MountPath: stringPtr("/workspace"),
		},
		Env: []teamapi.EnvVar{
			{
				Name:  "FOO",
				Value: stringPtr("bar"),
			},
		},
		Nix: &nix,
	}
	configJSON := mustMarshal(t, config)
	prior := workspaceConfigurationModelV0{
		ID:          types.StringValue("workspace-id"),
		Title:       types.StringValue("Workspace"),
		Description: types.StringValue("Workspace description"),
		Config:      types.StringValue(configJSON),
	}

	state, diags := runUpgrade(t, &workspaceConfigurationResource{}, prior)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	var upgraded workspaceConfigurationModel
	ctx := context.Background()
	if diags := state.Get(ctx, &upgraded); diags.HasError() {
		t.Fatalf("failed to get upgraded state: %v", diags)
	}
	if !upgraded.Config.IsNull() {
		t.Fatalf("expected config to be null")
	}
	assertStringValue(t, upgraded.Image, "ubuntu:22.04")
	assertStringValue(t, upgraded.InitialScript, "echo hello")
	assertStringValue(t, upgraded.CpuLimit, "2")
	assertStringValue(t, upgraded.MemoryLimit, "4Gi")
	assertStringValue(t, upgraded.Platform, "auto")
	assertBoolValue(t, upgraded.EnableDinD, false)
	assertInt64Value(t, upgraded.TtlSeconds, 600)
	if upgraded.Volumes == nil {
		t.Fatalf("expected volumes to be set")
	}
	assertBoolValue(t, upgraded.Volumes.Enabled, true)
	assertStringValue(t, upgraded.Volumes.MountPath, "/workspace")
	if len(upgraded.Env) != 1 {
		t.Fatalf("expected 1 env entry, got %d", len(upgraded.Env))
	}
	assertStringValue(t, upgraded.Env[0].Name, "FOO")
	assertStringValue(t, upgraded.Env[0].Value, "bar")
	if upgraded.Nix.IsNull() || upgraded.Nix.ValueString() != `{"packages":["go"]}` {
		t.Fatalf("unexpected nix value: %v", upgraded.Nix)
	}
}

func TestMemoryBucketUpgradeStateV0(t *testing.T) {
	config := teamapi.MemoryBucketConfig{
		Scope:            stringPtr("global"),
		CollectionPrefix: stringPtr("prefix"),
	}
	configJSON := mustMarshal(t, config)
	prior := memoryBucketModelV0{
		ID:          types.StringValue("bucket-id"),
		Title:       types.StringValue("Bucket"),
		Description: types.StringValue("Bucket description"),
		Config:      types.StringValue(configJSON),
	}

	state, diags := runUpgrade(t, &memoryBucketResource{}, prior)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	var upgraded memoryBucketModel
	ctx := context.Background()
	if diags := state.Get(ctx, &upgraded); diags.HasError() {
		t.Fatalf("failed to get upgraded state: %v", diags)
	}
	if !upgraded.Config.IsNull() {
		t.Fatalf("expected config to be null")
	}
	assertStringValue(t, upgraded.Scope, "global")
	assertStringValue(t, upgraded.CollectionPrefix, "prefix")
}

func runUpgrade(t *testing.T, r resource.ResourceWithUpgradeState, priorState any) (tfsdk.State, diag.Diagnostics) {
	t.Helper()
	ctx := context.Background()
	upgraders := r.UpgradeState(ctx)
	upgrader, ok := upgraders[0]
	if !ok {
		t.Fatalf("missing upgrader for version 0")
	}
	if upgrader.PriorSchema == nil {
		t.Fatalf("missing prior schema")
	}

	res, ok := r.(resource.Resource)
	if !ok {
		t.Fatalf("resource does not implement schema")
	}
	var schemaResp resource.SchemaResponse
	res.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	prior := tfsdk.State{Schema: *upgrader.PriorSchema}
	if diags := prior.Set(ctx, priorState); diags.HasError() {
		t.Fatalf("failed to set prior state: %v", diags)
	}

	resp := resource.UpgradeStateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	upgrader.StateUpgrader(ctx, resource.UpgradeStateRequest{State: &prior}, &resp)
	return resp.State, resp.Diagnostics
}

func mustMarshal(t *testing.T, value any) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	return string(raw)
}

func stringPtr(value string) *string {
	return &value
}

func int64Ptr(value int64) *int64 {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func assertStringValue(t *testing.T, value types.String, expected string) {
	t.Helper()
	if value.IsNull() || value.IsUnknown() || value.ValueString() != expected {
		t.Fatalf("expected %q, got %v", expected, value)
	}
}

func assertInt64Value(t *testing.T, value types.Int64, expected int64) {
	t.Helper()
	if value.IsNull() || value.IsUnknown() || value.ValueInt64() != expected {
		t.Fatalf("expected %d, got %v", expected, value)
	}
}

func assertBoolValue(t *testing.T, value types.Bool, expected bool) {
	t.Helper()
	if value.IsNull() || value.IsUnknown() || value.ValueBool() != expected {
		t.Fatalf("expected %t, got %v", expected, value)
	}
}
