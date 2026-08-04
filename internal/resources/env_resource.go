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

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type envResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &envResource{}
var _ resource.ResourceWithImportState = &envResource{}

type envModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	AgentID     types.String `tfsdk:"agent_id"`
	McpID       types.String `tfsdk:"mcp_id"`
	Value       types.String `tfsdk:"value"`
	SecretID    types.String `tfsdk:"secret_id"`
}

func NewEnvResource() resource.Resource { return &envResource{} }

func (r *envResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env"
}

func (r *envResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	ownerValidators := []validator.String{
		stringvalidator.ExactlyOneOf(
			path.MatchRoot("agent_id"),
			path.MatchRoot("mcp_id"),
		),
	}
	valueValidators := []validator.String{
		stringvalidator.ExactlyOneOf(
			path.MatchRoot("value"),
			path.MatchRoot("secret_id"),
		),
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn environment variable.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the environment variable.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Environment variable name.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable description.",
			},
			"agent_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Target agent identifier.",
				Validators:          ownerValidators,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"mcp_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Target MCP identifier.",
				Validators:          ownerValidators,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"value": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Plain-text value.",
				Validators:          valueValidators,
			},
			"secret_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Secret reference identifier.",
				Validators:          valueValidators,
			},
		},
	}
}

func (r *envResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *envResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan envModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &agentsv1.CreateEnvRequest{
		Name:        plan.Name.ValueString(),
		Description: stringValue(plan.Description),
	}
	if setEnvTarget(input, plan.AgentID, plan.McpID, "create env", resp) {
		return
	}
	if setEnvSource(input, plan.Value, plan.SecretID, "create env", resp) {
		return
	}

	env, err := r.client.CreateEnv(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create env", err.Error())
		return
	}

	agentID, mcpID := envTargetState(env)
	value, secretID := envSourceState(env, plan.Value)

	updatedState := envModel{
		ID:          types.StringValue(env.Meta.Id),
		Name:        types.StringValue(env.Name),
		Description: optionalString(env.Description),
		AgentID:     agentID,
		McpID:       mcpID,
		Value:       value,
		SecretID:    secretID,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *envResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state envModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := r.client.GetEnv(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read env", err.Error())
		return
	}

	agentID, mcpID := envTargetState(env)
	value, secretID := envSourceState(env, state.Value)

	state.Name = types.StringValue(env.Name)
	state.Description = optionalString(env.Description)
	state.AgentID = agentID
	state.McpID = mcpID
	state.Value = value
	state.SecretID = secretID

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *envResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan envModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state envModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &agentsv1.UpdateEnvRequest{
		Id:          plan.ID.ValueString(),
		Name:        updateStringPointer(plan.Name, state.Name),
		Description: updateStringPointer(plan.Description, state.Description),
		Value:       updateStringPointer(plan.Value, state.Value),
		SecretId:    updateStringPointer(plan.SecretID, state.SecretID),
	}

	env, err := r.client.UpdateEnv(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update env", err.Error())
		return
	}

	agentID, mcpID := envTargetState(env)
	value, secretID := envSourceState(env, plan.Value)

	updatedState := envModel{
		ID:          types.StringValue(env.Meta.Id),
		Name:        types.StringValue(env.Name),
		Description: optionalString(env.Description),
		AgentID:     agentID,
		McpID:       mcpID,
		Value:       value,
		SecretID:    secretID,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *envResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state envModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteEnv(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete env", err.Error())
		return
	}
}

func (r *envResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func setEnvTarget(req *agentsv1.CreateEnvRequest, agentID types.String, mcpID types.String, op string, resp *resource.CreateResponse) bool {
	if !agentID.IsNull() && !agentID.IsUnknown() {
		req.Target = &agentsv1.CreateEnvRequest_AgentId{AgentId: agentID.ValueString()}
		return false
	}
	if !mcpID.IsNull() && !mcpID.IsUnknown() {
		req.Target = &agentsv1.CreateEnvRequest_McpId{McpId: mcpID.ValueString()}
		return false
	}
	resp.Diagnostics.AddError("Missing env target", op+" requires one of agent_id or mcp_id")
	return true
}

func setEnvSource(req *agentsv1.CreateEnvRequest, value types.String, secretID types.String, op string, resp *resource.CreateResponse) bool {
	if !value.IsNull() && !value.IsUnknown() {
		req.Source = &agentsv1.CreateEnvRequest_Value{Value: value.ValueString()}
		return false
	}
	if !secretID.IsNull() && !secretID.IsUnknown() {
		req.Source = &agentsv1.CreateEnvRequest_SecretId{SecretId: secretID.ValueString()}
		return false
	}
	resp.Diagnostics.AddError("Missing env value", op+" requires value or secret_id")
	return true
}

func envTargetState(env *agentsv1.Env) (types.String, types.String) {
	agentID := types.StringNull()
	mcpID := types.StringNull()
	switch target := env.GetTarget().(type) {
	case *agentsv1.Env_AgentId:
		agentID = types.StringValue(target.AgentId)
	case *agentsv1.Env_McpId:
		mcpID = types.StringValue(target.McpId)
	default:
		panic(fmt.Sprintf("unexpected env target type: %T", target))
	}
	return agentID, mcpID
}

func envSourceState(env *agentsv1.Env, fallback types.String) (types.String, types.String) {
	value := types.StringNull()
	secretID := types.StringNull()
	switch source := env.GetSource().(type) {
	case *agentsv1.Env_Value:
		value = preserveSensitiveString(fallback, source.Value)
	case *agentsv1.Env_SecretId:
		secretID = types.StringValue(source.SecretId)
	default:
		panic(fmt.Sprintf("unexpected env source type: %T", source))
	}
	return value, secretID
}
