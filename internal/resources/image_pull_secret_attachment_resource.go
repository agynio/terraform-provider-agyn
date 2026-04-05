package resources

import (
	"context"
	"fmt"

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

type imagePullSecretAttachmentResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &imagePullSecretAttachmentResource{}
var _ resource.ResourceWithImportState = &imagePullSecretAttachmentResource{}

type imagePullSecretAttachmentModel struct {
	ID                types.String `tfsdk:"id"`
	ImagePullSecretID types.String `tfsdk:"image_pull_secret_id"`
	AgentID           types.String `tfsdk:"agent_id"`
	McpID             types.String `tfsdk:"mcp_id"`
	HookID            types.String `tfsdk:"hook_id"`
}

func NewImagePullSecretAttachmentResource() resource.Resource {
	return &imagePullSecretAttachmentResource{}
}

func (r *imagePullSecretAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_pull_secret_attachment"
}

func (r *imagePullSecretAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	ownerValidators := []validator.String{
		stringvalidator.ExactlyOneOf(
			path.MatchRoot("agent_id"),
			path.MatchRoot("mcp_id"),
			path.MatchRoot("hook_id"),
		),
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn image pull secret attachment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the image pull secret attachment.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"image_pull_secret_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Image pull secret identifier.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
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

func (r *imagePullSecretAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *imagePullSecretAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan imagePullSecretAttachmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &agentsv1.CreateImagePullSecretAttachmentRequest{ImagePullSecretId: plan.ImagePullSecretID.ValueString()}
	if setImagePullSecretAttachmentTarget(input, plan.AgentID, plan.McpID, plan.HookID, resp) {
		return
	}

	attachment, err := r.client.CreateImagePullSecretAttachment(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create image pull secret attachment", err.Error())
		return
	}

	agentID, mcpID, hookID := imagePullSecretAttachmentTargetState(attachment)
	state := imagePullSecretAttachmentModel{
		ID:                types.StringValue(attachment.Meta.Id),
		ImagePullSecretID: types.StringValue(attachment.ImagePullSecretId),
		AgentID:           agentID,
		McpID:             mcpID,
		HookID:            hookID,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *imagePullSecretAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state imagePullSecretAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attachment, err := r.client.GetImagePullSecretAttachment(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read image pull secret attachment", err.Error())
		return
	}

	agentID, mcpID, hookID := imagePullSecretAttachmentTargetState(attachment)
	state.ImagePullSecretID = types.StringValue(attachment.ImagePullSecretId)
	state.AgentID = agentID
	state.McpID = mcpID
	state.HookID = hookID

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *imagePullSecretAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	resp.Diagnostics.AddError(
		"Update not supported",
		"Image pull secret attachments are immutable. This is an internal error.",
	)
}

func (r *imagePullSecretAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state imagePullSecretAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteImagePullSecretAttachment(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete image pull secret attachment", err.Error())
		return
	}
}

func (r *imagePullSecretAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func setImagePullSecretAttachmentTarget(req *agentsv1.CreateImagePullSecretAttachmentRequest, agentID types.String, mcpID types.String, hookID types.String, resp *resource.CreateResponse) bool {
	if !agentID.IsNull() && !agentID.IsUnknown() {
		req.Target = &agentsv1.CreateImagePullSecretAttachmentRequest_AgentId{AgentId: agentID.ValueString()}
		return false
	}
	if !mcpID.IsNull() && !mcpID.IsUnknown() {
		req.Target = &agentsv1.CreateImagePullSecretAttachmentRequest_McpId{McpId: mcpID.ValueString()}
		return false
	}
	if !hookID.IsNull() && !hookID.IsUnknown() {
		req.Target = &agentsv1.CreateImagePullSecretAttachmentRequest_HookId{HookId: hookID.ValueString()}
		return false
	}
	resp.Diagnostics.AddError("Missing attachment target", "image pull secret attachment requires one of agent_id, mcp_id, or hook_id")
	return true
}

func imagePullSecretAttachmentTargetState(attachment *agentsv1.ImagePullSecretAttachment) (types.String, types.String, types.String) {
	agentID := types.StringNull()
	mcpID := types.StringNull()
	hookID := types.StringNull()
	switch target := attachment.GetTarget().(type) {
	case *agentsv1.ImagePullSecretAttachment_AgentId:
		agentID = types.StringValue(target.AgentId)
	case *agentsv1.ImagePullSecretAttachment_McpId:
		mcpID = types.StringValue(target.McpId)
	case *agentsv1.ImagePullSecretAttachment_HookId:
		hookID = types.StringValue(target.HookId)
	default:
		panic(fmt.Sprintf("unexpected image pull secret attachment target type: %T", target))
	}
	return agentID, mcpID, hookID
}
