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

type mcpServerResource struct {
	client *teamapi.Client
}

type mcpServerModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Config      types.String `tfsdk:"config"`
}

func NewMCPServerResource() resource.Resource { return &mcpServerResource{} }

func (r *mcpServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_server"
}

func (r *mcpServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn MCP server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the MCP server.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"title": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable title of the MCP server.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable description of the MCP server.",
			},
			"config": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "JSON-encoded MCP server configuration. Use `jsonencode()` to construct the value.",
			},
		},
	}
}

func (r *mcpServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *mcpServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing MCP servers.")
		return
	}

	var plan mcpServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configValue := plan.Config.ValueString()
	if configValue == "" {
		resp.Diagnostics.AddAttributeError(path.Root("config"), "Missing Config", "MCP server config must be provided and cannot be empty.")
		return
	}
	if !json.Valid([]byte(configValue)) {
		resp.Diagnostics.AddAttributeError(path.Root("config"), "Invalid JSON", "MCP server config must be valid JSON.")
		return
	}

	create := teamapi.MCPServerCreate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
		Config:      json.RawMessage(configValue),
	}

	server, err := r.client.CreateMCPServer(ctx, create)
	if err != nil {
		resp.Diagnostics.AddError("Create MCP Server Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(server.ID)
	plan.Title = optionalString(server.Title)
	plan.Description = optionalString(server.Description)
	plan.Config = types.StringValue(string(server.Config))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mcpServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing MCP servers.")
		return
	}

	var state mcpServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.GetMCPServer(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read MCP Server Failed", err.Error())
		return
	}

	state.ID = types.StringValue(server.ID)
	state.Title = optionalString(server.Title)
	state.Description = optionalString(server.Description)
	state.Config = types.StringValue(string(server.Config))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *mcpServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing MCP servers.")
		return
	}

	var plan mcpServerModel
	var state mcpServerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	update := teamapi.MCPServerUpdate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
	}

	if !plan.Config.IsUnknown() && !plan.Config.IsNull() {
		configValue := plan.Config.ValueString()
		if configValue == "" || !json.Valid([]byte(configValue)) {
			resp.Diagnostics.AddAttributeError(path.Root("config"), "Invalid JSON", "MCP server config must be valid JSON when provided.")
			return
		}
		raw := json.RawMessage(configValue)
		update.Config = &raw
	}

	server, err := r.client.UpdateMCPServer(ctx, state.ID.ValueString(), update)
	if err != nil {
		resp.Diagnostics.AddError("Update MCP Server Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(server.ID)
	plan.Title = optionalString(server.Title)
	plan.Description = optionalString(server.Description)
	plan.Config = types.StringValue(string(server.Config))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mcpServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing MCP servers.")
		return
	}

	var state mcpServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteMCPServer(ctx, state.ID.ValueString()); err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			return
		}
		resp.Diagnostics.AddError("Delete MCP Server Failed", err.Error())
	}
}

func (r *mcpServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
