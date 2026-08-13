package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	imagesv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/images/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type imageResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &imageResource{}
var _ resource.ResourceWithImportState = &imageResource{}

// Versions are absent by design: they are discovered from the upstream
// repository, never authored, so there is nothing for Terraform to declare.
type imageModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Type           types.String `tfsdk:"type"`
	Repository     types.String `tfsdk:"repository"`
	Username       types.String `tfsdk:"username"`
	SecretID       types.String `tfsdk:"secret_id"`
	Visibility     types.String `tfsdk:"visibility"`
	TagFilter      types.String `tfsdk:"tag_filter"`
}

// Discovery reads the upstream registry, so this covers a slow one without
// holding an apply open indefinitely.
const imageDiscoveryTimeout = 2 * time.Minute

func NewImageResource() resource.Resource { return &imageResource{} }

func (r *imageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (r *imageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an image in the Agyn catalog. Versions are discovered from the upstream repository and cannot be declared.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the image.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization that owns the image.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Image name, unique within the organization. Matches `^[a-z0-9-]+$`.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Free text shown beside the name.",
			},
			"type": schema.StringAttribute{
				Required: true,
				// Immutable: the type is a statement about what the record is,
				// and changing it silently redefines every environment
				// referencing it.
				MarkdownDescription: "Which slot the image is built for: `workspace`, `agent_runtime`, or `mcp`. Immutable.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"repository": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Upstream repository, e.g. `ghcr.io/agynio/devcontainer-go`. Must not carry a tag or digest. Immutable.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"username": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Registry username. Omit for an anonymously readable repository.",
			},
			"secret_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				// An id rather than the password itself: the value lives in a
				// Secret, so it never appears in a plan, a state file, or here.
				MarkdownDescription: "UUID of the Secret holding the registry password. Must belong to the same organization. Omit for an anonymously readable repository.",
			},
			"visibility": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "`public` (usable by every organization) or `internal` (the owning organization only).",
			},
			"tag_filter": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional glob limiting which tags appear in pickers.",
			},
		},
	}
}

func (r *imageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *imageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan imageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	imageType, diag := imageTypeFromString(plan.Type.ValueString())
	if diag != "" {
		resp.Diagnostics.AddError("Invalid image type", diag)
		return
	}
	visibility, diag := imageVisibilityFromString(plan.Visibility.ValueString())
	if diag != "" {
		resp.Diagnostics.AddError("Invalid image visibility", diag)
		return
	}

	image, err := r.client.CreateImage(ctx, &imagesv1.CreateImageRequest{
		OrganizationId: plan.OrganizationID.ValueString(),
		Name:           plan.Name.ValueString(),
		Description:    plan.Description.ValueString(),
		Type:           imageType,
		Repository:     plan.Repository.ValueString(),
		Username:       plan.Username.ValueString(),
		SecretId:       plan.SecretID.ValueString(),
		Visibility:     visibility,
		TagFilter:      plan.TagFilter.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create image", err.Error())
		return
	}

	// Versions are discovered after the record exists, and anything naming a
	// tag of this image resolves against them. Returning before the first one
	// is known makes every dependent resource a race.
	if err := r.client.WaitForVersion(ctx, image.Meta.Id, imageDiscoveryTimeout); err != nil {
		resp.Diagnostics.AddWarning("No image version discovered yet",
			fmt.Sprintf("%s. A resource naming a tag of this image may fail until one is.", err))
	}

	state := imageStateFrom(image)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *imageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state imageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	image, err := r.client.GetImage(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read image", err.Error())
		return
	}

	// The credential is a reference, so unlike a password it round-trips and
	// drift on it is detectable.
	state = imageStateFrom(image)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *imageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan, state imageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &imagesv1.UpdateImageRequest{
		Id:          plan.ID.ValueString(),
		Name:        updateStringPointer(plan.Name, state.Name),
		Description: updateStringPointer(plan.Description, state.Description),
		Username:    updateStringPointer(plan.Username, state.Username),
		TagFilter:   updateStringPointer(plan.TagFilter, state.TagFilter),
		SecretId:    updateStringPointer(plan.SecretID, state.SecretID),
	}
	if !plan.Visibility.Equal(state.Visibility) {
		visibility, diag := imageVisibilityFromString(plan.Visibility.ValueString())
		if diag != "" {
			resp.Diagnostics.AddError("Invalid image visibility", diag)
			return
		}
		input.Visibility = &visibility
	}

	image, err := r.client.UpdateImage(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update image", err.Error())
		return
	}

	updated := imageStateFrom(image)
	resp.Diagnostics.Append(resp.State.Set(ctx, &updated)...)
}

func (r *imageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state imageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Deleting is not blocked by references: environments naming the image are
	// flagged unschedulable rather than repaired.
	if err := r.client.DeleteImage(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete image", err.Error())
	}
}

func (r *imageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func imageStateFrom(image *imagesv1.Image) imageModel {
	return imageModel{
		ID:             types.StringValue(image.GetMeta().GetId()),
		OrganizationID: types.StringValue(image.GetOrganizationId()),
		Name:           types.StringValue(image.GetName()),
		Description:    types.StringValue(image.GetDescription()),
		Type:           types.StringValue(imageTypeToString(image.GetType())),
		Repository:     types.StringValue(image.GetRepository()),
		Username:       types.StringValue(image.GetUsername()),
		SecretID:       types.StringValue(image.GetSecretId()),
		Visibility:     types.StringValue(imageVisibilityToString(image.GetVisibility())),
		TagFilter:      types.StringValue(image.GetTagFilter()),
	}
}

// The configuration names types and visibilities the way the product does, so
// an operator writes "workspace" rather than an enum constant.
func imageTypeFromString(value string) (imagesv1.ImageType, string) {
	switch value {
	case "workspace":
		return imagesv1.ImageType_IMAGE_TYPE_WORKSPACE, ""
	case "agent_runtime":
		return imagesv1.ImageType_IMAGE_TYPE_AGENT_RUNTIME, ""
	case "mcp":
		return imagesv1.ImageType_IMAGE_TYPE_MCP, ""
	default:
		return 0, fmt.Sprintf("type %q must be workspace, agent_runtime, or mcp", value)
	}
}

func imageTypeToString(value imagesv1.ImageType) string {
	switch value {
	case imagesv1.ImageType_IMAGE_TYPE_WORKSPACE:
		return "workspace"
	case imagesv1.ImageType_IMAGE_TYPE_AGENT_RUNTIME:
		return "agent_runtime"
	case imagesv1.ImageType_IMAGE_TYPE_MCP:
		return "mcp"
	default:
		return ""
	}
}

func imageVisibilityFromString(value string) (imagesv1.ImageVisibility, string) {
	switch value {
	case "public":
		return imagesv1.ImageVisibility_IMAGE_VISIBILITY_PUBLIC, ""
	case "internal":
		return imagesv1.ImageVisibility_IMAGE_VISIBILITY_INTERNAL, ""
	default:
		return 0, fmt.Sprintf("visibility %q must be public or internal", value)
	}
}

func imageVisibilityToString(value imagesv1.ImageVisibility) string {
	switch value {
	case imagesv1.ImageVisibility_IMAGE_VISIBILITY_PUBLIC:
		return "public"
	case imagesv1.ImageVisibility_IMAGE_VISIBILITY_INTERNAL:
		return "internal"
	default:
		return ""
	}
}
