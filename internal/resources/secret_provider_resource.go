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

	secretsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/secrets/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type secretProviderResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &secretProviderResource{}
var _ resource.ResourceWithImportState = &secretProviderResource{}

type secretProviderVaultModel struct {
	Address types.String `tfsdk:"address"`
	Token   types.String `tfsdk:"token"`
}

type secretProviderModel struct {
	ID             types.String              `tfsdk:"id"`
	OrganizationID types.String              `tfsdk:"organization_id"`
	Name           types.String              `tfsdk:"name"`
	Description    types.String              `tfsdk:"description"`
	Type           types.String              `tfsdk:"type"`
	Vault          *secretProviderVaultModel `tfsdk:"vault"`
}

func NewSecretProviderResource() resource.Resource { return &secretProviderResource{} }

func (r *secretProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_provider"
}

func (r *secretProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn secret provider.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the secret provider.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization identifier for the secret provider.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Secret provider name.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable description.",
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Secret provider type. Only \"vault\" is supported.",
				Validators:          []validator.String{stringvalidator.OneOf("vault")},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"vault": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "Vault configuration.",
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Vault server address.",
					},
					"token": schema.StringAttribute{
						Required:            true,
						Sensitive:           true,
						MarkdownDescription: "Vault authentication token.",
					},
				},
			},
		},
	}
}

func (r *secretProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *secretProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan secretProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.Vault == nil {
		resp.Diagnostics.AddError("Missing vault configuration", "secret provider requires vault configuration")
		return
	}

	input := &secretsv1.CreateSecretProviderRequest{
		Title:          plan.Name.ValueString(),
		Description:    stringValue(plan.Description),
		Type:           toProtoSecretProviderType(plan.Type.ValueString()),
		Config:         secretProviderConfigFromModel(plan.Vault),
		OrganizationId: plan.OrganizationID.ValueString(),
	}

	provider, err := r.client.CreateSecretProvider(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create secret provider", err.Error())
		return
	}

	updatedState := secretProviderModel{
		ID:             types.StringValue(provider.Meta.Id),
		OrganizationID: plan.OrganizationID,
		Name:           types.StringValue(provider.Title),
		Description:    optionalString(provider.Description),
		Type:           types.StringValue(fromProtoSecretProviderType(provider.Type)),
		Vault:          secretProviderVaultState(provider, plan.Vault.Token),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *secretProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state secretProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	provider, err := r.client.GetSecretProvider(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read secret provider", err.Error())
		return
	}

	fallbackToken := types.StringNull()
	if state.Vault != nil {
		fallbackToken = state.Vault.Token
	}
	state.Name = types.StringValue(provider.Title)
	state.Description = optionalString(provider.Description)
	state.Type = types.StringValue(fromProtoSecretProviderType(provider.Type))
	state.Vault = secretProviderVaultState(provider, fallbackToken)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *secretProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan secretProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state secretProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.Vault == nil {
		resp.Diagnostics.AddError("Missing vault configuration", "secret provider requires vault configuration")
		return
	}

	input := &secretsv1.UpdateSecretProviderRequest{
		Id:          plan.ID.ValueString(),
		Title:       updateStringPointer(plan.Name, state.Name),
		Description: updateStringPointer(plan.Description, state.Description),
		Config:      secretProviderConfigFromModel(plan.Vault),
	}

	provider, err := r.client.UpdateSecretProvider(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update secret provider", err.Error())
		return
	}

	updatedState := secretProviderModel{
		ID:             types.StringValue(provider.Meta.Id),
		OrganizationID: state.OrganizationID,
		Name:           types.StringValue(provider.Title),
		Description:    optionalString(provider.Description),
		Type:           types.StringValue(fromProtoSecretProviderType(provider.Type)),
		Vault:          secretProviderVaultState(provider, plan.Vault.Token),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *secretProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state secretProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSecretProvider(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete secret provider", err.Error())
		return
	}
}

func (r *secretProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func secretProviderConfigFromModel(vault *secretProviderVaultModel) *secretsv1.SecretProviderConfig {
	return &secretsv1.SecretProviderConfig{
		Provider: &secretsv1.SecretProviderConfig_Vault{
			Vault: &secretsv1.VaultConfig{
				Address: vault.Address.ValueString(),
				Token:   vault.Token.ValueString(),
			},
		},
	}
}

func secretProviderVaultState(provider *secretsv1.SecretProvider, fallbackToken types.String) *secretProviderVaultModel {
	if provider.Type != secretsv1.SecretProviderType_SECRET_PROVIDER_TYPE_VAULT {
		panic(fmt.Sprintf("unexpected secret provider type: %v", provider.Type))
	}
	config := provider.GetConfig()
	if config == nil {
		panic("unexpected secret provider config")
	}
	vault := config.GetVault()
	if vault == nil {
		panic("unexpected secret provider vault config")
	}

	return &secretProviderVaultModel{
		Address: types.StringValue(vault.Address),
		Token:   preserveSensitiveString(fallbackToken, vault.Token),
	}
}

func toProtoSecretProviderType(v string) secretsv1.SecretProviderType {
	switch v {
	case "vault":
		return secretsv1.SecretProviderType_SECRET_PROVIDER_TYPE_VAULT
	default:
		panic("unreachable: validated by schema")
	}
}

func fromProtoSecretProviderType(v secretsv1.SecretProviderType) string {
	switch v {
	case secretsv1.SecretProviderType_SECRET_PROVIDER_TYPE_VAULT:
		return "vault"
	default:
		panic(fmt.Sprintf("unreachable: unexpected secret provider type %v", v))
	}
}
