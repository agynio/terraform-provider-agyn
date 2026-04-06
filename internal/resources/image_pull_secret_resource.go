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

	secretsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/secrets/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type imagePullSecretResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &imagePullSecretResource{}
var _ resource.ResourceWithImportState = &imagePullSecretResource{}

type imagePullSecretModel struct {
	ID                     types.String `tfsdk:"id"`
	OrganizationID         types.String `tfsdk:"organization_id"`
	Description            types.String `tfsdk:"description"`
	Registry               types.String `tfsdk:"registry"`
	Username               types.String `tfsdk:"username"`
	Password               types.String `tfsdk:"password"`
	RemoteSecretProviderID types.String `tfsdk:"remote_secret_provider_id"`
	RemoteSecretReference  types.String `tfsdk:"remote_secret_reference"`
}

func NewImagePullSecretResource() resource.Resource { return &imagePullSecretResource{} }

func (r *imagePullSecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_pull_secret"
}

func (r *imagePullSecretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	sourceValidators := []validator.String{
		stringvalidator.ExactlyOneOf(
			path.MatchRoot("password"),
			path.MatchRoot("remote_secret_provider_id"),
		),
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn image pull secret.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the image pull secret.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization identifier for the image pull secret.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable description.",
			},
			"registry": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Registry URL or hostname.",
			},
			"username": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Registry username.",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Registry password.",
				Validators:          sourceValidators,
			},
			"remote_secret_provider_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Remote secret provider identifier.",
				Validators:          sourceValidators,
			},
			"remote_secret_reference": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Remote secret reference key.",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("remote_secret_provider_id")),
				},
			},
		},
	}
}

func (r *imagePullSecretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *imagePullSecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan imagePullSecretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &secretsv1.CreateImagePullSecretRequest{
		OrganizationId: plan.OrganizationID.ValueString(),
		Description:    stringValue(plan.Description),
		Registry:       plan.Registry.ValueString(),
		Username:       plan.Username.ValueString(),
	}
	if setImagePullSecretCreateSource(input, plan.Password, plan.RemoteSecretProviderID, plan.RemoteSecretReference, resp) {
		return
	}

	secret, err := r.client.CreateImagePullSecret(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create image pull secret", err.Error())
		return
	}

	password, remoteProviderID, remoteReference := imagePullSecretSourceState(secret, plan.Password)
	updatedState := imagePullSecretModel{
		ID:                     types.StringValue(secret.Meta.Id),
		OrganizationID:         plan.OrganizationID,
		Description:            optionalString(secret.Description),
		Registry:               types.StringValue(secret.Registry),
		Username:               types.StringValue(secret.Username),
		Password:               password,
		RemoteSecretProviderID: remoteProviderID,
		RemoteSecretReference:  remoteReference,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *imagePullSecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state imagePullSecretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.GetImagePullSecret(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read image pull secret", err.Error())
		return
	}

	password, remoteProviderID, remoteReference := imagePullSecretSourceState(secret, state.Password)
	state.Description = optionalString(secret.Description)
	state.Registry = types.StringValue(secret.Registry)
	state.Username = types.StringValue(secret.Username)
	state.Password = password
	state.RemoteSecretProviderID = remoteProviderID
	state.RemoteSecretReference = remoteReference

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *imagePullSecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan imagePullSecretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state imagePullSecretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &secretsv1.UpdateImagePullSecretRequest{
		Id:          plan.ID.ValueString(),
		Description: updateStringPointer(plan.Description, state.Description),
		Registry:    updateStringPointer(plan.Registry, state.Registry),
		Username:    updateStringPointer(plan.Username, state.Username),
	}
	if setImagePullSecretUpdateSource(input, plan.Password, plan.RemoteSecretProviderID, plan.RemoteSecretReference, resp) {
		return
	}

	secret, err := r.client.UpdateImagePullSecret(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update image pull secret", err.Error())
		return
	}

	password, remoteProviderID, remoteReference := imagePullSecretSourceState(secret, plan.Password)
	updatedState := imagePullSecretModel{
		ID:                     types.StringValue(secret.Meta.Id),
		OrganizationID:         state.OrganizationID,
		Description:            optionalString(secret.Description),
		Registry:               types.StringValue(secret.Registry),
		Username:               types.StringValue(secret.Username),
		Password:               password,
		RemoteSecretProviderID: remoteProviderID,
		RemoteSecretReference:  remoteReference,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *imagePullSecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state imagePullSecretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteImagePullSecret(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete image pull secret", err.Error())
		return
	}
}

func (r *imagePullSecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func setImagePullSecretCreateSource(req *secretsv1.CreateImagePullSecretRequest, password types.String, remoteProviderID types.String, remoteReference types.String, resp *resource.CreateResponse) bool {
	if !password.IsNull() && !password.IsUnknown() {
		req.Source = &secretsv1.CreateImagePullSecretRequest_Value{Value: password.ValueString()}
		return false
	}
	if !remoteProviderID.IsNull() && !remoteProviderID.IsUnknown() {
		req.Source = &secretsv1.CreateImagePullSecretRequest_Remote{
			Remote: &secretsv1.RemoteSecretRef{
				ValueProviderId: remoteProviderID.ValueString(),
				ValueReference:  stringValue(remoteReference),
			},
		}
		return false
	}
	resp.Diagnostics.AddError("Missing image pull secret source", "create image pull secret requires password or remote_secret_provider_id")
	return true
}

func setImagePullSecretUpdateSource(req *secretsv1.UpdateImagePullSecretRequest, password types.String, remoteProviderID types.String, remoteReference types.String, resp *resource.UpdateResponse) bool {
	if !password.IsNull() && !password.IsUnknown() {
		req.Source = &secretsv1.UpdateImagePullSecretRequest_Value{Value: password.ValueString()}
		return false
	}
	if !remoteProviderID.IsNull() && !remoteProviderID.IsUnknown() {
		req.Source = &secretsv1.UpdateImagePullSecretRequest_Remote{
			Remote: &secretsv1.RemoteSecretRef{
				ValueProviderId: remoteProviderID.ValueString(),
				ValueReference:  stringValue(remoteReference),
			},
		}
		return false
	}
	resp.Diagnostics.AddError("Missing image pull secret source", "update image pull secret requires password or remote_secret_provider_id")
	return true
}

func imagePullSecretSourceState(secret *secretsv1.ImagePullSecret, fallback types.String) (types.String, types.String, types.String) {
	password := types.StringNull()
	remoteProviderID := types.StringNull()
	remoteReference := types.StringNull()

	switch source := secret.GetSource().(type) {
	case *secretsv1.ImagePullSecret_Value:
		password = preserveSensitiveString(fallback, source.Value)
	case *secretsv1.ImagePullSecret_Remote:
		if source.Remote == nil {
			panic("unexpected image pull secret remote reference")
		}
		remoteProviderID = types.StringValue(source.Remote.ValueProviderId)
		remoteReference = optionalString(source.Remote.ValueReference)
	default:
		// API does not echo sensitive source back; preserve plan value
		password = fallback
	}

	return password, remoteProviderID, remoteReference
}
