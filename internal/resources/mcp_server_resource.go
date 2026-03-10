package resources

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-agyn/internal/teamapi"
)

type mcpServerResource struct {
	client *teamapi.Client
}

var _ resource.ResourceWithUpgradeState = &mcpServerResource{}

type mcpServerRestartModel struct {
	MaxAttempts types.Int64 `tfsdk:"max_attempts"`
	BackoffMs   types.Int64 `tfsdk:"backoff_ms"`
}

type mcpServerModel struct {
	ID                  types.String           `tfsdk:"id"`
	Title               types.String           `tfsdk:"title"`
	Description         types.String           `tfsdk:"description"`
	Config              types.String           `tfsdk:"config"`
	Namespace           types.String           `tfsdk:"namespace"`
	Command             types.String           `tfsdk:"command"`
	Workdir             types.String           `tfsdk:"workdir"`
	RequestTimeoutMs    types.Int64            `tfsdk:"request_timeout_ms"`
	StartupTimeoutMs    types.Int64            `tfsdk:"startup_timeout_ms"`
	HeartbeatIntervalMs types.Int64            `tfsdk:"heartbeat_interval_ms"`
	StaleTimeoutMs      types.Int64            `tfsdk:"stale_timeout_ms"`
	Restart             *mcpServerRestartModel `tfsdk:"restart"`
	Env                 []envVarModel          `tfsdk:"env"`
}

type mcpServerModelV0 struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Config      types.String `tfsdk:"config"`
}

func NewMCPServerResource() resource.Resource { return &mcpServerResource{} }

func (r *mcpServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_server"
}

func (r *mcpServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn MCP server.",
		Version:             1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the MCP server.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"title": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable title of the MCP server.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable description of the MCP server.",
			},
			"config": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Deprecated JSON-encoded MCP server configuration. Use structured attributes instead.",
				DeprecationMessage:  "Use structured configuration attributes instead of config.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("namespace"),
						path.MatchRoot("command"),
						path.MatchRoot("workdir"),
						path.MatchRoot("request_timeout_ms"),
						path.MatchRoot("startup_timeout_ms"),
						path.MatchRoot("heartbeat_interval_ms"),
						path.MatchRoot("stale_timeout_ms"),
						path.MatchRoot("restart"),
						path.MatchRoot("env"),
					),
				},
			},
			"namespace": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Namespace for the MCP server.",
			},
			"command": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Command to run for the MCP server.",
			},
			"workdir": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Working directory for the MCP server process.",
			},
			"request_timeout_ms": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Request timeout in milliseconds.",
			},
			"startup_timeout_ms": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Startup timeout in milliseconds.",
			},
			"heartbeat_interval_ms": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Heartbeat interval in milliseconds.",
			},
			"stale_timeout_ms": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Stale timeout in milliseconds.",
			},
			"restart": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Restart policy configuration.",
				Attributes: map[string]schema.Attribute{
					"max_attempts": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "Maximum restart attempts.",
					},
					"backoff_ms": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "Backoff duration in milliseconds between restarts.",
					},
				},
			},
			"env": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Environment variables for the MCP server process.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Environment variable name.",
						},
						"value": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Literal environment variable value.",
						},
						"value_ref": schema.SingleNestedAttribute{
							Optional:            true,
							MarkdownDescription: "Reference to a vault or variable-backed value.",
							Attributes: map[string]schema.Attribute{
								"kind": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "Reference kind (vault or variable).",
									Validators: []validator.String{
										stringvalidator.OneOf("vault", "variable"),
									},
								},
								"mount": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "Mount name for the referenced secret store.",
								},
								"path": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "Secret path within the referenced store.",
								},
								"key": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "Key within the referenced secret.",
								},
							},
						},
					},
					Validators: []validator.Object{
						objectvalidator.ExactlyOneOf(
							path.MatchRelative().AtName("value"),
							path.MatchRelative().AtName("value_ref"),
						),
					},
				},
			},
		},
	}
}

func (r *mcpServerResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
					},
					"title": schema.StringAttribute{
						Optional: true,
					},
					"description": schema.StringAttribute{
						Optional: true,
					},
					"config": schema.StringAttribute{
						Required: true,
					},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior mcpServerModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}

				config, diags := mcpServerConfigFromString(prior.Config)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgraded := mcpServerModel{
					ID:          prior.ID,
					Title:       prior.Title,
					Description: prior.Description,
					Config:      types.StringNull(),
				}
				applyMCPServerConfigToModel(&upgraded, config)

				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
	}
}

func (r *mcpServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*teamapi.Client)
	if !ok || client == nil {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Unable to obtain configured API client")
		return
	}
	r.client = client
}

func (r *mcpServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing MCP servers.")
		return
	}

	var plan mcpServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, diags := mcpServerConfigFromPlan(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	create := teamapi.MCPServerCreate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
		Config:      config,
	}

	server, err := r.client.CreateMCPServer(ctx, create)
	if err != nil {
		resp.Diagnostics.AddError("Create MCP Server Failed", err.Error())
		return
	}

	configValue := plan.Config
	plan.ID = types.StringValue(server.ID)
	plan.Title = optionalString(server.Title)
	plan.Description = optionalString(server.Description)
	applyMCPServerConfigToModel(&plan, server.Config)
	plan.Config, diags = configStateValue(configValue, server.Config)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mcpServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing MCP servers.")
		return
	}

	var state mcpServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.GetMCPServer(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read MCP Server Failed", err.Error())
		return
	}

	configValue := state.Config
	state.ID = types.StringValue(server.ID)
	state.Title = optionalString(server.Title)
	state.Description = optionalString(server.Description)
	applyMCPServerConfigToModel(&state, server.Config)
	var diags diag.Diagnostics
	state.Config, diags = configStateValue(configValue, server.Config)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *mcpServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing MCP servers.")
		return
	}

	var plan mcpServerModel
	var state mcpServerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, sendConfig, diags := mcpServerConfigForUpdate(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	update := teamapi.MCPServerUpdate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
	}
	if sendConfig {
		update.Config = &config
	}

	server, err := r.client.UpdateMCPServer(ctx, state.ID.ValueString(), update)
	if err != nil {
		resp.Diagnostics.AddError("Update MCP Server Failed", err.Error())
		return
	}

	configValue := plan.Config
	plan.ID = types.StringValue(server.ID)
	plan.Title = optionalString(server.Title)
	plan.Description = optionalString(server.Description)
	applyMCPServerConfigToModel(&plan, server.Config)
	plan.Config, diags = configStateValue(configValue, server.Config)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mcpServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing MCP servers.")
		return
	}

	var state mcpServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteMCPServer(ctx, state.ID.ValueString()); err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			return
		}
		resp.Diagnostics.AddError("Delete MCP Server Failed", err.Error())
	}
}

func (r *mcpServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mcpServerConfigFromPlan(plan mcpServerModel) (teamapi.MCPServerConfig, diag.Diagnostics) {
	if !plan.Config.IsNull() && !plan.Config.IsUnknown() {
		return mcpServerConfigFromString(plan.Config)
	}
	return mcpServerConfigFromFields(plan)
}

func mcpServerConfigForUpdate(plan mcpServerModel) (teamapi.MCPServerConfig, bool, diag.Diagnostics) {
	if !plan.Config.IsNull() && !plan.Config.IsUnknown() {
		config, diags := mcpServerConfigFromString(plan.Config)
		return config, true, diags
	}

	if !mcpServerHasTypedConfig(plan) {
		return teamapi.MCPServerConfig{}, false, nil
	}

	config, diags := mcpServerConfigFromFields(plan)
	return config, true, diags
}

func mcpServerHasTypedConfig(plan mcpServerModel) bool {
	if !(plan.Namespace.IsNull() || plan.Namespace.IsUnknown()) ||
		!(plan.Command.IsNull() || plan.Command.IsUnknown()) ||
		!(plan.Workdir.IsNull() || plan.Workdir.IsUnknown()) ||
		!(plan.RequestTimeoutMs.IsNull() || plan.RequestTimeoutMs.IsUnknown()) ||
		!(plan.StartupTimeoutMs.IsNull() || plan.StartupTimeoutMs.IsUnknown()) ||
		!(plan.HeartbeatIntervalMs.IsNull() || plan.HeartbeatIntervalMs.IsUnknown()) ||
		!(plan.StaleTimeoutMs.IsNull() || plan.StaleTimeoutMs.IsUnknown()) {
		return true
	}
	if plan.Restart != nil {
		if !(plan.Restart.MaxAttempts.IsNull() || plan.Restart.MaxAttempts.IsUnknown()) ||
			!(plan.Restart.BackoffMs.IsNull() || plan.Restart.BackoffMs.IsUnknown()) {
			return true
		}
	}
	return len(plan.Env) > 0
}

func mcpServerConfigFromFields(plan mcpServerModel) (teamapi.MCPServerConfig, diag.Diagnostics) {
	envVars, diags := envVarsFromModels(plan.Env)
	if diags.HasError() {
		return teamapi.MCPServerConfig{}, diags
	}
	return teamapi.MCPServerConfig{
		Namespace:           stringPointer(plan.Namespace),
		Command:             stringPointer(plan.Command),
		Workdir:             stringPointer(plan.Workdir),
		RequestTimeoutMs:    int64Pointer(plan.RequestTimeoutMs),
		StartupTimeoutMs:    int64Pointer(plan.StartupTimeoutMs),
		HeartbeatIntervalMs: int64Pointer(plan.HeartbeatIntervalMs),
		StaleTimeoutMs:      int64Pointer(plan.StaleTimeoutMs),
		Restart:             mcpServerRestartFromModel(plan.Restart),
		Env:                 envVars,
	}, diags
}

func mcpServerConfigFromString(value types.String) (teamapi.MCPServerConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		diags.AddAttributeError(path.Root("config"), "Missing Config", "MCP server config must be provided and cannot be empty.")
		return teamapi.MCPServerConfig{}, diags
	}
	configValue := value.ValueString()
	if configValue == "" {
		diags.AddAttributeError(path.Root("config"), "Missing Config", "MCP server config must be provided and cannot be empty.")
		return teamapi.MCPServerConfig{}, diags
	}
	var config teamapi.MCPServerConfig
	if err := json.Unmarshal([]byte(configValue), &config); err != nil {
		diags.AddAttributeError(path.Root("config"), "Invalid JSON", "MCP server config must be valid JSON.")
		return teamapi.MCPServerConfig{}, diags
	}
	return config, diags
}

func applyMCPServerConfigToModel(model *mcpServerModel, config teamapi.MCPServerConfig) {
	model.Namespace = optionalString(config.Namespace)
	model.Command = optionalString(config.Command)
	model.Workdir = optionalString(config.Workdir)
	model.RequestTimeoutMs = optionalInt64(config.RequestTimeoutMs)
	model.StartupTimeoutMs = optionalInt64(config.StartupTimeoutMs)
	model.HeartbeatIntervalMs = optionalInt64(config.HeartbeatIntervalMs)
	model.StaleTimeoutMs = optionalInt64(config.StaleTimeoutMs)
	model.Restart = mcpServerRestartModelFromAPI(config.Restart)
	model.Env = envVarModelsFromAPI(config.Env)
}

func mcpServerRestartFromModel(model *mcpServerRestartModel) *teamapi.RestartPolicy {
	if model == nil {
		return nil
	}
	maxAttempts := int64Pointer(model.MaxAttempts)
	backoffMs := int64Pointer(model.BackoffMs)
	if maxAttempts == nil && backoffMs == nil {
		return nil
	}
	return &teamapi.RestartPolicy{
		MaxAttempts: maxAttempts,
		BackoffMs:   backoffMs,
	}
}

func mcpServerRestartModelFromAPI(restart *teamapi.RestartPolicy) *mcpServerRestartModel {
	if restart == nil {
		return nil
	}
	return &mcpServerRestartModel{
		MaxAttempts: optionalInt64(restart.MaxAttempts),
		BackoffMs:   optionalInt64(restart.BackoffMs),
	}
}
