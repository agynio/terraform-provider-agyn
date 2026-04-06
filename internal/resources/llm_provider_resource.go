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

	llmv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/llm/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type llmProviderResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &llmProviderResource{}
var _ resource.ResourceWithImportState = &llmProviderResource{}

type llmProviderModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Endpoint       types.String `tfsdk:"endpoint"`
	AuthMethod     types.String `tfsdk:"auth_method"`
	Token          types.String `tfsdk:"token"`
}

func NewLLMProviderResource() resource.Resource { return &llmProviderResource{} }

func (r *llmProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_llm_provider"
}

func (r *llmProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn LLM provider.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the LLM provider.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization identifier for the LLM provider.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"endpoint": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Provider base URL.",
			},
			"auth_method": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Authentication method for the provider.",
				Validators:          []validator.String{stringvalidator.OneOf("bearer")},
			},
			"token": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Bearer token for the provider.",
			},
		},
	}
}

func (r *llmProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *llmProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan llmProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &llmv1.CreateLLMProviderRequest{
		Endpoint:       plan.Endpoint.ValueString(),
		AuthMethod:     toProtoAuthMethod(plan.AuthMethod.ValueString()),
		Token:          plan.Token.ValueString(),
		OrganizationId: plan.OrganizationID.ValueString(),
	}

	provider, err := r.client.CreateLLMProvider(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create LLM provider", err.Error())
		return
	}

	updatedState := llmProviderModel{
		ID:             types.StringValue(provider.Meta.Id),
		OrganizationID: types.StringValue(provider.OrganizationId),
		Endpoint:       types.StringValue(provider.Endpoint),
		AuthMethod:     types.StringValue(fromProtoAuthMethod(provider.AuthMethod)),
		Token:          types.StringValue(plan.Token.ValueString()),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *llmProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state llmProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	provider, err := r.client.GetLLMProvider(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read LLM provider", err.Error())
		return
	}

	state.ID = types.StringValue(provider.Meta.Id)
	state.OrganizationID = types.StringValue(provider.OrganizationId)
	state.Endpoint = types.StringValue(provider.Endpoint)
	state.AuthMethod = types.StringValue(fromProtoAuthMethod(provider.AuthMethod))
	state.Token = preserveSensitiveString(state.Token, "")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *llmProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan llmProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state llmProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &llmv1.UpdateLLMProviderRequest{
		Id:         plan.ID.ValueString(),
		Endpoint:   updateStringPointer(plan.Endpoint, state.Endpoint),
		AuthMethod: updateAuthMethodPointer(plan.AuthMethod, state.AuthMethod),
		Token:      updateStringPointer(plan.Token, state.Token),
	}

	provider, err := r.client.UpdateLLMProvider(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update LLM provider", err.Error())
		return
	}

	updatedState := llmProviderModel{
		ID:             types.StringValue(provider.Meta.Id),
		OrganizationID: types.StringValue(provider.OrganizationId),
		Endpoint:       types.StringValue(provider.Endpoint),
		AuthMethod:     types.StringValue(fromProtoAuthMethod(provider.AuthMethod)),
		Token:          types.StringValue(plan.Token.ValueString()),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *llmProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state llmProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteLLMProvider(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete LLM provider", err.Error())
		return
	}
}

func (r *llmProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func updateAuthMethodPointer(plan types.String, prior types.String) *llmv1.AuthMethod {
	if plan.IsUnknown() {
		return nil
	}
	if plan.IsNull() {
		if prior.IsNull() || prior.IsUnknown() {
			return nil
		}
		value := llmv1.AuthMethod_AUTH_METHOD_UNSPECIFIED
		return &value
	}
	value := toProtoAuthMethod(plan.ValueString())
	return &value
}

func toProtoAuthMethod(v string) llmv1.AuthMethod {
	switch v {
	case "bearer":
		return llmv1.AuthMethod_AUTH_METHOD_BEARER
	default:
		panic("unreachable: validated by schema")
	}
}

func fromProtoAuthMethod(v llmv1.AuthMethod) string {
	switch v {
	case llmv1.AuthMethod_AUTH_METHOD_BEARER:
		return "bearer"
	default:
		panic(fmt.Sprintf("unreachable: unexpected proto auth method %v", v))
	}
}
