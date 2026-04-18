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

	usersv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/users/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type userResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &userResource{}
var _ resource.ResourceWithImportState = &userResource{}

type userModel struct {
	IdentityID  types.String `tfsdk:"identity_id"`
	OIDCSubject types.String `tfsdk:"oidc_subject"`
	Name        types.String `tfsdk:"name"`
	PhotoURL    types.String `tfsdk:"photo_url"`
	Nickname    types.String `tfsdk:"nickname"`
	ClusterRole types.String `tfsdk:"cluster_role"`
}

func NewUserResource() resource.Resource { return &userResource{} }

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn user.",
		Attributes: map[string]schema.Attribute{
			"identity_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identity identifier for the user.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"oidc_subject": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "OIDC subject for the user.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Display name for the user.",
			},
			"photo_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Photo URL for the user.",
			},
			"nickname": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Nickname for the user.",
			},
			"cluster_role": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Cluster role for the user (admin or none).",
				Validators:          []validator.String{stringvalidator.OneOf("admin", "none")},
			},
		},
	}
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &usersv1.CreateUserRequest{OidcSubject: plan.OIDCSubject.ValueString()}
	input.Name = planStringPointer(plan.Name)
	input.PhotoUrl = planStringPointer(plan.PhotoURL)
	input.Nickname = planStringPointer(plan.Nickname)

	user, err := r.client.CreateUser(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create user", err.Error())
		return
	}

	identityID, err := userIdentityID(user)
	if err != nil {
		resp.Diagnostics.AddError("Invalid create user response", err.Error())
		return
	}

	if clusterRole := clusterRolePointerFromPlan(plan.ClusterRole); clusterRole != nil {
		updateReq := &usersv1.UpdateUserRequest{
			IdentityId:  identityID,
			ClusterRole: clusterRole,
		}
		if _, err := r.client.UpdateUser(ctx, updateReq); err != nil {
			resp.Diagnostics.AddError("Unable to update user cluster role", err.Error())
			return
		}
	}

	updatedState, err := r.readUser(ctx, identityID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read user", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedState, err := r.readUser(ctx, state.IdentityID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read user", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &usersv1.UpdateUserRequest{IdentityId: state.IdentityID.ValueString()}
	needsUpdate := false
	if value := updateStringPointer(plan.Name, state.Name); value != nil {
		input.Name = value
		needsUpdate = true
	}
	if value := updateStringPointer(plan.PhotoURL, state.PhotoURL); value != nil {
		input.PhotoUrl = value
		needsUpdate = true
	}
	if value := updateStringPointer(plan.Nickname, state.Nickname); value != nil {
		input.Nickname = value
		needsUpdate = true
	}
	if value := updateClusterRolePointer(plan.ClusterRole, state.ClusterRole); value != nil {
		input.ClusterRole = value
		needsUpdate = true
	}

	if needsUpdate {
		if _, err := r.client.UpdateUser(ctx, input); err != nil {
			resp.Diagnostics.AddError("Unable to update user", err.Error())
			return
		}
	}

	updatedState, err := r.readUser(ctx, state.IdentityID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read user", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteUser(ctx, state.IdentityID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete user", err.Error())
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identity_id"), req, resp)
}

func (r *userResource) readUser(ctx context.Context, identityID string) (userModel, error) {
	user, clusterRole, err := r.client.GetUser(ctx, identityID)
	if err != nil {
		return userModel{}, err
	}
	value, err := userIdentityID(user)
	if err != nil {
		return userModel{}, err
	}

	return userModel{
		IdentityID:  types.StringValue(value),
		OIDCSubject: types.StringValue(user.GetOidcSubject()),
		Name:        optionalString(user.GetName()),
		PhotoURL:    optionalString(user.GetPhotoUrl()),
		Nickname:    optionalString(user.GetNickname()),
		ClusterRole: types.StringValue(fromProtoClusterRole(clusterRole)),
	}, nil
}

func userIdentityID(user *usersv1.User) (string, error) {
	if user == nil {
		return "", fmt.Errorf("missing user")
	}
	meta := user.GetMeta()
	if meta == nil || meta.Id == "" {
		return "", fmt.Errorf("missing user identity id")
	}
	return meta.Id, nil
}

func planStringPointer(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	result := value.ValueString()
	return &result
}

func clusterRolePointerFromPlan(value types.String) *usersv1.ClusterRole {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	role := toProtoClusterRole(value.ValueString())
	return &role
}

func updateClusterRolePointer(plan types.String, prior types.String) *usersv1.ClusterRole {
	if plan.IsUnknown() {
		return nil
	}
	if plan.IsNull() {
		if prior.IsNull() || prior.IsUnknown() {
			return nil
		}
		value := usersv1.ClusterRole_CLUSTER_ROLE_UNSPECIFIED
		return &value
	}
	value := toProtoClusterRole(plan.ValueString())
	return &value
}

func toProtoClusterRole(v string) usersv1.ClusterRole {
	switch v {
	case "admin":
		return usersv1.ClusterRole_CLUSTER_ROLE_ADMIN
	case "none":
		return usersv1.ClusterRole_CLUSTER_ROLE_UNSPECIFIED
	default:
		panic("unreachable: validated by schema")
	}
}

func fromProtoClusterRole(role usersv1.ClusterRole) string {
	switch role {
	case usersv1.ClusterRole_CLUSTER_ROLE_ADMIN:
		return "admin"
	case usersv1.ClusterRole_CLUSTER_ROLE_UNSPECIFIED:
		return "none"
	default:
		panic(fmt.Sprintf("unreachable: unexpected cluster role %v", role))
	}
}
