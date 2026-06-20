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

	groupsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/groups/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type groupMembershipResource struct{ client *agentapi.Client }

var _ resource.Resource = &groupMembershipResource{}
var _ resource.ResourceWithImportState = &groupMembershipResource{}

type groupMembershipModel struct {
	ID         types.String `tfsdk:"id"`
	GroupID    types.String `tfsdk:"group_id"`
	MemberType types.String `tfsdk:"member_type"`
	MemberID   types.String `tfsdk:"member_id"`
	Source     types.String `tfsdk:"source"`
}

func NewGroupMembershipResource() resource.Resource { return &groupMembershipResource{} }

func (r *groupMembershipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_membership"
}

func (r *groupMembershipResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Manages an Agyn group membership.", Attributes: map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "UUID identifier of the group membership.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"group_id":    schema.StringAttribute{Required: true, MarkdownDescription: "Group identifier.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"member_type": schema.StringAttribute{Required: true, MarkdownDescription: "Member type. One of `user`, `agent`, or `app`.", Validators: []validator.String{stringvalidator.OneOf("user", "agent", "app")}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"member_id":   schema.StringAttribute{Required: true, MarkdownDescription: "Member identifier.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"source":      schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Membership source. One of `platform` or `scim`.", Validators: []validator.String{stringvalidator.OneOf("platform", "scim")}, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func (r *groupMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *groupMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan groupMembershipModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	memberType, err := groupMemberTypeFromString(plan.MemberType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid group membership", err.Error())
		return
	}
	source, err := groupSourceFromString(stringValue(plan.Source))
	if err != nil {
		resp.Diagnostics.AddError("Invalid group membership", err.Error())
		return
	}
	membership, err := r.client.AddGroupMember(ctx, &groupsv1.AddMemberRequest{GroupId: plan.GroupID.ValueString(), MemberType: memberType, MemberId: plan.MemberID.ValueString(), Source: source})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create group membership", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, groupMembershipState(membership))...)
}

func (r *groupMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state groupMembershipModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	memberType, err := groupMemberTypeFromString(state.MemberType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid group membership", err.Error())
		return
	}
	membership, err := r.client.GetGroupMembershipByGroupAndMember(ctx, state.GroupID.ValueString(), memberType, state.MemberID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read group membership", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, groupMembershipState(membership))...)
}

func (r *groupMembershipResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Group memberships are immutable. This is an internal error.")
}

func (r *groupMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state groupMembershipModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RemoveGroupMember(ctx, state.GroupID.ValueString(), state.MemberID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete group membership", err.Error())
	}
}

func (r *groupMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 4 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected group_id:member_type:member_id:membership_id")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("member_type"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("member_id"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[3])...)
}

func groupMembershipState(membership *groupsv1.GroupMembership) groupMembershipModel {
	return groupMembershipModel{ID: types.StringValue(membership.GetMeta().GetId()), GroupID: types.StringValue(membership.GetGroupId()), MemberType: types.StringValue(groupMemberTypeToString(membership.GetMemberType())), MemberID: types.StringValue(membership.GetMemberId()), Source: types.StringValue(groupSourceToString(membership.GetSource()))}
}
