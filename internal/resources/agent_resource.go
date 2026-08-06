package resources

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/proto"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type agentResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &agentResource{}
var _ resource.ResourceWithImportState = &agentResource{}

type agentModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Nickname       types.String `tfsdk:"nickname"`
	Role           types.String `tfsdk:"role"`
	Model          types.String `tfsdk:"model"`
	Image          types.String `tfsdk:"image"`
	EnvironmentID  types.String `tfsdk:"environment_id"`
	Description    types.String `tfsdk:"description"`
	Configuration  types.String `tfsdk:"configuration"`
	IdleTimeout    types.String `tfsdk:"idle_timeout"`
	// Policies an instance of this agent inherits. See the agent-instances
	// change in agynio/architecture.
	DefaultThread   types.String           `tfsdk:"default_thread"`
	FinalMessage    types.String           `tfsdk:"final_message"`
	InstanceIdleTTL types.String           `tfsdk:"instance_idle_ttl"`
	Capabilities    types.List             `tfsdk:"capabilities"`
	Availability    types.String           `tfsdk:"availability"`
	Resources       *computeResourcesModel `tfsdk:"resources"`
}

var agentNicknameRegex = regexp.MustCompile("^[a-z0-9_-]+$")

const (
	agentAvailabilityInternal = "internal"
	agentAvailabilityPrivate  = "private"
)

const (
	agentDefaultThreadOrigin     = "origin"
	agentDefaultThreadNone       = "none"
	agentFinalMessageDiscard     = "discard"
	agentFinalMessageDefaultThrd = "default_thread"
)

func agentDefaultThreadToProto(v string) agentsv1.AgentDefaultThread {
	switch v {
	case agentDefaultThreadOrigin:
		return agentsv1.AgentDefaultThread_AGENT_DEFAULT_THREAD_ORIGIN
	case agentDefaultThreadNone:
		return agentsv1.AgentDefaultThread_AGENT_DEFAULT_THREAD_NONE
	default:
		return agentsv1.AgentDefaultThread_AGENT_DEFAULT_THREAD_UNSPECIFIED
	}
}

func agentDefaultThreadFromProto(v agentsv1.AgentDefaultThread) types.String {
	switch v {
	case agentsv1.AgentDefaultThread_AGENT_DEFAULT_THREAD_ORIGIN:
		return types.StringValue(agentDefaultThreadOrigin)
	case agentsv1.AgentDefaultThread_AGENT_DEFAULT_THREAD_NONE:
		return types.StringValue(agentDefaultThreadNone)
	default:
		return types.StringNull()
	}
}

func agentFinalMessageToProto(v string) agentsv1.AgentFinalMessage {
	switch v {
	case agentFinalMessageDiscard:
		return agentsv1.AgentFinalMessage_AGENT_FINAL_MESSAGE_DISCARD
	case agentFinalMessageDefaultThrd:
		return agentsv1.AgentFinalMessage_AGENT_FINAL_MESSAGE_DEFAULT_THREAD
	default:
		return agentsv1.AgentFinalMessage_AGENT_FINAL_MESSAGE_UNSPECIFIED
	}
}

func agentFinalMessageFromProto(v agentsv1.AgentFinalMessage) types.String {
	switch v {
	case agentsv1.AgentFinalMessage_AGENT_FINAL_MESSAGE_DISCARD:
		return types.StringValue(agentFinalMessageDiscard)
	case agentsv1.AgentFinalMessage_AGENT_FINAL_MESSAGE_DEFAULT_THREAD:
		return types.StringValue(agentFinalMessageDefaultThrd)
	default:
		return types.StringNull()
	}
}

func agentAvailabilityToProto(v string) agentsv1.AgentAvailability {
	switch v {
	case agentAvailabilityInternal:
		return agentsv1.AgentAvailability_AGENT_AVAILABILITY_INTERNAL
	case agentAvailabilityPrivate:
		return agentsv1.AgentAvailability_AGENT_AVAILABILITY_PRIVATE
	default:
		return agentsv1.AgentAvailability_AGENT_AVAILABILITY_UNSPECIFIED
	}
}

func agentAvailabilityFromProto(v agentsv1.AgentAvailability) types.String {
	switch v {
	case agentsv1.AgentAvailability_AGENT_AVAILABILITY_INTERNAL:
		return types.StringValue(agentAvailabilityInternal)
	case agentsv1.AgentAvailability_AGENT_AVAILABILITY_PRIVATE:
		return types.StringValue(agentAvailabilityPrivate)
	default:
		return types.StringNull()
	}
}

func NewAgentResource() resource.Resource { return &agentResource{} }

func (r *agentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

func (r *agentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	nicknameValidators := []validator.String{
		stringvalidator.LengthBetween(1, 32),
		stringvalidator.RegexMatches(agentNicknameRegex, "must contain only lowercase letters, numbers, underscores, or hyphens"),
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn agent.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the agent.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization identifier for the agent.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Agent name.",
			},
			"nickname": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional nickname for the agent.",
				Validators:          nicknameValidators,
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
			"environment_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Environment the agent runs in.",
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
			"idle_timeout": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Go duration string for idle timeout (for example, \"30s\", \"5m\", \"1h\").",
			},
			"default_thread": schema.StringAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "Where an instance's default thread comes from when the platform creates it: " +
					"\"origin\" takes the thread that added the instance, \"none\" infers nothing. Defaults to \"origin\".",
			},
			"final_message": schema.StringAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "What becomes of the text the agent CLI produces at the end of a turn: " +
					"\"discard\", or \"default_thread\" to post it. Defaults to \"discard\" -- an agent that " +
					"sends its own messages would otherwise post twice.",
			},
			"instance_idle_ttl": schema.StringAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "Go duration string. How long an instance of this agent may sit idle before " +
					"the platform pauses it. Unset means never.",
			},
			"capabilities": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Capabilities supported by this agent.",
			},
			"availability": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Agent availability. One of `internal` or `private`.",
				Validators: []validator.String{
					stringvalidator.OneOf(agentAvailabilityInternal, agentAvailabilityPrivate),
				},
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

	capabilities, diags := stringListFromPlan(ctx, plan.Capabilities)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &agentsv1.CreateAgentRequest{
		OrganizationId: plan.OrganizationID.ValueString(),
		Name:           plan.Name.ValueString(),
		Nickname:       stringValue(plan.Nickname),
		Role:           plan.Role.ValueString(),
		Model:          plan.Model.ValueString(),
		Image:          plan.Image.ValueString(),
		EnvironmentId:  plan.EnvironmentID.ValueString(),
		Description:    stringValue(plan.Description),
		Configuration:  stringValue(plan.Configuration),
		Capabilities:   capabilities,
		Availability:   agentAvailabilityToProto(plan.Availability.ValueString()),
		Resources:      computeResourcesFromModel(plan.Resources),
	}
	if !plan.IdleTimeout.IsNull() && !plan.IdleTimeout.IsUnknown() {
		input.IdleTimeout = proto.String(plan.IdleTimeout.ValueString())
	}
	if !plan.DefaultThread.IsNull() && !plan.DefaultThread.IsUnknown() {
		input.DefaultThread = agentDefaultThreadToProto(plan.DefaultThread.ValueString())
	}
	if !plan.FinalMessage.IsNull() && !plan.FinalMessage.IsUnknown() {
		input.FinalMessage = agentFinalMessageToProto(plan.FinalMessage.ValueString())
	}
	if !plan.InstanceIdleTTL.IsNull() && !plan.InstanceIdleTTL.IsUnknown() {
		input.InstanceIdleTtl = proto.String(plan.InstanceIdleTTL.ValueString())
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

	capabilitiesState, diags := stringListToState(ctx, agent.Capabilities, plan.Capabilities)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedState := agentModel{
		ID:              types.StringValue(agent.Meta.Id),
		OrganizationID:  types.StringValue(agent.OrganizationId),
		Name:            types.StringValue(agent.Name),
		Nickname:        optionalString(agent.Nickname),
		Role:            types.StringValue(agent.Role),
		Model:           types.StringValue(agent.Model),
		Image:           types.StringValue(agent.Image),
		EnvironmentID:   optionalString(agent.GetEnvironmentId()),
		Description:     optionalString(agent.Description),
		Configuration:   configuration,
		IdleTimeout:     optionalString(agent.GetIdleTimeout()),
		DefaultThread:   agentDefaultThreadFromProto(agent.GetDefaultThread()),
		FinalMessage:    agentFinalMessageFromProto(agent.GetFinalMessage()),
		InstanceIdleTTL: optionalString(agent.GetInstanceIdleTtl()),
		Capabilities:    capabilitiesState,
		Availability:    agentAvailabilityFromProto(agent.Availability),
		Resources:       computeResourcesToModel(agent.Resources),
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

	capabilitiesState, diags := stringListToState(ctx, agent.Capabilities, state.Capabilities)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Name = types.StringValue(agent.Name)
	state.Nickname = optionalString(agent.Nickname)
	state.Role = types.StringValue(agent.Role)
	state.Model = types.StringValue(agent.Model)
	state.Image = types.StringValue(agent.Image)
	state.EnvironmentID = optionalString(agent.GetEnvironmentId())
	state.OrganizationID = types.StringValue(agent.OrganizationId)
	state.Description = optionalString(agent.Description)
	state.Configuration = configuration
	state.IdleTimeout = optionalString(agent.GetIdleTimeout())
	state.DefaultThread = agentDefaultThreadFromProto(agent.GetDefaultThread())
	state.FinalMessage = agentFinalMessageFromProto(agent.GetFinalMessage())
	state.InstanceIdleTTL = optionalString(agent.GetInstanceIdleTtl())
	state.Capabilities = capabilitiesState
	state.Availability = agentAvailabilityFromProto(agent.Availability)
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

	capabilities, diags := stringListFromPlan(ctx, plan.Capabilities)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &agentsv1.UpdateAgentRequest{
		Id:              plan.ID.ValueString(),
		Name:            updateStringPointer(plan.Name, state.Name),
		Nickname:        updateStringPointer(plan.Nickname, state.Nickname),
		Role:            updateStringPointer(plan.Role, state.Role),
		Model:           updateStringPointer(plan.Model, state.Model),
		Image:           updateStringPointer(plan.Image, state.Image),
		EnvironmentId:   updateStringPointer(plan.EnvironmentID, state.EnvironmentID),
		Description:     updateStringPointer(plan.Description, state.Description),
		Configuration:   updateStringPointer(plan.Configuration, state.Configuration),
		IdleTimeout:     updateStringPointer(plan.IdleTimeout, state.IdleTimeout),
		InstanceIdleTtl: updateStringPointer(plan.InstanceIdleTTL, state.InstanceIdleTTL),
		Capabilities:    capabilities,
		Availability:    updateAgentAvailability(plan.Availability, state.Availability),
		Resources:       updateComputeResources(plan.Resources, state.Resources),
	}
	// Enums, so the string helper does not apply: send one only when the plan
	// names a value the state does not already hold.
	if !plan.DefaultThread.IsNull() && !plan.DefaultThread.IsUnknown() && !plan.DefaultThread.Equal(state.DefaultThread) {
		value := agentDefaultThreadToProto(plan.DefaultThread.ValueString())
		input.DefaultThread = &value
	}
	if !plan.FinalMessage.IsNull() && !plan.FinalMessage.IsUnknown() && !plan.FinalMessage.Equal(state.FinalMessage) {
		value := agentFinalMessageToProto(plan.FinalMessage.ValueString())
		input.FinalMessage = &value
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

	capabilitiesState, diags := stringListToState(ctx, agent.Capabilities, plan.Capabilities)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedState := agentModel{
		ID:              types.StringValue(agent.Meta.Id),
		OrganizationID:  types.StringValue(agent.OrganizationId),
		Name:            types.StringValue(agent.Name),
		Nickname:        optionalString(agent.Nickname),
		Role:            types.StringValue(agent.Role),
		Model:           types.StringValue(agent.Model),
		Image:           types.StringValue(agent.Image),
		EnvironmentID:   optionalString(agent.GetEnvironmentId()),
		Description:     optionalString(agent.Description),
		Configuration:   configuration,
		IdleTimeout:     optionalString(agent.GetIdleTimeout()),
		DefaultThread:   agentDefaultThreadFromProto(agent.GetDefaultThread()),
		FinalMessage:    agentFinalMessageFromProto(agent.GetFinalMessage()),
		InstanceIdleTTL: optionalString(agent.GetInstanceIdleTtl()),
		Capabilities:    capabilitiesState,
		Availability:    agentAvailabilityFromProto(agent.Availability),
		Resources:       computeResourcesToModel(agent.Resources),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func updateAgentAvailability(plan types.String, prior types.String) *agentsv1.AgentAvailability {
	if plan.IsUnknown() || plan.IsNull() {
		return nil
	}
	if !prior.IsNull() && !prior.IsUnknown() && plan.ValueString() == prior.ValueString() {
		return nil
	}
	v := agentAvailabilityToProto(plan.ValueString())
	return &v
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
