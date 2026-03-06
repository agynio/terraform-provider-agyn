package resources

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-agyn/internal/teamapi"
)

type agentResource struct {
	client *teamapi.Client
}

type agentModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Config      types.String `tfsdk:"config"`
}

func NewAgentResource() resource.Resource { return &agentResource{} }

func (r *agentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

func (r *agentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"title":       schema.StringAttribute{Optional: true},
			"description": schema.StringAttribute{Optional: true},
			"config": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "JSON-encoded agent configuration.",
			},
		},
	}
}

func (r *agentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*teamapi.Client)
	if !ok || client == nil {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Unable to obtain configured API client")
		return
	}
	r.client = client
}

func (r *agentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing agents.")
		return
	}

	var plan agentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configValue := plan.Config.ValueString()
	if configValue == "" {
		resp.Diagnostics.AddAttributeError(path.Root("config"), "Missing Config", "Agent config must be provided and cannot be empty.")
		return
	}
	if !json.Valid([]byte(configValue)) {
		resp.Diagnostics.AddAttributeError(path.Root("config"), "Invalid JSON", "Agent config must be valid JSON.")
		return
	}

	create := teamapi.AgentCreate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
		Config:      json.RawMessage(configValue),
	}

	agent, err := r.client.CreateAgent(ctx, create)
	if err != nil {
		resp.Diagnostics.AddError("Create Agent Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(agent.ID)
	plan.Title = optionalString(agent.Title)
	plan.Description = optionalString(agent.Description)
	plan.Config = types.StringValue(configValue)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *agentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing agents.")
		return
	}

	var state agentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	agent, err := r.client.GetAgent(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Agent Failed", err.Error())
		return
	}

	state.ID = types.StringValue(agent.ID)
	state.Title = optionalString(agent.Title)
	state.Description = optionalString(agent.Description)
	configValue := state.Config
	if configValue.IsNull() || configValue.IsUnknown() || configValue.ValueString() == "" {
		configValue = types.StringValue(string(agent.Config))
	}
	state.Config = configValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *agentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing agents.")
		return
	}

	var plan agentModel
	var state agentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	update := teamapi.AgentUpdate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
	}

	if !plan.Config.IsUnknown() && !plan.Config.IsNull() {
		configValue := plan.Config.ValueString()
		if configValue == "" || !json.Valid([]byte(configValue)) {
			resp.Diagnostics.AddAttributeError(path.Root("config"), "Invalid JSON", "Agent config must be valid JSON when provided.")
			return
		}
		raw := json.RawMessage(configValue)
		update.Config = &raw
	}

	agent, err := r.client.UpdateAgent(ctx, state.ID.ValueString(), update)
	if err != nil {
		resp.Diagnostics.AddError("Update Agent Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(agent.ID)
	plan.Title = optionalString(agent.Title)
	plan.Description = optionalString(agent.Description)
	configValue := plan.Config
	if configValue.IsNull() || configValue.IsUnknown() || configValue.ValueString() == "" {
		configValue = state.Config
	}
	plan.Config = configValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *agentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing agents.")
		return
	}

	var state agentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAgent(ctx, state.ID.ValueString()); err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			return
		}
		resp.Diagnostics.AddError("Delete Agent Failed", err.Error())
	}
}

func (r *agentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
