package resources

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type volumeResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &volumeResource{}
var _ resource.ResourceWithImportState = &volumeResource{}

type volumeModel struct {
	ID          types.String `tfsdk:"id"`
	Persistent  types.Bool   `tfsdk:"persistent"`
	MountPath   types.String `tfsdk:"mount_path"`
	Size        types.String `tfsdk:"size"`
	Description types.String `tfsdk:"description"`
}

func NewVolumeResource() resource.Resource { return &volumeResource{} }

func (r *volumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *volumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn volume.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the volume.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"persistent": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether the volume is persistent.",
			},
			"mount_path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Mount path inside the container.",
			},
			"size": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Volume size (required when persistent is true).",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable description.",
			},
		},
	}
}

func (r *volumeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *volumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan volumeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := agentapi.VolumeCreate{
		Persistent:  plan.Persistent.ValueBool(),
		MountPath:   plan.MountPath.ValueString(),
		Size:        stringPointer(plan.Size),
		Description: stringPointer(plan.Description),
	}

	volume, err := r.client.CreateVolume(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create volume", err.Error())
		return
	}

	updatedState := volumeModel{
		ID:          types.StringValue(volume.ID),
		Persistent:  types.BoolValue(volume.Persistent),
		MountPath:   types.StringValue(volume.MountPath),
		Size:        optionalString(volume.Size),
		Description: optionalString(volume.Description),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *volumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state volumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volume, err := r.client.GetVolume(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *agentapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read volume", err.Error())
		return
	}

	state.Persistent = types.BoolValue(volume.Persistent)
	state.MountPath = types.StringValue(volume.MountPath)
	state.Size = optionalString(volume.Size)
	state.Description = optionalString(volume.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *volumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan volumeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state volumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := agentapi.VolumeUpdate{
		Persistent:  boolPointer(plan.Persistent),
		MountPath:   stringPointer(plan.MountPath),
		Size:        updateStringPointer(plan.Size, state.Size),
		Description: updateStringPointer(plan.Description, state.Description),
	}

	volume, err := r.client.UpdateVolume(ctx, plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update volume", err.Error())
		return
	}

	updatedState := volumeModel{
		ID:          types.StringValue(volume.ID),
		Persistent:  types.BoolValue(volume.Persistent),
		MountPath:   types.StringValue(volume.MountPath),
		Size:        optionalString(volume.Size),
		Description: optionalString(volume.Description),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *volumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state volumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteVolume(ctx, state.ID.ValueString()); err != nil {
		var apiErr *agentapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			return
		}
		resp.Diagnostics.AddError("Unable to delete volume", err.Error())
		return
	}
}

func (r *volumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
