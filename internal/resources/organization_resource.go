package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	organizationsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/organizations/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type organizationResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &organizationResource{}
var _ resource.ResourceWithImportState = &organizationResource{}

type organizationModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewOrganizationResource() resource.Resource { return &organizationResource{} }

func (r *organizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *organizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the organization.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization name.",
			},
		},
	}
}

func (r *organizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan organizationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &organizationsv1.CreateOrganizationRequest{Name: plan.Name.ValueString()}
	org, err := r.client.CreateOrganization(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create organization", err.Error())
		return
	}

	state := organizationModel{
		ID:   types.StringValue(org.Id),
		Name: types.StringValue(org.Name),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state organizationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.GetOrganization(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read organization", err.Error())
		return
	}

	state.ID = types.StringValue(org.Id)
	state.Name = types.StringValue(org.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan organizationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state organizationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &organizationsv1.UpdateOrganizationRequest{
		Id:   state.ID.ValueString(),
		Name: updateStringPointer(plan.Name, state.Name),
	}

	org, err := r.client.UpdateOrganization(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update organization", err.Error())
		return
	}

	updatedState := organizationModel{
		ID:   types.StringValue(org.Id),
		Name: types.StringValue(org.Name),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *organizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state organizationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteOrganization(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete organization", err.Error())
		return
	}
}

func (r *organizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
