package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	networksv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/networks/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type privateResourceResource struct{ client *agentapi.Client }

var _ resource.Resource = &privateResourceResource{}
var _ resource.ResourceWithImportState = &privateResourceResource{}

type privateResourceModel struct {
	ID                types.String `tfsdk:"id"`
	OrganizationID    types.String `tfsdk:"organization_id"`
	NetworkID         types.String `tfsdk:"network_id"`
	Name              types.String `tfsdk:"name"`
	Protocol          types.String `tfsdk:"protocol"`
	TargetHost        types.String `tfsdk:"target_host"`
	TargetPorts       types.List   `tfsdk:"target_ports"`
	InterceptHost     types.String `tfsdk:"intercept_host"`
	InterceptPorts    types.List   `tfsdk:"intercept_ports"`
	ProvisioningState types.String `tfsdk:"provisioning_state"`
}

func NewPrivateResourceResource() resource.Resource { return &privateResourceResource{} }

func (r *privateResourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_resource"
}

func (r *privateResourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	portValidators := []validator.List{listvalidator.ValueInt32sAre(int32validator.Between(1, 65535))}
	resp.Schema = schema.Schema{MarkdownDescription: "Manages an Agyn private resource exposed through a private network.", Attributes: map[string]schema.Attribute{
		"id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "UUID identifier of the private resource.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"organization_id":    schema.StringAttribute{Computed: true, MarkdownDescription: "Organization identifier for the private resource.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"network_id":         schema.StringAttribute{Required: true, MarkdownDescription: "Private network identifier.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"name":               schema.StringAttribute{Required: true, MarkdownDescription: "Private resource name."},
		"protocol":           schema.StringAttribute{Required: true, MarkdownDescription: "Private resource protocol. One of `tcp`, `http`, or `https`.", Validators: []validator.String{stringvalidator.OneOf("tcp", "http", "https")}},
		"target_host":        schema.StringAttribute{Required: true, MarkdownDescription: "Private target host reached by the operator-run tunneler."},
		"target_ports":       schema.ListAttribute{Required: true, ElementType: types.Int32Type, MarkdownDescription: "Private target ports.", Validators: portValidators, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
		"intercept_host":     schema.StringAttribute{Required: true, MarkdownDescription: "Hostname agents use to reach the private resource."},
		"intercept_ports":    schema.ListAttribute{Required: true, ElementType: types.Int32Type, MarkdownDescription: "Intercept ports agents use to reach the private resource.", Validators: portValidators, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
		"provisioning_state": schema.StringAttribute{Computed: true, MarkdownDescription: "OpenZiti provisioning state.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func (r *privateResourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *privateResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan privateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input, err := createPrivateResourceRequest(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid private resource", err.Error())
		return
	}
	privateResource, err := r.client.CreatePrivateResource(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create private resource", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, privateResourceState(privateResource))...)
}

func (r *privateResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state privateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	privateResource, err := r.client.GetPrivateResource(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read private resource", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, privateResourceState(privateResource))...)
}

func (r *privateResourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan privateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input, err := updatePrivateResourceRequest(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid private resource", err.Error())
		return
	}
	privateResource, err := r.client.UpdatePrivateResource(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update private resource", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, privateResourceState(privateResource))...)
}

func (r *privateResourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state privateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeletePrivateResource(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete private resource", err.Error())
	}
}

func (r *privateResourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func createPrivateResourceRequest(plan privateResourceModel) (*networksv1.CreatePrivateResourceRequest, error) {
	protocol, err := privateResourceProtocolFromString(plan.Protocol.ValueString())
	if err != nil {
		return nil, err
	}
	targetPorts, err := int32ListFromPlan(plan.TargetPorts)
	if err != nil {
		return nil, err
	}
	interceptPorts, err := int32ListFromPlan(plan.InterceptPorts)
	if err != nil {
		return nil, err
	}
	return &networksv1.CreatePrivateResourceRequest{NetworkId: plan.NetworkID.ValueString(), Name: plan.Name.ValueString(), Protocol: protocol, TargetHost: plan.TargetHost.ValueString(), TargetPorts: targetPorts, InterceptHost: plan.InterceptHost.ValueString(), InterceptPorts: interceptPorts}, nil
}

func updatePrivateResourceRequest(plan privateResourceModel) (*networksv1.UpdatePrivateResourceRequest, error) {
	protocol, err := privateResourceProtocolFromString(plan.Protocol.ValueString())
	if err != nil {
		return nil, err
	}
	targetPorts, err := int32ListFromPlan(plan.TargetPorts)
	if err != nil {
		return nil, err
	}
	interceptPorts, err := int32ListFromPlan(plan.InterceptPorts)
	if err != nil {
		return nil, err
	}
	return &networksv1.UpdatePrivateResourceRequest{Id: plan.ID.ValueString(), Name: stringPtr(plan.Name.ValueString()), Protocol: protocol.Enum(), TargetHost: stringPtr(plan.TargetHost.ValueString()), TargetPortsUpdate: &networksv1.PortListUpdate{Ports: targetPorts}, InterceptHost: stringPtr(plan.InterceptHost.ValueString()), InterceptPortsUpdate: &networksv1.PortListUpdate{Ports: interceptPorts}}, nil
}

func privateResourceState(privateResource *networksv1.PrivateResource) privateResourceModel {
	return privateResourceModel{ID: types.StringValue(privateResource.GetMeta().GetId()), OrganizationID: types.StringValue(privateResource.GetOrganizationId()), NetworkID: types.StringValue(privateResource.GetNetworkId()), Name: types.StringValue(privateResource.GetName()), Protocol: types.StringValue(privateResourceProtocolToString(privateResource.GetProtocol())), TargetHost: types.StringValue(privateResource.GetTargetHost()), TargetPorts: int32ListState(privateResource.GetTargetPorts()), InterceptHost: types.StringValue(privateResource.GetInterceptHost()), InterceptPorts: int32ListState(privateResource.GetInterceptPorts()), ProvisioningState: optionalString(provisioningStateToString(privateResource.GetProvisioningState()))}
}
