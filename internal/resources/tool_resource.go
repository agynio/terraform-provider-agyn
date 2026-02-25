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

type toolResource struct {
	client *teamapi.Client
}

type toolModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Config      types.String `tfsdk:"config"`
}

func NewToolResource() resource.Resource { return &toolResource{} }

func (r *toolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tool"
}

func (r *toolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name":        schema.StringAttribute{Optional: true},
			"description": schema.StringAttribute{Optional: true},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Tool type identifier.",
			},
			"config": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional JSON-encoded tool configuration.",
			},
		},
	}
}

func (r *toolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *toolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing tools.")
		return
	}

	var plan toolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	typeValue := plan.Type.ValueString()
	if typeValue == "" {
		resp.Diagnostics.AddAttributeError(path.Root("type"), "Missing Type", "Tool type must be provided.")
		return
	}

	create := teamapi.ToolCreate{
		Type:        typeValue,
		Name:        stringPointer(plan.Name),
		Description: stringPointer(plan.Description),
	}

	if !plan.Config.IsNull() && !plan.Config.IsUnknown() {
		configValue := plan.Config.ValueString()
		if configValue != "" && !json.Valid([]byte(configValue)) {
			resp.Diagnostics.AddAttributeError(path.Root("config"), "Invalid JSON", "Tool config must be valid JSON when provided.")
			return
		}
		if configValue != "" {
			raw := json.RawMessage(configValue)
			create.Config = &raw
		}
	}

	tool, err := r.client.CreateTool(ctx, create)
	if err != nil {
		resp.Diagnostics.AddError("Create Tool Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(tool.ID)
	plan.Name = optionalString(tool.Name)
	plan.Description = optionalString(tool.Description)
	plan.Type = types.StringValue(tool.Type)
	if tool.Config != nil {
		plan.Config = types.StringValue(string(*tool.Config))
	} else {
		plan.Config = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *toolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing tools.")
		return
	}

	var state toolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tool, err := r.client.GetTool(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Tool Failed", err.Error())
		return
	}

	state.ID = types.StringValue(tool.ID)
	state.Name = optionalString(tool.Name)
	state.Description = optionalString(tool.Description)
	state.Type = types.StringValue(tool.Type)
	if tool.Config != nil {
		state.Config = types.StringValue(string(*tool.Config))
	} else {
		state.Config = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *toolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing tools.")
		return
	}

	var plan toolModel
	var state toolModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	update := teamapi.ToolUpdate{
		Name:        stringPointer(plan.Name),
		Description: stringPointer(plan.Description),
	}

	if !plan.Config.IsNull() && !plan.Config.IsUnknown() {
		configValue := plan.Config.ValueString()
		if configValue != "" && !json.Valid([]byte(configValue)) {
			resp.Diagnostics.AddAttributeError(path.Root("config"), "Invalid JSON", "Tool config must be valid JSON when provided.")
			return
		}
		if configValue != "" {
			raw := json.RawMessage(configValue)
			update.Config = &raw
		}
	}

	tool, err := r.client.UpdateTool(ctx, state.ID.ValueString(), update)
	if err != nil {
		resp.Diagnostics.AddError("Update Tool Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(tool.ID)
	plan.Name = optionalString(tool.Name)
	plan.Description = optionalString(tool.Description)
	plan.Type = types.StringValue(tool.Type)
	if tool.Config != nil {
		plan.Config = types.StringValue(string(*tool.Config))
	} else {
		plan.Config = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *toolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing tools.")
		return
	}

	var state toolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTool(ctx, state.ID.ValueString()); err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			return
		}
		resp.Diagnostics.AddError("Delete Tool Failed", err.Error())
	}
}

func (r *toolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
