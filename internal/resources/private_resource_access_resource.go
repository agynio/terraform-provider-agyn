package resources

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	networksv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/networks/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type privateResourceAccessResource struct{ client *agentapi.Client }

var _ resource.Resource = &privateResourceAccessResource{}
var _ resource.ResourceWithImportState = &privateResourceAccessResource{}

type privateResourceAccessModel struct {
	ID                types.String `tfsdk:"id"`
	PrivateResourceID types.String `tfsdk:"private_resource_id"`
	PrincipalType     types.String `tfsdk:"principal_type"`
	PrincipalID       types.String `tfsdk:"principal_id"`
	ProvisioningState types.String `tfsdk:"provisioning_state"`
}

func NewPrivateResourceAccessResource() resource.Resource { return &privateResourceAccessResource{} }

func (r *privateResourceAccessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_resource_access"
}

func (r *privateResourceAccessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Manages an Agyn private resource access grant.", Attributes: map[string]schema.Attribute{
		"id":                  schema.StringAttribute{Computed: true, MarkdownDescription: "UUID identifier of the private resource access grant.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"private_resource_id": schema.StringAttribute{Required: true, MarkdownDescription: "Private resource identifier.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"principal_type":      schema.StringAttribute{Required: true, MarkdownDescription: "Principal type. One of `agent`, `user`, `app`, or `group`.", Validators: []validator.String{stringvalidator.OneOf("agent", "user", "app", "group")}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"principal_id":        schema.StringAttribute{Required: true, MarkdownDescription: "Principal identifier.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"provisioning_state":  schema.StringAttribute{Computed: true, MarkdownDescription: "OpenZiti provisioning state.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func (r *privateResourceAccessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *privateResourceAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan privateResourceAccessModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	principalType, err := privateResourceAccessPrincipalTypeFromString(plan.PrincipalType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid private resource access", err.Error())
		return
	}
	access, err := r.client.CreatePrivateResourceAccess(ctx, &networksv1.CreatePrivateResourceAccessRequest{PrivateResourceId: plan.PrivateResourceID.ValueString(), PrincipalType: principalType, PrincipalId: plan.PrincipalID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create private resource access", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, privateResourceAccessState(access))...)
}

func (r *privateResourceAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state privateResourceAccessModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	principalType, err := privateResourceAccessPrincipalTypeFromString(state.PrincipalType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid private resource access", err.Error())
		return
	}
	access, err := r.client.GetPrivateResourceAccessByResourceAndPrincipal(ctx, state.PrivateResourceID.ValueString(), principalType, state.PrincipalID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read private resource access", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, privateResourceAccessState(access))...)
}

func (r *privateResourceAccessResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Private resource access grants are immutable. This is an internal error.")
}

func (r *privateResourceAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state privateResourceAccessModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeletePrivateResourceAccess(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete private resource access", err.Error())
	}
}

func (r *privateResourceAccessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 4 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected private_resource_id:principal_type:principal_id:access_id")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("private_resource_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("principal_type"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("principal_id"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[3])...)
}

func privateResourceAccessState(access *networksv1.PrivateResourceAccess) privateResourceAccessModel {
	return privateResourceAccessModel{ID: types.StringValue(access.GetMeta().GetId()), PrivateResourceID: types.StringValue(access.GetPrivateResourceId()), PrincipalType: types.StringValue(privateResourceAccessPrincipalTypeToString(access.GetPrincipalType())), PrincipalID: types.StringValue(access.GetPrincipalId()), ProvisioningState: optionalString(provisioningStateToString(access.GetProvisioningState()))}
}
