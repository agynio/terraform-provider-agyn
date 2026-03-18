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

	"github.com/agynio/terraform-provider-agyn/internal/teamapi"
)

type initScriptResource struct {
	client *teamapi.Client
}

var _ resource.Resource = &initScriptResource{}
var _ resource.ResourceWithImportState = &initScriptResource{}

type initScriptModel struct {
	ID          types.String `tfsdk:"id"`
	Script      types.String `tfsdk:"script"`
	Description types.String `tfsdk:"description"`
	AgentID     types.String `tfsdk:"agent_id"`
	McpID       types.String `tfsdk:"mcp_id"`
	HookID      types.String `tfsdk:"hook_id"`
}

func NewInitScriptResource() resource.Resource { return &initScriptResource{} }

func (r *initScriptResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_init_script"
}

func (r *initScriptResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	ownerValidators := []validator.String{
		stringvalidator.ExactlyOneOf(
			path.MatchRoot("agent_id"),
			path.MatchRoot("mcp_id"),
			path.MatchRoot("hook_id"),
		),
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn init script.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the init script.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"script": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Init script content.",
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
		},
	}
}

func (r *initScriptResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*teamapi.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *teamapi.Client")
		return
	}
	r.client = client
}

func (r *initScriptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan initScriptModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := teamapi.InitScriptCreate{
		Script:      plan.Script.ValueString(),
		Description: stringPointer(plan.Description),
		AgentID:     stringPointer(plan.AgentID),
		McpID:       stringPointer(plan.McpID),
		HookID:      stringPointer(plan.HookID),
	}

	script, err := r.client.CreateInitScript(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create init script", err.Error())
		return
	}

	updatedState := initScriptModel{
		ID:          types.StringValue(script.ID),
		Script:      types.StringValue(script.Script),
		Description: optionalString(script.Description),
		AgentID:     optionalString(script.AgentID),
		McpID:       optionalString(script.McpID),
		HookID:      optionalString(script.HookID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *initScriptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state initScriptModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	script, err := r.client.GetInitScript(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read init script", err.Error())
		return
	}

	state.Script = types.StringValue(script.Script)
	state.Description = optionalString(script.Description)
	state.AgentID = optionalString(script.AgentID)
	state.McpID = optionalString(script.McpID)
	state.HookID = optionalString(script.HookID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *initScriptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan initScriptModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state initScriptModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := teamapi.InitScriptUpdate{
		Script:      stringPointer(plan.Script),
		Description: updateStringPointer(plan.Description, state.Description),
	}

	script, err := r.client.UpdateInitScript(ctx, plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update init script", err.Error())
		return
	}

	updatedState := initScriptModel{
		ID:          types.StringValue(script.ID),
		Script:      types.StringValue(script.Script),
		Description: optionalString(script.Description),
		AgentID:     optionalString(script.AgentID),
		McpID:       optionalString(script.McpID),
		HookID:      optionalString(script.HookID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *initScriptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state initScriptModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteInitScript(ctx, state.ID.ValueString()); err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			return
		}
		resp.Diagnostics.AddError("Unable to delete init script", err.Error())
		return
	}
}

func (r *initScriptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
