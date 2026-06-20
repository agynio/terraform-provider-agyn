package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	networksv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/networks/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type tunnelCredentialResource struct{ client *agentapi.Client }

var _ resource.Resource = &tunnelCredentialResource{}
var _ resource.ResourceWithImportState = &tunnelCredentialResource{}

type tunnelCredentialModel struct {
	ID                     types.String `tfsdk:"id"`
	NetworkID              types.String `tfsdk:"network_id"`
	EnrollmentJWT          types.String `tfsdk:"enrollment_jwt"`
	EnrollmentJWTRevealed  types.Bool   `tfsdk:"enrollment_jwt_revealed"`
	EnrollmentJWTExpiresAt types.String `tfsdk:"enrollment_jwt_expires_at"`
	EnrollmentState        types.String `tfsdk:"enrollment_state"`
	Connectivity           types.String `tfsdk:"connectivity"`
	ProvisioningState      types.String `tfsdk:"provisioning_state"`
	EnrolledAt             types.String `tfsdk:"enrolled_at"`
	LastSeenAt             types.String `tfsdk:"last_seen_at"`
}

func NewTunnelCredentialResource() resource.Resource { return &tunnelCredentialResource{} }

func (r *tunnelCredentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tunnel_credential"
}

func (r *tunnelCredentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Manages an Agyn private network tunnel credential.", Attributes: map[string]schema.Attribute{
		"id":                        schema.StringAttribute{Computed: true, MarkdownDescription: "UUID identifier of the tunnel credential.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"network_id":                schema.StringAttribute{Required: true, MarkdownDescription: "Private network identifier.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"enrollment_jwt":            schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "One-time OpenZiti enrollment JWT. The API only returns this during creation.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"enrollment_jwt_revealed":   schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the enrollment JWT has been revealed."},
		"enrollment_jwt_expires_at": schema.StringAttribute{Computed: true, MarkdownDescription: "Enrollment JWT expiration time in RFC3339 format."},
		"enrollment_state":          schema.StringAttribute{Computed: true, MarkdownDescription: "Tunnel enrollment state."},
		"connectivity":              schema.StringAttribute{Computed: true, MarkdownDescription: "Tunnel connectivity state."},
		"provisioning_state":        schema.StringAttribute{Computed: true, MarkdownDescription: "OpenZiti provisioning state."},
		"enrolled_at":               schema.StringAttribute{Computed: true, MarkdownDescription: "Enrollment completion time in RFC3339 format."},
		"last_seen_at":              schema.StringAttribute{Computed: true, MarkdownDescription: "Last observed tunnel connection time in RFC3339 format."},
	}}
}

func (r *tunnelCredentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *tunnelCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan tunnelCredentialModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateTunnelCredential(ctx, &networksv1.CreateTunnelCredentialRequest{NetworkId: plan.NetworkID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create tunnel credential", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tunnelCredentialState(created.TunnelCredential, types.StringValue(created.EnrollmentJWT)))...)
}

func (r *tunnelCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state tunnelCredentialModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	credential, err := r.client.GetTunnelCredential(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read tunnel credential", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tunnelCredentialState(credential, state.EnrollmentJWT))...)
}

func (r *tunnelCredentialResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Tunnel credentials are immutable. This is an internal error.")
}

func (r *tunnelCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state tunnelCredentialModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteTunnelCredential(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete tunnel credential", err.Error())
	}
}

func (r *tunnelCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func tunnelCredentialState(credential *networksv1.TunnelCredential, enrollmentJWT types.String) tunnelCredentialModel {
	return tunnelCredentialModel{ID: types.StringValue(credential.GetMeta().GetId()), NetworkID: types.StringValue(credential.GetNetworkId()), EnrollmentJWT: enrollmentJWT, EnrollmentJWTRevealed: types.BoolValue(credential.GetEnrollmentJwtRevealed()), EnrollmentJWTExpiresAt: timestampString(credential.GetEnrollmentJwtExpiresAt()), EnrollmentState: optionalString(tunnelEnrollmentStateToString(credential.GetEnrollmentState())), Connectivity: optionalString(tunnelConnectivityToString(credential.GetConnectivity())), ProvisioningState: optionalString(provisioningStateToString(credential.GetProvisioningState())), EnrolledAt: timestampString(credential.GetEnrolledAt()), LastSeenAt: timestampString(credential.GetLastSeenAt())}
}
