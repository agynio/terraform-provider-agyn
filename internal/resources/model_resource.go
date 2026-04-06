package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	llmv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/llm/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type modelResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &modelResource{}
var _ resource.ResourceWithImportState = &modelResource{}

type modelModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	LLMProviderID  types.String `tfsdk:"llm_provider_id"`
	RemoteName     types.String `tfsdk:"remote_name"`
}

func NewModelResource() resource.Resource { return &modelResource{} }

func (r *modelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model"
}

func (r *modelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn model.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the model.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization identifier for the model.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Model name.",
			},
			"llm_provider_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "LLM provider identifier.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"remote_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Remote model identifier.",
			},
		},
	}
}

func (r *modelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *modelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan modelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &llmv1.CreateModelRequest{
		Name:           plan.Name.ValueString(),
		LlmProviderId:  plan.LLMProviderID.ValueString(),
		RemoteName:     plan.RemoteName.ValueString(),
		OrganizationId: plan.OrganizationID.ValueString(),
	}

	model, err := r.client.CreateModel(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create model", err.Error())
		return
	}

	updatedState := modelModel{
		ID:             types.StringValue(model.Meta.Id),
		OrganizationID: types.StringValue(plan.OrganizationID.ValueString()),
		Name:           types.StringValue(model.Name),
		LLMProviderID:  types.StringValue(model.LlmProviderId),
		RemoteName:     types.StringValue(model.RemoteName),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *modelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state modelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, err := r.client.GetModel(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read model", err.Error())
		return
	}

	state.ID = types.StringValue(model.Meta.Id)
	state.Name = types.StringValue(model.Name)
	state.LLMProviderID = types.StringValue(model.LlmProviderId)
	state.RemoteName = types.StringValue(model.RemoteName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *modelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan modelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state modelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &llmv1.UpdateModelRequest{
		Id:            plan.ID.ValueString(),
		Name:          updateStringPointer(plan.Name, state.Name),
		LlmProviderId: updateStringPointer(plan.LLMProviderID, state.LLMProviderID),
		RemoteName:    updateStringPointer(plan.RemoteName, state.RemoteName),
	}

	model, err := r.client.UpdateModel(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update model", err.Error())
		return
	}

	updatedState := modelModel{
		ID:             types.StringValue(model.Meta.Id),
		OrganizationID: types.StringValue(plan.OrganizationID.ValueString()),
		Name:           types.StringValue(model.Name),
		LLMProviderID:  types.StringValue(model.LlmProviderId),
		RemoteName:     types.StringValue(model.RemoteName),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *modelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state modelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteModel(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete model", err.Error())
		return
	}
}

func (r *modelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
