package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type initScriptResource struct {
	client *agentapi.Client
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
	client, ok := req.ProviderData.(*agentapi.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *agentapi.Client")
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

	input := &agentsv1.CreateInitScriptRequest{
		Script:      plan.Script.ValueString(),
		Description: stringValue(plan.Description),
	}
	if setInitScriptTarget(input, plan.AgentID, plan.McpID, plan.HookID, "create init script", resp) {
		return
	}

	script, err := r.client.CreateInitScript(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create init script", err.Error())
		return
	}

	agentID, mcpID, hookID := initScriptTargetState(script)
	updatedState := initScriptModel{
		ID:          types.StringValue(script.Meta.Id),
		Script:      types.StringValue(script.Script),
		Description: optionalString(script.Description),
		AgentID:     agentID,
		McpID:       mcpID,
		HookID:      hookID,
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
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read init script", err.Error())
		return
	}

	agentID, mcpID, hookID := initScriptTargetState(script)
	state.Script = types.StringValue(script.Script)
	state.Description = optionalString(script.Description)
	state.AgentID = agentID
	state.McpID = mcpID
	state.HookID = hookID

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

	input := &agentsv1.UpdateInitScriptRequest{
		Id:          plan.ID.ValueString(),
		Script:      updateStringPointer(plan.Script, state.Script),
		Description: updateStringPointer(plan.Description, state.Description),
	}

	script, err := r.client.UpdateInitScript(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update init script", err.Error())
		return
	}

	agentID, mcpID, hookID := initScriptTargetState(script)
	updatedState := initScriptModel{
		ID:          types.StringValue(script.Meta.Id),
		Script:      types.StringValue(script.Script),
		Description: optionalString(script.Description),
		AgentID:     agentID,
		McpID:       mcpID,
		HookID:      hookID,
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
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete init script", err.Error())
		return
	}
}

func (r *initScriptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func setInitScriptTarget(req *agentsv1.CreateInitScriptRequest, agentID types.String, mcpID types.String, hookID types.String, op string, resp *resource.CreateResponse) bool {
	if !agentID.IsNull() && !agentID.IsUnknown() {
		req.Target = &agentsv1.CreateInitScriptRequest_AgentId{AgentId: agentID.ValueString()}
		return false
	}
	if !mcpID.IsNull() && !mcpID.IsUnknown() {
		req.Target = &agentsv1.CreateInitScriptRequest_McpId{McpId: mcpID.ValueString()}
		return false
	}
	if !hookID.IsNull() && !hookID.IsUnknown() {
		req.Target = &agentsv1.CreateInitScriptRequest_HookId{HookId: hookID.ValueString()}
		return false
	}
	resp.Diagnostics.AddError("Missing init script target", op+" requires one of agent_id, mcp_id, or hook_id")
	return true
}

func initScriptTargetState(script *agentsv1.InitScript) (types.String, types.String, types.String) {
	agentID := types.StringNull()
	mcpID := types.StringNull()
	hookID := types.StringNull()
	switch target := script.GetTarget().(type) {
	case *agentsv1.InitScript_AgentId:
		agentID = types.StringValue(target.AgentId)
	case *agentsv1.InitScript_McpId:
		mcpID = types.StringValue(target.McpId)
	case *agentsv1.InitScript_HookId:
		hookID = types.StringValue(target.HookId)
	}
	return agentID, mcpID, hookID
}
