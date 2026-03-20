package resources

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type envResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &envResource{}
var _ resource.ResourceWithImportState = &envResource{}

type envModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	AgentID     types.String `tfsdk:"agent_id"`
	McpID       types.String `tfsdk:"mcp_id"`
	HookID      types.String `tfsdk:"hook_id"`
	Value       types.String `tfsdk:"value"`
	SecretID    types.String `tfsdk:"secret_id"`
}

func NewEnvResource() resource.Resource { return &envResource{} }

func (r *envResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env"
}

func (r *envResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	ownerValidators := []validator.String{
		stringvalidator.ExactlyOneOf(
			path.MatchRoot("agent_id"),
			path.MatchRoot("mcp_id"),
			path.MatchRoot("hook_id"),
		),
	}
	valueValidators := []validator.String{
		stringvalidator.ExactlyOneOf(
			path.MatchRoot("value"),
			path.MatchRoot("secret_id"),
		),
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn environment variable.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the environment variable.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Environment variable name.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable description.",
			},
			"agent_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Target agent identifier.",
				Validators:          ownerValidators,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"mcp_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Target MCP identifier.",
				Validators:          ownerValidators,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"hook_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Target hook identifier.",
				Validators:          ownerValidators,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"value": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Plain-text value.",
				Validators:          valueValidators,
			},
			"secret_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Secret reference identifier.",
				Validators:          valueValidators,
			},
		},
	}
}

func (r *envResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*agentapi.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *agentapi.Client")
		return
	}
	r.client = client
}

func (r *envResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan envModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := agentapi.EnvCreate{
		Name:        plan.Name.ValueString(),
		Description: stringPointer(plan.Description),
		AgentID:     stringPointer(plan.AgentID),
		McpID:       stringPointer(plan.McpID),
		HookID:      stringPointer(plan.HookID),
		Value:       stringPointer(plan.Value),
		SecretID:    stringPointer(plan.SecretID),
	}

	env, err := r.client.CreateEnv(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create env", err.Error())
		return
	}

	updatedState := envModel{
		ID:          types.StringValue(env.ID),
		Name:        types.StringValue(env.Name),
		Description: optionalString(env.Description),
		AgentID:     optionalString(env.AgentID),
		McpID:       optionalString(env.McpID),
		HookID:      optionalString(env.HookID),
		Value:       preserveSensitiveString(plan.Value, env.Value),
		SecretID:    optionalString(env.SecretID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *envResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state envModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := r.client.GetEnv(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *agentapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read env", err.Error())
		return
	}

	state.Name = types.StringValue(env.Name)
	state.Description = optionalString(env.Description)
	state.AgentID = optionalString(env.AgentID)
	state.McpID = optionalString(env.McpID)
	state.HookID = optionalString(env.HookID)
	state.Value = preserveSensitiveString(state.Value, env.Value)
	state.SecretID = optionalString(env.SecretID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *envResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan envModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state envModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := agentapi.EnvUpdate{
		Name:        stringPointer(plan.Name),
		Description: updateStringPointer(plan.Description, state.Description),
		Value:       updateStringPointer(plan.Value, state.Value),
		SecretID:    updateStringPointer(plan.SecretID, state.SecretID),
	}

	env, err := r.client.UpdateEnv(ctx, plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update env", err.Error())
		return
	}

	updatedState := envModel{
		ID:          types.StringValue(env.ID),
		Name:        types.StringValue(env.Name),
		Description: optionalString(env.Description),
		AgentID:     optionalString(env.AgentID),
		McpID:       optionalString(env.McpID),
		HookID:      optionalString(env.HookID),
		Value:       preserveSensitiveString(plan.Value, env.Value),
		SecretID:    optionalString(env.SecretID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *envResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state envModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteEnv(ctx, state.ID.ValueString()); err != nil {
		var apiErr *agentapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			return
		}
		resp.Diagnostics.AddError("Unable to delete env", err.Error())
		return
	}
}

func (r *envResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
