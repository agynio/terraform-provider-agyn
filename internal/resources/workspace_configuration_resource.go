package resources

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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

type workspaceConfigurationResource struct {
	client *teamapi.Client
}

var _ resource.ResourceWithUpgradeState = &workspaceConfigurationResource{}

type workspaceVolumesModel struct {
	Enabled   types.Bool   `tfsdk:"enabled"`
	MountPath types.String `tfsdk:"mount_path"`
}

type workspaceConfigurationModel struct {
	ID            types.String           `tfsdk:"id"`
	Title         types.String           `tfsdk:"title"`
	Description   types.String           `tfsdk:"description"`
	Config        types.String           `tfsdk:"config"`
	Image         types.String           `tfsdk:"image"`
	Env           []envVarModel          `tfsdk:"env"`
	InitialScript types.String           `tfsdk:"initial_script"`
	CpuLimit      types.String           `tfsdk:"cpu_limit"`
	MemoryLimit   types.String           `tfsdk:"memory_limit"`
	Platform      types.String           `tfsdk:"platform"`
	EnableDinD    types.Bool             `tfsdk:"enable_dind"`
	TtlSeconds    types.Int64            `tfsdk:"ttl_seconds"`
	Volumes       *workspaceVolumesModel `tfsdk:"volumes"`
	Nix           jsontypes.Normalized   `tfsdk:"nix"`
}

type workspaceConfigurationModelV0 struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Config      types.String `tfsdk:"config"`
}

func NewWorkspaceConfigurationResource() resource.Resource { return &workspaceConfigurationResource{} }

func (r *workspaceConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_configuration"
}

func (r *workspaceConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn workspace configuration.",
		Version:             1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the workspace configuration.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"title": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable title of the workspace configuration.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable description of the workspace configuration.",
			},
			"config": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Deprecated JSON-encoded workspace configuration. Use structured attributes instead.",
				DeprecationMessage:  "Use structured configuration attributes instead of config.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("image"),
						path.MatchRoot("env"),
						path.MatchRoot("initial_script"),
						path.MatchRoot("cpu_limit"),
						path.MatchRoot("memory_limit"),
						path.MatchRoot("platform"),
						path.MatchRoot("enable_dind"),
						path.MatchRoot("ttl_seconds"),
						path.MatchRoot("volumes"),
						path.MatchRoot("nix"),
					),
				},
			},
			"image": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Container image for the workspace.",
			},
			"env": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Environment variables for the workspace.",
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
			"initial_script": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Initial script to run in the workspace.",
			},
			"cpu_limit": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "CPU limit for the workspace (string or number as a string).",
			},
			"memory_limit": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Memory limit for the workspace (string or number as a string).",
			},
			"platform": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Platform selection (auto, linux/amd64, linux/arm64).",
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "linux/amd64", "linux/arm64"),
				},
			},
			"enable_dind": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to enable Docker-in-Docker.",
			},
			"ttl_seconds": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Time-to-live in seconds for the workspace.",
			},
			"volumes": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Workspace volume configuration.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether volumes are enabled.",
					},
					"mount_path": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Mount path for the volume.",
					},
				},
			},
			"nix": schema.StringAttribute{
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
				MarkdownDescription: "Nix configuration as JSON.",
			},
		},
	}
}

func (r *workspaceConfigurationResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
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
				var prior workspaceConfigurationModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}

				config, diags := workspaceConfigFromString(prior.Config)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgraded := workspaceConfigurationModel{
					ID:          prior.ID,
					Title:       prior.Title,
					Description: prior.Description,
					Config:      types.StringNull(),
					Nix:         jsontypes.NewNormalizedNull(),
				}
				resp.Diagnostics.Append(applyWorkspaceConfigToModel(&upgraded, config)...)
				if resp.Diagnostics.HasError() {
					return
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
	}
}

func (r *workspaceConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *workspaceConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing workspace configurations.")
		return
	}

	var plan workspaceConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, diags := workspaceConfigFromPlan(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	create := teamapi.WorkspaceConfigurationCreate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
		Config:      config,
	}

	cfg, err := r.client.CreateWorkspaceConfiguration(ctx, create)
	if err != nil {
		resp.Diagnostics.AddError("Create Workspace Configuration Failed", err.Error())
		return
	}

	configValue := plan.Config
	plan.ID = types.StringValue(cfg.ID)
	plan.Title = optionalString(cfg.Title)
	plan.Description = optionalString(cfg.Description)
	if configValue.IsNull() || configValue.IsUnknown() {
		resp.Diagnostics.Append(applyWorkspaceConfigToModel(&plan, cfg.Config)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	plan.Config, diags = configStateValue(configValue, cfg.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workspaceConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing workspace configurations.")
		return
	}

	var state workspaceConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetWorkspaceConfiguration(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Workspace Configuration Failed", err.Error())
		return
	}

	configValue := state.Config
	state.ID = types.StringValue(cfg.ID)
	state.Title = optionalString(cfg.Title)
	state.Description = optionalString(cfg.Description)
	if configValue.IsNull() || configValue.IsUnknown() {
		resp.Diagnostics.Append(applyWorkspaceConfigToModel(&state, cfg.Config)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	var diags diag.Diagnostics
	state.Config, diags = configStateValue(configValue, cfg.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workspaceConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing workspace configurations.")
		return
	}

	var plan workspaceConfigurationModel
	var state workspaceConfigurationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, sendConfig, diags := workspaceConfigForUpdate(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	update := teamapi.WorkspaceConfigurationUpdate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
	}
	if sendConfig {
		update.Config = &config
	}

	cfg, err := r.client.UpdateWorkspaceConfiguration(ctx, state.ID.ValueString(), update)
	if err != nil {
		resp.Diagnostics.AddError("Update Workspace Configuration Failed", err.Error())
		return
	}

	configValue := plan.Config
	plan.ID = types.StringValue(cfg.ID)
	plan.Title = optionalString(cfg.Title)
	plan.Description = optionalString(cfg.Description)
	if configValue.IsNull() || configValue.IsUnknown() {
		resp.Diagnostics.Append(applyWorkspaceConfigToModel(&plan, cfg.Config)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	plan.Config, diags = configStateValue(configValue, cfg.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workspaceConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing workspace configurations.")
		return
	}

	var state workspaceConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteWorkspaceConfiguration(ctx, state.ID.ValueString()); err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			return
		}
		resp.Diagnostics.AddError("Delete Workspace Configuration Failed", err.Error())
	}
}

func (r *workspaceConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func workspaceConfigFromPlan(plan workspaceConfigurationModel) (teamapi.WorkspaceConfigurationConfig, diag.Diagnostics) {
	if !plan.Config.IsNull() && !plan.Config.IsUnknown() {
		return workspaceConfigFromString(plan.Config)
	}
	return workspaceConfigFromFields(plan)
}

func workspaceConfigForUpdate(plan workspaceConfigurationModel) (teamapi.WorkspaceConfigurationConfig, bool, diag.Diagnostics) {
	if !plan.Config.IsNull() && !plan.Config.IsUnknown() {
		config, diags := workspaceConfigFromString(plan.Config)
		return config, true, diags
	}

	if !workspaceHasTypedConfig(plan) {
		return teamapi.WorkspaceConfigurationConfig{}, false, nil
	}

	config, diags := workspaceConfigFromFields(plan)
	return config, true, diags
}

func workspaceHasTypedConfig(plan workspaceConfigurationModel) bool {
	if !(plan.Image.IsNull() || plan.Image.IsUnknown()) ||
		!(plan.InitialScript.IsNull() || plan.InitialScript.IsUnknown()) ||
		!(plan.CpuLimit.IsNull() || plan.CpuLimit.IsUnknown()) ||
		!(plan.MemoryLimit.IsNull() || plan.MemoryLimit.IsUnknown()) ||
		!(plan.Platform.IsNull() || plan.Platform.IsUnknown()) ||
		!(plan.EnableDinD.IsNull() || plan.EnableDinD.IsUnknown()) ||
		!(plan.TtlSeconds.IsNull() || plan.TtlSeconds.IsUnknown()) ||
		!(plan.Nix.IsNull() || plan.Nix.IsUnknown()) {
		return true
	}
	if plan.Volumes != nil {
		if !(plan.Volumes.Enabled.IsNull() || plan.Volumes.Enabled.IsUnknown()) ||
			!(plan.Volumes.MountPath.IsNull() || plan.Volumes.MountPath.IsUnknown()) {
			return true
		}
	}
	return len(plan.Env) > 0
}

func workspaceConfigFromFields(plan workspaceConfigurationModel) (teamapi.WorkspaceConfigurationConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	envVars, envDiags := envVarsFromModels(plan.Env)
	diags.Append(envDiags...)
	if diags.HasError() {
		return teamapi.WorkspaceConfigurationConfig{}, diags
	}

	cpuLimit, cpuDiags := rawMessageFromString(plan.CpuLimit, "cpu_limit")
	diags.Append(cpuDiags...)
	memoryLimit, memDiags := rawMessageFromString(plan.MemoryLimit, "memory_limit")
	diags.Append(memDiags...)
	nixValue, nixDiags := rawMessageFromNormalized(plan.Nix)
	diags.Append(nixDiags...)
	if diags.HasError() {
		return teamapi.WorkspaceConfigurationConfig{}, diags
	}

	return teamapi.WorkspaceConfigurationConfig{
		Image:         stringPointer(plan.Image),
		Env:           envVars,
		InitialScript: stringPointer(plan.InitialScript),
		CpuLimit:      cpuLimit,
		MemoryLimit:   memoryLimit,
		Platform:      stringPointer(plan.Platform),
		EnableDinD:    boolPointer(plan.EnableDinD),
		TtlSeconds:    int64Pointer(plan.TtlSeconds),
		Volumes:       workspaceVolumesFromModel(plan.Volumes),
		Nix:           nixValue,
	}, diags
}

func workspaceConfigFromString(value types.String) (teamapi.WorkspaceConfigurationConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		diags.AddAttributeError(path.Root("config"), "Missing Config", "Workspace configuration config must be provided and cannot be empty.")
		return teamapi.WorkspaceConfigurationConfig{}, diags
	}
	configValue := value.ValueString()
	if configValue == "" {
		diags.AddAttributeError(path.Root("config"), "Missing Config", "Workspace configuration config must be provided and cannot be empty.")
		return teamapi.WorkspaceConfigurationConfig{}, diags
	}
	var config teamapi.WorkspaceConfigurationConfig
	if err := json.Unmarshal([]byte(configValue), &config); err != nil {
		diags.AddAttributeError(path.Root("config"), "Invalid JSON", "Workspace configuration config must be valid JSON.")
		return teamapi.WorkspaceConfigurationConfig{}, diags
	}
	return config, diags
}

func applyWorkspaceConfigToModel(model *workspaceConfigurationModel, config teamapi.WorkspaceConfigurationConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	model.Image = optionalString(config.Image)
	model.Env = envVarModelsFromAPI(config.Env)
	model.InitialScript = optionalString(config.InitialScript)
	cpuLimit, cpuDiags := stringFromRawMessage(config.CpuLimit, "cpu_limit")
	diags.Append(cpuDiags...)
	model.CpuLimit = cpuLimit
	memoryLimit, memDiags := stringFromRawMessage(config.MemoryLimit, "memory_limit")
	diags.Append(memDiags...)
	model.MemoryLimit = memoryLimit
	model.Platform = optionalString(config.Platform)
	model.EnableDinD = optionalBool(config.EnableDinD)
	model.TtlSeconds = optionalInt64(config.TtlSeconds)
	model.Volumes = workspaceVolumesModelFromAPI(config.Volumes)
	model.Nix = normalizedFromRawMessage(config.Nix)
	return diags
}

func workspaceVolumesFromModel(model *workspaceVolumesModel) *teamapi.WorkspaceVolumes {
	if model == nil {
		return nil
	}
	enabled := boolPointer(model.Enabled)
	mountPath := stringPointer(model.MountPath)
	if enabled == nil && mountPath == nil {
		return nil
	}
	return &teamapi.WorkspaceVolumes{
		Enabled:   enabled,
		MountPath: mountPath,
	}
}

func workspaceVolumesModelFromAPI(volumes *teamapi.WorkspaceVolumes) *workspaceVolumesModel {
	if volumes == nil {
		return nil
	}
	return &workspaceVolumesModel{
		Enabled:   optionalBool(volumes.Enabled),
		MountPath: optionalString(volumes.MountPath),
	}
}

func rawMessageFromString(value types.String, label string) (*json.RawMessage, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}
	input := value.ValueString()
	var parsed any
	if err := json.Unmarshal([]byte(input), &parsed); err == nil {
		switch parsed.(type) {
		case string, float64:
			normalized, err := json.Marshal(parsed)
			if err != nil {
				diags.AddError("Invalid "+label, err.Error())
				return nil, diags
			}
			msg := json.RawMessage(normalized)
			return &msg, diags
		default:
			diags.AddError("Invalid "+label, "Expected string or number.")
			return nil, diags
		}
	}
	raw, err := json.Marshal(input)
	if err != nil {
		diags.AddError("Invalid "+label, err.Error())
		return nil, diags
	}
	msg := json.RawMessage(raw)
	return &msg, diags
}

func stringFromRawMessage(raw *json.RawMessage, label string) (types.String, diag.Diagnostics) {
	var diags diag.Diagnostics
	if raw == nil {
		return types.StringNull(), diags
	}
	dec := json.NewDecoder(bytes.NewReader(*raw))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		diags.AddError("Invalid "+label, err.Error())
		return types.StringNull(), diags
	}
	switch typed := value.(type) {
	case string:
		return types.StringValue(typed), diags
	case json.Number:
		return types.StringValue(typed.String()), diags
	case nil:
		return types.StringNull(), diags
	default:
		diags.AddError("Invalid "+label, "Expected string or number.")
		return types.StringNull(), diags
	}
}

func rawMessageFromNormalized(value jsontypes.Normalized) (*json.RawMessage, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}
	raw := json.RawMessage(value.ValueString())
	return &raw, diags
}

func normalizedFromRawMessage(raw *json.RawMessage) jsontypes.Normalized {
	if raw == nil || len(*raw) == 0 {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(*raw))
}
