package resources

import (
	"context"
	"encoding/json"
	"errors"

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

type agentResource struct {
	client *teamapi.Client
}

var _ resource.ResourceWithUpgradeState = &agentResource{}

type agentModel struct {
	ID                        types.String `tfsdk:"id"`
	Title                     types.String `tfsdk:"title"`
	Description               types.String `tfsdk:"description"`
	Config                    types.String `tfsdk:"config"`
	Name                      types.String `tfsdk:"name"`
	Role                      types.String `tfsdk:"role"`
	Model                     types.String `tfsdk:"model"`
	SystemPrompt              types.String `tfsdk:"system_prompt"`
	DebounceMs                types.Int64  `tfsdk:"debounce_ms"`
	WhenBusy                  types.String `tfsdk:"when_busy"`
	ProcessBuffer             types.String `tfsdk:"process_buffer"`
	SendFinalResponseToThread types.Bool   `tfsdk:"send_final_response_to_thread"`
	RestrictOutput            types.Bool   `tfsdk:"restrict_output"`
	RestrictionMessage        types.String `tfsdk:"restriction_message"`
	RestrictionMaxInjections  types.Int64  `tfsdk:"restriction_max_injections"`
	SummarizationKeepTokens   types.Int64  `tfsdk:"summarization_keep_tokens"`
	SummarizationMaxTokens    types.Int64  `tfsdk:"summarization_max_tokens"`
}

type agentModelV0 struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Config      types.String `tfsdk:"config"`
}

func NewAgentResource() resource.Resource { return &agentResource{} }

func (r *agentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

func (r *agentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn agent.",
		Version:             1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the agent.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"title": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable title of the agent.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable description of the agent.",
			},
			"config": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Deprecated JSON-encoded agent configuration. Use structured attributes instead.",
				DeprecationMessage:  "Use structured configuration attributes instead of config.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("name"),
						path.MatchRoot("role"),
						path.MatchRoot("model"),
						path.MatchRoot("system_prompt"),
						path.MatchRoot("debounce_ms"),
						path.MatchRoot("when_busy"),
						path.MatchRoot("process_buffer"),
						path.MatchRoot("send_final_response_to_thread"),
						path.MatchRoot("restrict_output"),
						path.MatchRoot("restriction_message"),
						path.MatchRoot("restriction_max_injections"),
						path.MatchRoot("summarization_keep_tokens"),
						path.MatchRoot("summarization_max_tokens"),
					),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Agent name for the configuration.",
			},
			"role": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Role assigned to the agent.",
			},
			"model": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Model identifier override for the agent.",
			},
			"system_prompt": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "System prompt override for the agent.",
			},
			"debounce_ms": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Debounce duration in milliseconds.",
			},
			"when_busy": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Behavior when the agent is busy (injectAfterTools or wait).",
				Validators: []validator.String{
					stringvalidator.OneOf("injectAfterTools", "wait"),
				},
			},
			"process_buffer": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Processing strategy for buffered content (allTogether or oneByOne).",
				Validators: []validator.String{
					stringvalidator.OneOf("allTogether", "oneByOne"),
				},
			},
			"send_final_response_to_thread": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to send the final response to the thread.",
			},
			"restrict_output": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to enforce output restrictions.",
			},
			"restriction_message": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Message to inject when output restrictions are applied.",
			},
			"restriction_max_injections": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of restriction message injections.",
			},
			"summarization_keep_tokens": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of tokens to keep during summarization.",
			},
			"summarization_max_tokens": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of tokens for summarization.",
			},
		},
	}
}

func (r *agentResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
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
				var prior agentModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}

				config, diags := agentConfigFromString(prior.Config)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgraded := agentModel{
					ID:          prior.ID,
					Title:       prior.Title,
					Description: prior.Description,
					Config:      types.StringNull(),
				}
				applyAgentConfigToModel(&upgraded, config)

				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
	}
}

func (r *agentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *agentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing agents.")
		return
	}

	var plan agentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, diags := agentConfigFromPlan(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	create := teamapi.AgentCreate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
		Config:      config,
	}

	agent, err := r.client.CreateAgent(ctx, create)
	if err != nil {
		resp.Diagnostics.AddError("Create Agent Failed", err.Error())
		return
	}

	configValue := plan.Config
	plan.ID = types.StringValue(agent.ID)
	plan.Title = optionalString(agent.Title)
	plan.Description = optionalString(agent.Description)
	applyAgentConfigToModel(&plan, agent.Config)
	plan.Config, diags = configStateValue(configValue, agent.Config)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *agentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing agents.")
		return
	}

	var state agentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	agent, err := r.client.GetAgent(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Agent Failed", err.Error())
		return
	}

	configValue := state.Config
	state.ID = types.StringValue(agent.ID)
	state.Title = optionalString(agent.Title)
	state.Description = optionalString(agent.Description)
	applyAgentConfigToModel(&state, agent.Config)
	var diags diag.Diagnostics
	state.Config, diags = configStateValue(configValue, agent.Config)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *agentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing agents.")
		return
	}

	var plan agentModel
	var state agentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, sendConfig, diags := agentConfigForUpdate(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	update := teamapi.AgentUpdate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
	}
	if sendConfig {
		update.Config = &config
	}

	agent, err := r.client.UpdateAgent(ctx, state.ID.ValueString(), update)
	if err != nil {
		resp.Diagnostics.AddError("Update Agent Failed", err.Error())
		return
	}

	configValue := plan.Config
	plan.ID = types.StringValue(agent.ID)
	plan.Title = optionalString(agent.Title)
	plan.Description = optionalString(agent.Description)
	applyAgentConfigToModel(&plan, agent.Config)
	plan.Config, diags = configStateValue(configValue, agent.Config)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *agentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing agents.")
		return
	}

	var state agentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAgent(ctx, state.ID.ValueString()); err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			return
		}
		resp.Diagnostics.AddError("Delete Agent Failed", err.Error())
	}
}

func (r *agentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func agentConfigFromPlan(plan agentModel) (teamapi.AgentConfig, diag.Diagnostics) {
	if !plan.Config.IsNull() && !plan.Config.IsUnknown() {
		return agentConfigFromString(plan.Config)
	}
	return agentConfigFromFields(plan), nil
}

func agentConfigForUpdate(plan agentModel) (teamapi.AgentConfig, bool, diag.Diagnostics) {
	if !plan.Config.IsNull() && !plan.Config.IsUnknown() {
		config, diags := agentConfigFromString(plan.Config)
		return config, true, diags
	}

	if !agentHasTypedConfig(plan) {
		return teamapi.AgentConfig{}, false, nil
	}

	return agentConfigFromFields(plan), true, nil
}

func agentHasTypedConfig(plan agentModel) bool {
	return !(plan.Name.IsNull() || plan.Name.IsUnknown()) ||
		!(plan.Role.IsNull() || plan.Role.IsUnknown()) ||
		!(plan.Model.IsNull() || plan.Model.IsUnknown()) ||
		!(plan.SystemPrompt.IsNull() || plan.SystemPrompt.IsUnknown()) ||
		!(plan.DebounceMs.IsNull() || plan.DebounceMs.IsUnknown()) ||
		!(plan.WhenBusy.IsNull() || plan.WhenBusy.IsUnknown()) ||
		!(plan.ProcessBuffer.IsNull() || plan.ProcessBuffer.IsUnknown()) ||
		!(plan.SendFinalResponseToThread.IsNull() || plan.SendFinalResponseToThread.IsUnknown()) ||
		!(plan.RestrictOutput.IsNull() || plan.RestrictOutput.IsUnknown()) ||
		!(plan.RestrictionMessage.IsNull() || plan.RestrictionMessage.IsUnknown()) ||
		!(plan.RestrictionMaxInjections.IsNull() || plan.RestrictionMaxInjections.IsUnknown()) ||
		!(plan.SummarizationKeepTokens.IsNull() || plan.SummarizationKeepTokens.IsUnknown()) ||
		!(plan.SummarizationMaxTokens.IsNull() || plan.SummarizationMaxTokens.IsUnknown())
}

func agentConfigFromFields(plan agentModel) teamapi.AgentConfig {
	return teamapi.AgentConfig{
		Name:                      stringPointer(plan.Name),
		Role:                      stringPointer(plan.Role),
		Model:                     stringPointer(plan.Model),
		SystemPrompt:              stringPointer(plan.SystemPrompt),
		DebounceMs:                int64Pointer(plan.DebounceMs),
		WhenBusy:                  stringPointer(plan.WhenBusy),
		ProcessBuffer:             stringPointer(plan.ProcessBuffer),
		SendFinalResponseToThread: boolPointer(plan.SendFinalResponseToThread),
		RestrictOutput:            boolPointer(plan.RestrictOutput),
		RestrictionMessage:        stringPointer(plan.RestrictionMessage),
		RestrictionMaxInjections:  int64Pointer(plan.RestrictionMaxInjections),
		SummarizationKeepTokens:   int64Pointer(plan.SummarizationKeepTokens),
		SummarizationMaxTokens:    int64Pointer(plan.SummarizationMaxTokens),
	}
}

func agentConfigFromString(value types.String) (teamapi.AgentConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		diags.AddAttributeError(path.Root("config"), "Missing Config", "Agent config must be provided and cannot be empty.")
		return teamapi.AgentConfig{}, diags
	}
	configValue := value.ValueString()
	if configValue == "" {
		diags.AddAttributeError(path.Root("config"), "Missing Config", "Agent config must be provided and cannot be empty.")
		return teamapi.AgentConfig{}, diags
	}
	var config teamapi.AgentConfig
	if err := json.Unmarshal([]byte(configValue), &config); err != nil {
		diags.AddAttributeError(path.Root("config"), "Invalid JSON", "Agent config must be valid JSON.")
		return teamapi.AgentConfig{}, diags
	}
	return config, diags
}

func applyAgentConfigToModel(model *agentModel, config teamapi.AgentConfig) {
	model.Name = optionalString(config.Name)
	model.Role = optionalString(config.Role)
	model.Model = optionalString(config.Model)
	model.SystemPrompt = optionalString(config.SystemPrompt)
	model.DebounceMs = optionalInt64(config.DebounceMs)
	model.WhenBusy = optionalString(config.WhenBusy)
	model.ProcessBuffer = optionalString(config.ProcessBuffer)
	model.SendFinalResponseToThread = optionalBool(config.SendFinalResponseToThread)
	model.RestrictOutput = optionalBool(config.RestrictOutput)
	model.RestrictionMessage = optionalString(config.RestrictionMessage)
	model.RestrictionMaxInjections = optionalInt64(config.RestrictionMaxInjections)
	model.SummarizationKeepTokens = optionalInt64(config.SummarizationKeepTokens)
	model.SummarizationMaxTokens = optionalInt64(config.SummarizationMaxTokens)
}
