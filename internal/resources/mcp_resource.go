package resources

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-agyn/internal/teamapi"
)

type mcpResource struct {
	client *teamapi.Client
}

var _ resource.Resource = &mcpResource{}
var _ resource.ResourceWithImportState = &mcpResource{}

type mcpModel struct {
	ID          types.String           `tfsdk:"id"`
	AgentID     types.String           `tfsdk:"agent_id"`
	Image       types.String           `tfsdk:"image"`
	Command     types.String           `tfsdk:"command"`
	Description types.String           `tfsdk:"description"`
	Resources   *computeResourcesModel `tfsdk:"resources"`
}

func NewMcpResource() resource.Resource { return &mcpResource{} }

func (r *mcpResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp"
}

func (r *mcpResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn MCP.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the MCP.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"agent_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Agent identifier.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"image": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Container image.",
			},
			"command": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Command to execute.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable description.",
			},
			"resources": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Compute resource requests and limits.",
				Attributes:          computeResourcesSchemaAttributes(),
			},
		},
	}
}

func (r *mcpResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*teamapi.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *teamapi.Client")
		return
	}
	r.client = client
}

func (r *mcpResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan mcpModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := teamapi.McpCreate{
		AgentID:     plan.AgentID.ValueString(),
		Image:       plan.Image.ValueString(),
		Command:     plan.Command.ValueString(),
		Description: stringPointer(plan.Description),
		Resources:   computeResourcesFromModel(plan.Resources),
	}

	mcp, err := r.client.CreateMcp(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create MCP", err.Error())
		return
	}

	state := mcpModel{
		ID:          types.StringValue(mcp.ID),
		AgentID:     types.StringValue(mcp.AgentID),
		Image:       types.StringValue(mcp.Image),
		Command:     types.StringValue(mcp.Command),
		Description: optionalString(mcp.Description),
		Resources:   computeResourcesToModel(mcp.Resources),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *mcpResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state mcpModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mcp, err := r.client.GetMcp(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read MCP", err.Error())
		return
	}

	state.AgentID = types.StringValue(mcp.AgentID)
	state.Image = types.StringValue(mcp.Image)
	state.Command = types.StringValue(mcp.Command)
	state.Description = optionalString(mcp.Description)
	state.Resources = computeResourcesToModel(mcp.Resources)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *mcpResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan mcpModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := teamapi.McpUpdate{
		Image:       stringPointer(plan.Image),
		Command:     stringPointer(plan.Command),
		Description: stringPointer(plan.Description),
		Resources:   computeResourcesFromModel(plan.Resources),
	}

	mcp, err := r.client.UpdateMcp(ctx, plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update MCP", err.Error())
		return
	}

	state := mcpModel{
		ID:          types.StringValue(mcp.ID),
		AgentID:     types.StringValue(mcp.AgentID),
		Image:       types.StringValue(mcp.Image),
		Command:     types.StringValue(mcp.Command),
		Description: optionalString(mcp.Description),
		Resources:   computeResourcesToModel(mcp.Resources),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *mcpResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state mcpModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteMcp(ctx, state.ID.ValueString()); err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			return
		}
		resp.Diagnostics.AddError("Unable to delete MCP", err.Error())
		return
	}
}

func (r *mcpResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
