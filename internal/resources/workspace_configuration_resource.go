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

type workspaceConfigurationResource struct {
	client *teamapi.Client
}

type workspaceConfigurationModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Config      types.String `tfsdk:"config"`
}

func NewWorkspaceConfigurationResource() resource.Resource { return &workspaceConfigurationResource{} }

func (r *workspaceConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_configuration"
}

func (r *workspaceConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				MarkdownDescription: "JSON-encoded workspace configuration.",
			},
		},
	}
}

func (r *workspaceConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *workspaceConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing workspace configurations.")
		return
	}

	var plan workspaceConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configValue := plan.Config.ValueString()
	if configValue == "" {
		resp.Diagnostics.AddAttributeError(path.Root("config"), "Missing Config", "Workspace configuration config must be provided and cannot be empty.")
		return
	}
	if !json.Valid([]byte(configValue)) {
		resp.Diagnostics.AddAttributeError(path.Root("config"), "Invalid JSON", "Workspace configuration config must be valid JSON.")
		return
	}

	create := teamapi.WorkspaceConfigurationCreate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
		Config:      json.RawMessage(configValue),
	}

	cfg, err := r.client.CreateWorkspaceConfiguration(ctx, create)
	if err != nil {
		resp.Diagnostics.AddError("Create Workspace Configuration Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(cfg.ID)
	plan.Title = optionalString(cfg.Title)
	plan.Description = optionalString(cfg.Description)
	plan.Config = types.StringValue(string(cfg.Config))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workspaceConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing workspace configurations.")
		return
	}

	var state workspaceConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetWorkspaceConfiguration(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Workspace Configuration Failed", err.Error())
		return
	}

	state.ID = types.StringValue(cfg.ID)
	state.Title = optionalString(cfg.Title)
	state.Description = optionalString(cfg.Description)
	state.Config = types.StringValue(string(cfg.Config))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workspaceConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing workspace configurations.")
		return
	}

	var plan workspaceConfigurationModel
	var state workspaceConfigurationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	update := teamapi.WorkspaceConfigurationUpdate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
	}

	if !plan.Config.IsUnknown() && !plan.Config.IsNull() {
		configValue := plan.Config.ValueString()
		if configValue == "" || !json.Valid([]byte(configValue)) {
			resp.Diagnostics.AddAttributeError(path.Root("config"), "Invalid JSON", "Workspace configuration config must be valid JSON when provided.")
			return
		}
		raw := json.RawMessage(configValue)
		update.Config = &raw
	}

	cfg, err := r.client.UpdateWorkspaceConfiguration(ctx, state.ID.ValueString(), update)
	if err != nil {
		resp.Diagnostics.AddError("Update Workspace Configuration Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(cfg.ID)
	plan.Title = optionalString(cfg.Title)
	plan.Description = optionalString(cfg.Description)
	plan.Config = types.StringValue(string(cfg.Config))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workspaceConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing workspace configurations.")
		return
	}

	var state workspaceConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteWorkspaceConfiguration(ctx, state.ID.ValueString()); err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			return
		}
		resp.Diagnostics.AddError("Delete Workspace Configuration Failed", err.Error())
	}
}

func (r *workspaceConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
