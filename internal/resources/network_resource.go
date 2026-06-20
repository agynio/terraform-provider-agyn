package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	networksv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/networks/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type networkResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &networkResource{}
var _ resource.ResourceWithImportState = &networkResource{}

type networkModel struct {
	ID                types.String `tfsdk:"id"`
	OrganizationID    types.String `tfsdk:"organization_id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	ProvisioningState types.String `tfsdk:"provisioning_state"`
}

func NewNetworkResource() resource.Resource { return &networkResource{} }

func (r *networkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *networkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn private network.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the private network.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization identifier for the private network.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Private network name.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable description.",
			},
			"provisioning_state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "OpenZiti provisioning state for the private network.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *networkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *networkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan networkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	network, err := r.client.CreateNetwork(ctx, &networksv1.CreateNetworkRequest{OrganizationId: plan.OrganizationID.ValueString(), Name: plan.Name.ValueString(), Description: stringValue(plan.Description)})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create private network", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, networkState(network))...)
}

func (r *networkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state networkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	network, err := r.client.GetNetwork(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read private network", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, networkState(network))...)
}

func (r *networkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan networkModel
	var state networkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	network, err := r.client.UpdateNetwork(ctx, &networksv1.UpdateNetworkRequest{Id: plan.ID.ValueString(), Name: updateStringPointer(plan.Name, state.Name), Description: updateStringPointer(plan.Description, state.Description)})
	if err != nil {
		resp.Diagnostics.AddError("Unable to update private network", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, networkState(network))...)
}

func (r *networkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state networkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteNetwork(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete private network", err.Error())
	}
}

func (r *networkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func networkState(network *networksv1.Network) networkModel {
	return networkModel{ID: types.StringValue(network.GetMeta().GetId()), OrganizationID: types.StringValue(network.GetOrganizationId()), Name: types.StringValue(network.GetName()), Description: optionalString(network.GetDescription()), ProvisioningState: optionalString(provisioningStateToString(network.GetProvisioningState()))}
}
