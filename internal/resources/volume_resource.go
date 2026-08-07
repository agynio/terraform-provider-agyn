package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type volumeResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &volumeResource{}
var _ resource.ResourceWithImportState = &volumeResource{}

type volumeModel struct {
	ID            types.String `tfsdk:"id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	McpID         types.String `tfsdk:"mcp_id"`
	Name          types.String `tfsdk:"name"`
	MountPath     types.String `tfsdk:"mount_path"`
	Size          types.String `tfsdk:"size"`
	StorageClass  types.String `tfsdk:"storage_class"`
	TTL           types.String `tfsdk:"ttl"`
}

func NewVolumeResource() resource.Resource { return &volumeResource{} }

func (r *volumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *volumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a volume on an environment or an MCP server. A volume is a definition, not a disk: one disk is provisioned per agent instance and per sandbox that runs it.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the volume.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"environment_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Environment that mounts the volume. Conflicts with mcp_id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mcp_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "MCP server that mounts the volume. Conflicts with environment_id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Volume name, unique within its target. An MCP's shared_volumes references this.",
			},
			"mount_path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Absolute container path for the mount.",
			},
			"size": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Capacity, e.g. 10Gi. Setting it makes the volume persistent; omitting it makes it ephemeral scratch.",
			},
			"storage_class": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Storage class in the runner's catalog. Resolved at provisioning time.",
			},
			"ttl": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "How long after an owner's last workload stops before that owner's disk is deleted.",
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

	input := &agentsv1.CreateVolumeRequest{
		Name:      plan.Name.ValueString(),
		MountPath: plan.MountPath.ValueString(),
	}
	// size is what makes a volume persistent: the resource makes the two
	// biconditional, so there is no separate flag to disagree with.
	if size := stringValue(plan.Size); size != "" {
		input.Size = size
		input.Persistent = true
	}
	if class := stringValue(plan.StorageClass); class != "" {
		input.StorageClass = &class
	}
	if ttl := stringValue(plan.TTL); ttl != "" {
		input.Ttl = &ttl
	}
	switch {
	case !plan.EnvironmentID.IsNull():
		input.Target = &agentsv1.CreateVolumeRequest_EnvironmentId{EnvironmentId: plan.EnvironmentID.ValueString()}
	case !plan.McpID.IsNull():
		input.Target = &agentsv1.CreateVolumeRequest_McpId{McpId: plan.McpID.ValueString()}
	default:
		resp.Diagnostics.AddError("Missing volume target", "Set exactly one of environment_id or mcp_id")
		return
	}

	volume, err := r.client.CreateVolume(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create volume", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, volumeStateFrom(volume, plan))...)
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
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read volume", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, volumeStateFrom(volume, state))...)
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

	input := &agentsv1.UpdateVolumeRequest{
		Id:           plan.ID.ValueString(),
		Name:         updateStringPointer(plan.Name, state.Name),
		MountPath:    updateStringPointer(plan.MountPath, state.MountPath),
		Size:         updateStringPointer(plan.Size, state.Size),
		StorageClass: updateStringPointer(plan.StorageClass, state.StorageClass),
		Ttl:          updateStringPointer(plan.TTL, state.TTL),
	}
	if input.Size != nil {
		persistent := *input.Size != ""
		input.Persistent = &persistent
	}

	volume, err := r.client.UpdateVolume(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update volume", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, volumeStateFrom(volume, state))...)
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
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete volume", err.Error())
		return
	}
}

func (r *volumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// volumeStateFrom keeps the target from the prior model: it is immutable and the
// API returns it in a oneof the state shape splits into two attributes.
func volumeStateFrom(volume *agentsv1.Volume, prior volumeModel) *volumeModel {
	return &volumeModel{
		ID:            types.StringValue(volume.Meta.Id),
		EnvironmentID: prior.EnvironmentID,
		McpID:         prior.McpID,
		Name:          types.StringValue(volume.Name),
		MountPath:     types.StringValue(volume.MountPath),
		Size:          optionalString(volume.Size),
		StorageClass:  optionalString(volume.GetStorageClass()),
		TTL:           optionalString(volume.GetTtl()),
	}
}
