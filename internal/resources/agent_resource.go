package resources

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type agentResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &agentResource{}
var _ resource.ResourceWithImportState = &agentResource{}

type agentModel struct {
	ID            types.String           `tfsdk:"id"`
	Name          types.String           `tfsdk:"name"`
	Role          types.String           `tfsdk:"role"`
	Model         types.String           `tfsdk:"model"`
	Image         types.String           `tfsdk:"image"`
	Description   types.String           `tfsdk:"description"`
	Configuration types.String           `tfsdk:"configuration"`
	Resources     *computeResourcesModel `tfsdk:"resources"`
}

func NewAgentResource() resource.Resource { return &agentResource{} }

func (r *agentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

func (r *agentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn agent.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the agent.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Agent name.",
			},
			"role": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Agent role.",
			},
			"model": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Model identifier.",
			},
			"image": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Container image.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable description.",
			},
			"configuration": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "JSON-encoded agent configuration.",
			},
			"resources": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Compute resource requests and limits.",
				Attributes:          computeResourcesSchemaAttributes(),
			},
		},
	}
}

func (r *agentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *agentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan agentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Configuration.IsNull() && !plan.Configuration.IsUnknown() {
		if !json.Valid([]byte(plan.Configuration.ValueString())) {
			resp.Diagnostics.AddAttributeError(path.Root("configuration"), "Invalid JSON", "configuration must be valid JSON")
			return
		}
	}

	input := &agentsv1.CreateAgentRequest{
		Name:          plan.Name.ValueString(),
		Role:          plan.Role.ValueString(),
		Model:         plan.Model.ValueString(),
		Image:         plan.Image.ValueString(),
		Description:   stringValue(plan.Description),
		Configuration: stringValue(plan.Configuration),
		Resources:     computeResourcesFromModel(plan.Resources),
	}

	agent, err := r.client.CreateAgent(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create agent", err.Error())
		return
	}

	configuration, diags := normalizeJSONState(plan.Configuration, agent.Configuration)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedState := agentModel{
		ID:            types.StringValue(agent.Meta.Id),
		Name:          types.StringValue(agent.Name),
		Role:          types.StringValue(agent.Role),
		Model:         types.StringValue(agent.Model),
		Image:         types.StringValue(agent.Image),
		Description:   optionalString(agent.Description),
		Configuration: configuration,
		Resources:     computeResourcesToModel(agent.Resources),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *agentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state agentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	agent, err := r.client.GetAgent(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read agent", err.Error())
		return
	}

	configuration, diags := normalizeJSONState(state.Configuration, agent.Configuration)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Name = types.StringValue(agent.Name)
	state.Role = types.StringValue(agent.Role)
	state.Model = types.StringValue(agent.Model)
	state.Image = types.StringValue(agent.Image)
	state.Description = optionalString(agent.Description)
	state.Configuration = configuration
	state.Resources = computeResourcesToModel(agent.Resources)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *agentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan agentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state agentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Configuration.IsNull() && !plan.Configuration.IsUnknown() {
		if !json.Valid([]byte(plan.Configuration.ValueString())) {
			resp.Diagnostics.AddAttributeError(path.Root("configuration"), "Invalid JSON", "configuration must be valid JSON")
			return
		}
	}

	input := &agentsv1.UpdateAgentRequest{
		Id:            plan.ID.ValueString(),
		Name:          updateStringPointer(plan.Name, state.Name),
		Role:          updateStringPointer(plan.Role, state.Role),
		Model:         updateStringPointer(plan.Model, state.Model),
		Image:         updateStringPointer(plan.Image, state.Image),
		Description:   updateStringPointer(plan.Description, state.Description),
		Configuration: updateStringPointer(plan.Configuration, state.Configuration),
		Resources:     updateComputeResources(plan.Resources, state.Resources),
	}

	agent, err := r.client.UpdateAgent(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update agent", err.Error())
		return
	}

	configuration, diags := normalizeJSONState(plan.Configuration, agent.Configuration)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedState := agentModel{
		ID:            types.StringValue(agent.Meta.Id),
		Name:          types.StringValue(agent.Name),
		Role:          types.StringValue(agent.Role),
		Model:         types.StringValue(agent.Model),
		Image:         types.StringValue(agent.Image),
		Description:   optionalString(agent.Description),
		Configuration: configuration,
		Resources:     computeResourcesToModel(agent.Resources),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *agentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state agentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAgent(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete agent", err.Error())
		return
	}
}

func (r *agentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
