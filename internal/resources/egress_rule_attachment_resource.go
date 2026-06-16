package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	egressv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/egress/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type egressRuleAttachmentResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &egressRuleAttachmentResource{}

type egressRuleAttachmentModel struct {
	ID      types.String `tfsdk:"id"`
	RuleID  types.String `tfsdk:"rule_id"`
	AgentID types.String `tfsdk:"agent_id"`
}

func NewEgressRuleAttachmentResource() resource.Resource { return &egressRuleAttachmentResource{} }

func (r *egressRuleAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_egress_rule_attachment"
}

func (r *egressRuleAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches an Agyn egress rule to an agent.",
		Attributes: map[string]schema.Attribute{
			"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"rule_id":  schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"agent_id": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		},
	}
}

func (r *egressRuleAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *egressRuleAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan egressRuleAttachmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	attachment, err := r.client.CreateEgressRuleAttachment(ctx, &egressv1.CreateEgressRuleAttachmentRequest{RuleId: plan.RuleID.ValueString(), AgentId: plan.AgentID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create egress rule attachment", err.Error())
		return
	}
	plan.ID = types.StringValue(attachment.GetMeta().GetId())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *egressRuleAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state egressRuleAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *egressRuleAttachmentResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Egress rule attachments are immutable. This is an internal error.")
}

func (r *egressRuleAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state egressRuleAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteEgressRuleAttachment(ctx, state.ID.ValueString()); err != nil && !agentapi.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete egress rule attachment", err.Error())
	}
}
