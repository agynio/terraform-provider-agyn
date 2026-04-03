package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/proto"

	runnersv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/runners/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type runnerResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &runnerResource{}
var _ resource.ResourceWithImportState = &runnerResource{}

type runnerModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Labels         types.Map    `tfsdk:"labels"`
	IdentityID     types.String `tfsdk:"identity_id"`
	ServiceToken   types.String `tfsdk:"service_token"`
}

func NewRunnerResource() resource.Resource { return &runnerResource{} }

func (r *runnerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runner"
}

func (r *runnerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn runner.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the runner.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Runner name.",
			},
			"organization_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Organization identifier for the runner.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"labels": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Runner labels.",
			},
			"identity_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identity identifier for the runner.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"service_token": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Service token for the runner.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *runnerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *runnerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan runnerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	labels, diags := runnerLabelsFromPlan(ctx, plan.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &runnersv1.RegisterRunnerRequest{
		Name:   plan.Name.ValueString(),
		Labels: labels,
	}
	if !plan.OrganizationID.IsNull() && !plan.OrganizationID.IsUnknown() {
		input.OrganizationId = proto.String(plan.OrganizationID.ValueString())
	}

	result, err := r.client.RegisterRunner(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to register runner", err.Error())
		return
	}

	updatedState, diags := runnerToState(result.Runner, optionalString(result.ServiceToken))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *runnerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state runnerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	runner, err := r.client.GetRunner(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read runner", err.Error())
		return
	}

	updatedState, diags := runnerToState(runner, preserveSensitiveString(state.ServiceToken, ""))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *runnerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan runnerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state runnerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	labels, diags := runnerLabelsFromPlan(ctx, plan.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &runnersv1.UpdateRunnerRequest{
		Id:     plan.ID.ValueString(),
		Name:   updateStringPointer(plan.Name, state.Name),
		Labels: labels,
	}

	runner, err := r.client.UpdateRunner(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update runner", err.Error())
		return
	}

	updatedState, diags := runnerToState(runner, preserveSensitiveString(state.ServiceToken, ""))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *runnerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state runnerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteRunner(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete runner", err.Error())
		return
	}
}

func (r *runnerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func runnerToState(runner *runnersv1.Runner, serviceToken types.String) (runnerModel, diag.Diagnostics) {
	labels, diags := runnerLabelsToState(runner.Labels)
	return runnerModel{
		ID:             types.StringValue(runner.Meta.Id),
		Name:           types.StringValue(runner.Name),
		OrganizationID: optionalString(runner.GetOrganizationId()),
		Labels:         labels,
		IdentityID:     optionalString(runner.IdentityId),
		ServiceToken:   serviceToken,
	}, diags
}

func runnerLabelsToState(labels map[string]string) (types.Map, diag.Diagnostics) {
	if len(labels) == 0 {
		return types.MapNull(types.StringType), nil
	}
	values := make(map[string]attr.Value, len(labels))
	for key, value := range labels {
		values[key] = types.StringValue(value)
	}
	return types.MapValue(types.StringType, values)
}

func runnerLabelsFromPlan(ctx context.Context, labels types.Map) (map[string]string, diag.Diagnostics) {
	if labels.IsNull() || labels.IsUnknown() {
		return map[string]string{}, nil
	}
	values := map[string]string{}
	diags := labels.ElementsAs(ctx, &values, false)
	if values == nil {
		values = map[string]string{}
	}
	return values, diags
}
