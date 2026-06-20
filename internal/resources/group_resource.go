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

	groupsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/groups/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type groupResource struct{ client *agentapi.Client }

var _ resource.Resource = &groupResource{}
var _ resource.ResourceWithImportState = &groupResource{}

type groupModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Source         types.String `tfsdk:"source"`
	ExternalID     types.String `tfsdk:"external_id"`
}

func NewGroupResource() resource.Resource { return &groupResource{} }

func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Manages an Agyn group.", Attributes: map[string]schema.Attribute{
		"id":              schema.StringAttribute{Computed: true, MarkdownDescription: "UUID identifier of the group.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"organization_id": schema.StringAttribute{Required: true, MarkdownDescription: "Organization identifier for the group.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()}},
		"name":            schema.StringAttribute{Required: true, MarkdownDescription: "Group name."},
		"description":     schema.StringAttribute{Optional: true, MarkdownDescription: "Human-readable description."},
		"source":          schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Group source. One of `platform` or `scim`.", Validators: []validator.String{stringvalidator.OneOf("platform", "scim")}, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"external_id":     schema.StringAttribute{Optional: true, MarkdownDescription: "External identity-provider group identifier.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
	}}
}

func (r *groupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan groupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	source, err := groupSourceFromString(stringValue(plan.Source))
	if err != nil {
		resp.Diagnostics.AddError("Invalid group", err.Error())
		return
	}
	input := &groupsv1.CreateGroupRequest{OrganizationId: plan.OrganizationID.ValueString(), Name: plan.Name.ValueString(), Description: stringValue(plan.Description), Source: source}
	if !plan.ExternalID.IsNull() && !plan.ExternalID.IsUnknown() {
		input.ExternalId = stringPtr(plan.ExternalID.ValueString())
	}
	group, err := r.client.CreateGroup(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create group", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, groupState(group))...)
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state groupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := r.client.GetGroup(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read group", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, groupState(group))...)
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan groupModel
	var state groupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := r.client.UpdateGroup(ctx, &groupsv1.UpdateGroupRequest{Id: plan.ID.ValueString(), Name: updateStringPointer(plan.Name, state.Name), Description: updateStringPointer(plan.Description, state.Description)})
	if err != nil {
		resp.Diagnostics.AddError("Unable to update group", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, groupState(group))...)
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state groupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteGroup(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete group", err.Error())
	}
}

func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func groupState(group *groupsv1.Group) groupModel {
	externalID := types.StringNull()
	if group.ExternalId != nil {
		externalID = types.StringValue(group.GetExternalId())
	}
	return groupModel{ID: types.StringValue(group.GetMeta().GetId()), OrganizationID: types.StringValue(group.GetOrganizationId()), Name: types.StringValue(group.GetName()), Description: optionalString(group.GetDescription()), Source: types.StringValue(groupSourceToString(group.GetSource())), ExternalID: externalID}
}
