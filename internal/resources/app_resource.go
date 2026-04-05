package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	appsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/apps/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type appResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &appResource{}
var _ resource.ResourceWithImportState = &appResource{}

type appModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Slug           types.String `tfsdk:"slug"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Icon           types.String `tfsdk:"icon"`
	Visibility     types.String `tfsdk:"visibility"`
	Permissions    types.List   `tfsdk:"permissions"`
	IdentityID     types.String `tfsdk:"identity_id"`
	ServiceToken   types.String `tfsdk:"service_token"`
}

func NewAppResource() resource.Resource { return &appResource{} }

func (r *appResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (r *appResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn app.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the app.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization identifier for the app.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"slug": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "App slug.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "App name.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable description.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"icon": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Icon URL or identifier.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"visibility": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Visibility for the app (public or internal).",
				Validators:          []validator.String{stringvalidator.OneOf("public", "internal")},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"permissions": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Permissions granted to the app.",
				PlanModifiers:       []planmodifier.List{listplanmodifier.RequiresReplace()},
			},
			"identity_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identity identifier for the app.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"service_token": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Service token for the app.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *appResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *appResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var plan appModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	permissions, diags := permissionsFromPlan(ctx, plan.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &appsv1.CreateAppRequest{
		OrganizationId: plan.OrganizationID.ValueString(),
		Slug:           plan.Slug.ValueString(),
		Name:           plan.Name.ValueString(),
		Description:    stringValue(plan.Description),
		Icon:           stringValue(plan.Icon),
		Visibility:     toProtoVisibility(plan.Visibility.ValueString()),
		Permissions:    permissions,
	}
	result, err := r.client.CreateApp(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create app", err.Error())
		return
	}

	app := result.App
	permissionsState, diags := permissionsToState(ctx, app.Permissions, plan.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updatedState := appModel{
		ID:             types.StringValue(app.Meta.Id),
		OrganizationID: types.StringValue(app.OrganizationId),
		Slug:           types.StringValue(app.Slug),
		Name:           types.StringValue(app.Name),
		Description:    optionalString(app.Description),
		Icon:           optionalString(app.Icon),
		Visibility:     types.StringValue(fromProtoVisibility(app.Visibility)),
		Permissions:    permissionsState,
		IdentityID:     optionalString(app.IdentityId),
		ServiceToken:   optionalString(result.ServiceToken),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *appResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state appModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.GetApp(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read app", err.Error())
		return
	}

	permissionsState, diags := permissionsToState(ctx, app.Permissions, state.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ID = types.StringValue(app.Meta.Id)
	state.OrganizationID = types.StringValue(app.OrganizationId)
	state.Slug = types.StringValue(app.Slug)
	state.Name = types.StringValue(app.Name)
	state.Description = optionalString(app.Description)
	state.Icon = optionalString(app.Icon)
	state.Visibility = types.StringValue(fromProtoVisibility(app.Visibility))
	state.Permissions = permissionsState
	state.IdentityID = optionalString(app.IdentityId)
	state.ServiceToken = preserveSensitiveString(state.ServiceToken, "")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *appResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	resp.Diagnostics.AddError(
		"Update not supported",
		"Apps are immutable. This is an internal error.",
	)
}

func (r *appResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}

	var state appModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteApp(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete app", err.Error())
		return
	}
}

func (r *appResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func permissionsFromPlan(ctx context.Context, permissions types.List) ([]string, diag.Diagnostics) {
	if permissions.IsNull() || permissions.IsUnknown() {
		return nil, nil
	}
	values := []string{}
	diags := permissions.ElementsAs(ctx, &values, false)
	return values, diags
}

func permissionsToState(ctx context.Context, permissions []string, prior types.List) (types.List, diag.Diagnostics) {
	if len(permissions) == 0 {
		if prior.IsNull() || prior.IsUnknown() {
			return types.ListNull(types.StringType), nil
		}
	}
	return types.ListValueFrom(ctx, types.StringType, permissions)
}

func toProtoVisibility(v string) appsv1.AppVisibility {
	switch v {
	case "public":
		return appsv1.AppVisibility_APP_VISIBILITY_PUBLIC
	case "internal":
		return appsv1.AppVisibility_APP_VISIBILITY_INTERNAL
	default:
		panic("unreachable: validated by schema")
	}
}

func fromProtoVisibility(v appsv1.AppVisibility) string {
	switch v {
	case appsv1.AppVisibility_APP_VISIBILITY_PUBLIC:
		return "public"
	case appsv1.AppVisibility_APP_VISIBILITY_INTERNAL:
		return "internal"
	default:
		panic(fmt.Sprintf("unreachable: unexpected proto visibility %v", v))
	}
}
