package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	appsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/apps/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type appInstallationResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &appInstallationResource{}
var _ resource.ResourceWithImportState = &appInstallationResource{}

type appInstallationModel struct {
	ID             types.String `tfsdk:"id"`
	AppID          types.String `tfsdk:"app_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Slug           types.String `tfsdk:"slug"`
	Configuration  types.String `tfsdk:"configuration"`
}

func NewAppInstallationResource() resource.Resource { return &appInstallationResource{} }

func (r *appInstallationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_installation"
}

func (r *appInstallationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn app installation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the installation.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"app_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "App identifier.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization identifier.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"slug": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Installation slug.",
			},
			"configuration": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "JSON configuration for the installation.",
				Default:             stringdefault.StaticString("{}"),
			},
		},
	}
}

func (r *appInstallationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *appInstallationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan appInstallationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configuration, err := jsonStringToProtoStruct(plan.Configuration.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid configuration JSON", err.Error())
		return
	}

	input := &appsv1.InstallAppRequest{
		AppId:          plan.AppID.ValueString(),
		OrganizationId: plan.OrganizationID.ValueString(),
		Slug:           plan.Slug.ValueString(),
		Configuration:  configuration,
	}

	installation, err := r.client.InstallApp(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to install app", err.Error())
		return
	}

	state := appInstallationModel{
		ID:             types.StringValue(installation.Meta.Id),
		AppID:          types.StringValue(installation.AppId),
		OrganizationID: types.StringValue(installation.OrganizationId),
		Slug:           types.StringValue(installation.Slug),
		Configuration:  types.StringValue(protoStructToJSONString(installation.Configuration)),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *appInstallationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state appInstallationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	installation, err := r.client.GetInstallation(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read installation", err.Error())
		return
	}

	state.ID = types.StringValue(installation.Meta.Id)
	state.AppID = types.StringValue(installation.AppId)
	state.OrganizationID = types.StringValue(installation.OrganizationId)
	state.Slug = types.StringValue(installation.Slug)
	state.Configuration = types.StringValue(protoStructToJSONString(installation.Configuration))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *appInstallationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan appInstallationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state appInstallationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slugChanged := plan.Slug.ValueString() != state.Slug.ValueString()
	configChanged := plan.Configuration.ValueString() != state.Configuration.ValueString()
	if !slugChanged && !configChanged {
		state.Slug = plan.Slug
		state.Configuration = plan.Configuration
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	input := &appsv1.UpdateInstallationRequest{Id: state.ID.ValueString()}
	if slugChanged {
		input.Slug = proto.String(plan.Slug.ValueString())
	}
	if configChanged {
		configuration, err := jsonStringToProtoStruct(plan.Configuration.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid configuration JSON", err.Error())
			return
		}
		input.Configuration = configuration
	}

	installation, err := r.client.UpdateInstallation(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update installation", err.Error())
		return
	}

	updatedState := appInstallationModel{
		ID:             types.StringValue(installation.Meta.Id),
		AppID:          types.StringValue(installation.AppId),
		OrganizationID: types.StringValue(installation.OrganizationId),
		Slug:           types.StringValue(installation.Slug),
		Configuration:  types.StringValue(protoStructToJSONString(installation.Configuration)),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *appInstallationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state appInstallationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UninstallApp(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to uninstall app", err.Error())
		return
	}
}

func (r *appInstallationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func jsonStringToProtoStruct(s string) (*structpb.Struct, error) {
	st := &structpb.Struct{}
	if err := protojson.Unmarshal([]byte(s), st); err != nil {
		return nil, err
	}
	return st, nil
}

func protoStructToJSONString(s *structpb.Struct) string {
	if s == nil {
		return "{}"
	}
	b, _ := protojson.Marshal(s)
	return string(b)
}
