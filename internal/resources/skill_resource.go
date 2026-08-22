package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type skillResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &skillResource{}
var _ resource.ResourceWithImportState = &skillResource{}

type skillModel struct {
	ID          types.String `tfsdk:"id"`
	AgentID     types.String `tfsdk:"agent_id"`
	Name        types.String `tfsdk:"name"`
	Body        types.String `tfsdk:"body"`
	Description types.String `tfsdk:"description"`
}

func NewSkillResource() resource.Resource { return &skillResource{} }

func (r *skillResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_skill"
}

func (r *skillResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn skill.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the skill.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"agent_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Agent identifier.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Skill name. A slug: lowercase letters, digits and hyphens, up to 64 characters. Names the skill's directory on the agent filesystem.",
			},
			"body": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Skill body. Markdown instructions, without front matter.",
			},
			"description": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "When the skill applies. The agent reads this before opening the body.",
			},
		},
	}
}

func (r *skillResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *skillResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan skillModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &agentsv1.CreateSkillRequest{
		AgentId:     plan.AgentID.ValueString(),
		Name:        plan.Name.ValueString(),
		Body:        plan.Body.ValueString(),
		Description: stringValue(plan.Description),
	}

	skill, err := r.client.CreateSkill(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create skill", err.Error())
		return
	}

	updatedState := skillModel{
		ID:          types.StringValue(skill.Meta.Id),
		AgentID:     types.StringValue(skill.AgentId),
		Name:        types.StringValue(skill.Name),
		Body:        types.StringValue(skill.Body),
		Description: optionalString(skill.Description),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *skillResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state skillModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	skill, err := r.client.GetSkill(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read skill", err.Error())
		return
	}

	state.AgentID = types.StringValue(skill.AgentId)
	state.Name = types.StringValue(skill.Name)
	state.Body = types.StringValue(skill.Body)
	state.Description = optionalString(skill.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *skillResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan skillModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state skillModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &agentsv1.UpdateSkillRequest{
		Id:          plan.ID.ValueString(),
		Name:        updateStringPointer(plan.Name, state.Name),
		Body:        updateStringPointer(plan.Body, state.Body),
		Description: updateStringPointer(plan.Description, state.Description),
	}

	skill, err := r.client.UpdateSkill(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update skill", err.Error())
		return
	}

	updatedState := skillModel{
		ID:          types.StringValue(skill.Meta.Id),
		AgentID:     types.StringValue(skill.AgentId),
		Name:        types.StringValue(skill.Name),
		Body:        types.StringValue(skill.Body),
		Description: optionalString(skill.Description),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *skillResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state skillModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSkill(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete skill", err.Error())
		return
	}
}

func (r *skillResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
