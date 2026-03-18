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

type volumeAttachmentResource struct {
	client *teamapi.Client
}

var _ resource.Resource = &volumeAttachmentResource{}
var _ resource.ResourceWithImportState = &volumeAttachmentResource{}

type volumeAttachmentModel struct {
	ID       types.String `tfsdk:"id"`
	VolumeID types.String `tfsdk:"volume_id"`
	AgentID  types.String `tfsdk:"agent_id"`
	McpID    types.String `tfsdk:"mcp_id"`
	HookID   types.String `tfsdk:"hook_id"`
}

func NewVolumeAttachmentResource() resource.Resource { return &volumeAttachmentResource{} }

func (r *volumeAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_attachment"
}

func (r *volumeAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	ownerValidators := []validator.String{
		stringvalidator.ExactlyOneOf(
			path.MatchRoot("agent_id"),
			path.MatchRoot("mcp_id"),
			path.MatchRoot("hook_id"),
		),
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn volume attachment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the volume attachment.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"volume_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Volume identifier.",
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

func (r *volumeAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *volumeAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan volumeAttachmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := teamapi.VolumeAttachmentCreate{
		VolumeID: plan.VolumeID.ValueString(),
		AgentID:  stringPointer(plan.AgentID),
		McpID:    stringPointer(plan.McpID),
		HookID:   stringPointer(plan.HookID),
	}

	attachment, err := r.client.CreateVolumeAttachment(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create volume attachment", err.Error())
		return
	}

	state := volumeAttachmentModel{
		ID:       types.StringValue(attachment.ID),
		VolumeID: types.StringValue(attachment.VolumeID),
		AgentID:  optionalString(attachment.AgentID),
		McpID:    optionalString(attachment.McpID),
		HookID:   optionalString(attachment.HookID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *volumeAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state volumeAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attachment, err := r.client.GetVolumeAttachment(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read volume attachment", err.Error())
		return
	}

	state.VolumeID = types.StringValue(attachment.VolumeID)
	state.AgentID = optionalString(attachment.AgentID)
	state.McpID = optionalString(attachment.McpID)
	state.HookID = optionalString(attachment.HookID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *volumeAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state volumeAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attachment, err := r.client.GetVolumeAttachment(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read volume attachment", err.Error())
		return
	}

	state.VolumeID = types.StringValue(attachment.VolumeID)
	state.AgentID = optionalString(attachment.AgentID)
	state.McpID = optionalString(attachment.McpID)
	state.HookID = optionalString(attachment.HookID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *volumeAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state volumeAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteVolumeAttachment(ctx, state.ID.ValueString()); err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			return
		}
		resp.Diagnostics.AddError("Unable to delete volume attachment", err.Error())
		return
	}
}

func (r *volumeAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
