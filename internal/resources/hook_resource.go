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

type hookResource struct {
	client *teamapi.Client
}

var _ resource.Resource = &hookResource{}
var _ resource.ResourceWithImportState = &hookResource{}

type hookModel struct {
	ID          types.String           `tfsdk:"id"`
	AgentID     types.String           `tfsdk:"agent_id"`
	Event       types.String           `tfsdk:"event"`
	Function    types.String           `tfsdk:"function"`
	Image       types.String           `tfsdk:"image"`
	Description types.String           `tfsdk:"description"`
	Resources   *computeResourcesModel `tfsdk:"resources"`
}

func NewHookResource() resource.Resource { return &hookResource{} }

func (r *hookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hook"
}

func (r *hookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn hook.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the hook.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"agent_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Agent identifier.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"event": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Hook event.",
			},
			"function": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Hook function.",
			},
			"image": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Container image.",
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

func (r *hookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *hookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan hookModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := teamapi.HookCreate{
		AgentID:     plan.AgentID.ValueString(),
		Event:       plan.Event.ValueString(),
		Function:    plan.Function.ValueString(),
		Image:       plan.Image.ValueString(),
		Description: stringPointer(plan.Description),
		Resources:   computeResourcesFromModel(plan.Resources),
	}

	hook, err := r.client.CreateHook(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create hook", err.Error())
		return
	}

	state := hookModel{
		ID:          types.StringValue(hook.ID),
		AgentID:     types.StringValue(hook.AgentID),
		Event:       types.StringValue(hook.Event),
		Function:    types.StringValue(hook.Function),
		Image:       types.StringValue(hook.Image),
		Description: optionalString(hook.Description),
		Resources:   computeResourcesToModel(hook.Resources),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *hookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state hookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hook, err := r.client.GetHook(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read hook", err.Error())
		return
	}

	state.AgentID = types.StringValue(hook.AgentID)
	state.Event = types.StringValue(hook.Event)
	state.Function = types.StringValue(hook.Function)
	state.Image = types.StringValue(hook.Image)
	state.Description = optionalString(hook.Description)
	state.Resources = computeResourcesToModel(hook.Resources)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *hookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan hookModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := teamapi.HookUpdate{
		Event:       stringPointer(plan.Event),
		Function:    stringPointer(plan.Function),
		Image:       stringPointer(plan.Image),
		Description: stringPointer(plan.Description),
		Resources:   computeResourcesFromModel(plan.Resources),
	}

	hook, err := r.client.UpdateHook(ctx, plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update hook", err.Error())
		return
	}

	state := hookModel{
		ID:          types.StringValue(hook.ID),
		AgentID:     types.StringValue(hook.AgentID),
		Event:       types.StringValue(hook.Event),
		Function:    types.StringValue(hook.Function),
		Image:       types.StringValue(hook.Image),
		Description: optionalString(hook.Description),
		Resources:   computeResourcesToModel(hook.Resources),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *hookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state hookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteHook(ctx, state.ID.ValueString()); err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			return
		}
		resp.Diagnostics.AddError("Unable to delete hook", err.Error())
		return
	}
}

func (r *hookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
