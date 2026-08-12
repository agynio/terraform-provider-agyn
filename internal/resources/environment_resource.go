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

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type environmentResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &environmentResource{}
var _ resource.ResourceWithImportState = &environmentResource{}

type environmentModel struct {
	ID                   types.String `tfsdk:"id"`
	OrganizationID       types.String `tfsdk:"organization_id"`
	Name                 types.String `tfsdk:"name"`
	RunnerID             types.String `tfsdk:"runner_id"`
	Flavor               types.String `tfsdk:"flavor"`
	WorkspaceImageID     types.String `tfsdk:"workspace_image_id"`
	WorkspaceImageTag    types.String `tfsdk:"workspace_image_tag"`
	AgentRuntimeImageID  types.String `tfsdk:"agent_runtime_image_id"`
	AgentRuntimeImageTag types.String `tfsdk:"agent_runtime_image_tag"`
	Availability         types.String `tfsdk:"availability"`
}

const (
	environmentAvailabilityInternal = "internal"
	environmentAvailabilityPrivate  = "private"
)

func NewEnvironmentResource() resource.Resource { return &environmentResource{} }

func (r *environmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *environmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn environment: a runner, a flavor name on that runner, and the images a workload runs.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the environment.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization that owns the environment.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Environment name, unique within the organization.",
			},
			"runner_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Runner workloads are placed on. Must be visible to the organization.",
			},
			"flavor": schema.StringAttribute{
				Optional: true,
				Computed: true,
				// Late-bound on purpose: an environment and the runner
				// configuration naming its flavor can be applied in either
				// order, so the name is not validated here.
				MarkdownDescription: "Flavor name in the runner's catalog. Resolved at workload start, not on write. Empty uses the runner's default.",
			},
			"workspace_image_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Catalog image of type `workspace`, run as the workload's main container.",
			},
			"workspace_image_tag": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Tag within that image. Checked against discovered versions on write, resolved again at each workload start.",
			},
			"agent_runtime_image_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				// Omitting it is what makes a workspace-only environment:
				// usable by a sandbox, rejected by agent creation.
				MarkdownDescription: "Catalog image of type `agent_runtime`, supplying the agent CLI. Omit for a workspace-only environment.",
			},
			"agent_runtime_image_tag": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Tag within the agent runtime image.",
			},
			"availability": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "internal or private. Controls who may run workloads in the environment; running in one reaches its secrets, egress credentials and volume contents.",
				Validators: []validator.String{
					stringvalidator.OneOf(environmentAvailabilityInternal, environmentAvailabilityPrivate),
				},
			},
		},
	}
}

func (r *environmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *environmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan environmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	environment, err := r.client.CreateEnvironment(ctx, &agentsv1.CreateEnvironmentRequest{
		OrganizationId:       plan.OrganizationID.ValueString(),
		Name:                 plan.Name.ValueString(),
		RunnerId:             plan.RunnerID.ValueString(),
		Flavor:               plan.Flavor.ValueString(),
		WorkspaceImageId:     plan.WorkspaceImageID.ValueString(),
		WorkspaceImageTag:    plan.WorkspaceImageTag.ValueString(),
		AgentRuntimeImageId:  plan.AgentRuntimeImageID.ValueString(),
		AgentRuntimeImageTag: plan.AgentRuntimeImageTag.ValueString(),
		Availability:         environmentAvailabilityFromString(plan.Availability.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create environment", err.Error())
		return
	}

	state := environmentStateFrom(environment)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *environmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state environmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	environment, err := r.client.GetEnvironment(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read environment", err.Error())
		return
	}

	state = environmentStateFrom(environment)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *environmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan, state environmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	environment, err := r.client.UpdateEnvironment(ctx, &agentsv1.UpdateEnvironmentRequest{
		Id:                   plan.ID.ValueString(),
		Name:                 updateStringPointer(plan.Name, state.Name),
		RunnerId:             updateStringPointer(plan.RunnerID, state.RunnerID),
		Flavor:               updateStringPointer(plan.Flavor, state.Flavor),
		WorkspaceImageId:     updateStringPointer(plan.WorkspaceImageID, state.WorkspaceImageID),
		WorkspaceImageTag:    updateStringPointer(plan.WorkspaceImageTag, state.WorkspaceImageTag),
		AgentRuntimeImageId:  updateStringPointer(plan.AgentRuntimeImageID, state.AgentRuntimeImageID),
		AgentRuntimeImageTag: updateStringPointer(plan.AgentRuntimeImageTag, state.AgentRuntimeImageTag),
		Availability:         environmentAvailabilityUpdate(plan.Availability),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to update environment", err.Error())
		return
	}

	updated := environmentStateFrom(environment)
	resp.Diagnostics.Append(resp.State.Set(ctx, &updated)...)
}

func (r *environmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state environmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Delete is rejected while any agent or sandbox references the
	// environment, which surfaces as a FailedPrecondition.
	if err := r.client.DeleteEnvironment(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete environment", err.Error())
	}
}

func (r *environmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func environmentStateFrom(environment *agentsv1.Environment) environmentModel {
	return environmentModel{
		ID:                   types.StringValue(environment.GetMeta().GetId()),
		OrganizationID:       types.StringValue(environment.GetOrganizationId()),
		Name:                 types.StringValue(environment.GetName()),
		RunnerID:             types.StringValue(environment.GetRunnerId()),
		Flavor:               types.StringValue(environment.GetFlavor()),
		WorkspaceImageID:     types.StringValue(environment.GetWorkspaceImageId()),
		WorkspaceImageTag:    types.StringValue(environment.GetWorkspaceImageTag()),
		AgentRuntimeImageID:  types.StringValue(environment.GetAgentRuntimeImageId()),
		AgentRuntimeImageTag: types.StringValue(environment.GetAgentRuntimeImageTag()),
		Availability:         types.StringValue(environmentAvailabilityToString(environment.GetAvailability())),
	}
}

func environmentAvailabilityFromString(value string) agentsv1.EnvironmentAvailability {
	if strings.EqualFold(strings.TrimSpace(value), environmentAvailabilityPrivate) {
		return agentsv1.EnvironmentAvailability_ENVIRONMENT_AVAILABILITY_PRIVATE
	}
	return agentsv1.EnvironmentAvailability_ENVIRONMENT_AVAILABILITY_INTERNAL
}

func environmentAvailabilityToString(value agentsv1.EnvironmentAvailability) string {
	if value == agentsv1.EnvironmentAvailability_ENVIRONMENT_AVAILABILITY_PRIVATE {
		return environmentAvailabilityPrivate
	}
	return environmentAvailabilityInternal
}

// The attribute is required, so an unknown plan value is the only case with
// nothing to send: leaving it out would silently keep the prior availability.
func environmentAvailabilityUpdate(plan types.String) *agentsv1.EnvironmentAvailability {
	if plan.IsNull() || plan.IsUnknown() {
		return nil
	}
	availability := environmentAvailabilityFromString(plan.ValueString())
	return &availability
}
