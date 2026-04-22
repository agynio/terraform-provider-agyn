package resources

import (
	"context"
	"regexp"

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

type secretResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &secretResource{}
var _ resource.ResourceWithImportState = &secretResource{}

type secretModel struct {
	ID                 types.String `tfsdk:"id"`
	OrganizationID     types.String `tfsdk:"organization_id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Value              types.String `tfsdk:"value"`
	ProviderID         types.String `tfsdk:"provider_id"`
	ProviderSecretName types.String `tfsdk:"provider_secret_name"`
}

var providerSecretNameRegex = regexp.MustCompile("^[^/]+/.+/[^/]+$")

func NewSecretResource() resource.Resource { return &secretResource{} }

func (r *secretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *secretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	sourceValidators := []validator.String{
		stringvalidator.ExactlyOneOf(
			path.MatchRoot("value"),
			path.MatchRoot("provider_id"),
		),
	}
	providerValidators := append([]validator.String{}, sourceValidators...)
	providerValidators = append(providerValidators, stringvalidator.AlsoRequires(path.MatchRoot("provider_secret_name")))
	providerSecretNameValidators := []validator.String{
		stringvalidator.AlsoRequires(path.MatchRoot("provider_id")),
		stringvalidator.RegexMatches(providerSecretNameRegex, "must be in the format <mount>/<path>/<key>"),
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn secret.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the secret.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization identifier for the secret.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Secret name.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable description.",
			},
			"value": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Plain-text secret value.",
				Validators:          sourceValidators,
			},
			"provider_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Secret provider identifier for remote secrets.",
				Validators:          providerValidators,
			},
			"provider_secret_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Remote secret reference name. Format: <mount>/<path>/<key>.",
				Validators:          providerSecretNameValidators,
			},
		},
	}
}

func (r *secretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *secretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan secretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &secretsv1.CreateSecretRequest{
		Title:          plan.Name.ValueString(),
		Description:    stringValue(plan.Description),
		OrganizationId: plan.OrganizationID.ValueString(),
	}
	if setSecretCreateSource(input, plan.Value, plan.ProviderID, plan.ProviderSecretName, resp) {
		return
	}

	secret, err := r.client.CreateSecret(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create secret", err.Error())
		return
	}

	value, providerID, providerSecretName := secretSourceState(secret, plan.Value)
	updatedState := secretModel{
		ID:                 types.StringValue(secret.Meta.Id),
		OrganizationID:     plan.OrganizationID,
		Name:               types.StringValue(secret.Title),
		Description:        optionalString(secret.Description),
		Value:              value,
		ProviderID:         providerID,
		ProviderSecretName: providerSecretName,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *secretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state secretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.GetSecret(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read secret", err.Error())
		return
	}

	value, providerID, providerSecretName := secretSourceState(secret, state.Value)
	state.Name = types.StringValue(secret.Title)
	state.Description = optionalString(secret.Description)
	state.Value = value
	state.ProviderID = providerID
	state.ProviderSecretName = providerSecretName

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *secretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan secretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state secretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &secretsv1.UpdateSecretRequest{
		Id:               plan.ID.ValueString(),
		Title:            updateStringPointer(plan.Name, state.Name),
		Description:      updateStringPointer(plan.Description, state.Description),
		SecretProviderId: updateStringPointer(plan.ProviderID, state.ProviderID),
		RemoteName:       updateStringPointer(plan.ProviderSecretName, state.ProviderSecretName),
		Value:            updateStringPointer(plan.Value, state.Value),
	}

	secret, err := r.client.UpdateSecret(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update secret", err.Error())
		return
	}

	value, providerID, providerSecretName := secretSourceState(secret, plan.Value)
	updatedState := secretModel{
		ID:                 types.StringValue(secret.Meta.Id),
		OrganizationID:     state.OrganizationID,
		Name:               types.StringValue(secret.Title),
		Description:        optionalString(secret.Description),
		Value:              value,
		ProviderID:         providerID,
		ProviderSecretName: providerSecretName,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *secretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state secretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSecret(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete secret", err.Error())
		return
	}
}

func (r *secretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func setSecretCreateSource(req *secretsv1.CreateSecretRequest, value types.String, providerID types.String, providerSecretName types.String, resp *resource.CreateResponse) bool {
	if !value.IsNull() && !value.IsUnknown() {
		req.Value = value.ValueString()
		return false
	}
	if !providerID.IsNull() && !providerID.IsUnknown() {
		if providerSecretName.IsNull() || providerSecretName.IsUnknown() {
			resp.Diagnostics.AddError("Missing provider secret name", "provider_secret_name must be set when provider_id is provided")
			return true
		}
		req.SecretProviderId = providerID.ValueString()
		req.RemoteName = providerSecretName.ValueString()
		return false
	}
	resp.Diagnostics.AddError("Missing secret source", "create secret requires value or provider_id")
	return true
}

func secretSourceState(secret *secretsv1.Secret, fallback types.String) (types.String, types.String, types.String) {
	value := types.StringNull()
	providerID := types.StringNull()
	providerSecretName := types.StringNull()

	if secret.SecretProviderId != "" || secret.RemoteName != "" {
		if secret.SecretProviderId == "" {
			panic("unexpected secret remote reference")
		}
		providerID = types.StringValue(secret.SecretProviderId)
		providerSecretName = optionalString(secret.RemoteName)
		return value, providerID, providerSecretName
	}

	value = preserveSensitiveString(fallback, secret.Value)
	return value, providerID, providerSecretName
}
