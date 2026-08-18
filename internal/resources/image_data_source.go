package resources

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	imagesv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/images/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type imageDataSource struct {
	client *agentapi.Client
}

var _ datasource.DataSource = &imageDataSource{}

type imageDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	Repository     types.String `tfsdk:"repository"`
	Description    types.String `tfsdk:"description"`
	Visibility     types.String `tfsdk:"visibility"`
	Versions       types.List   `tfsdk:"versions"`
}

func NewImageDataSource() datasource.DataSource { return &imageDataSource{} }

func (d *imageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (d *imageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a registered image by name, so an environment can name the catalog image it runs instead of carrying its UUID. Resolves the organization's own images and every public one.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the image.",
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization the lookup runs in. Its own images and every public image are in scope.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Image name. An image the organization owns wins over a public one of the same name.",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Which slot the image is built for: `workspace`, `agent_runtime`, or `mcp`. Set it to disambiguate, or to assert the image fills the slot you are about to use it in.",
			},
			"repository": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Upstream repository the image is discovered from.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Free text shown beside the name.",
			},
			"visibility": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "`public` or `internal`.",
			},
			"versions": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Tags the platform has discovered, newest first. A tag an environment names must be one of these.",
			},
		},
	}
}

func (d *imageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*agentapi.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", "Expected *agentapi.Client")
		return
	}
	d.client = client
}

func (d *imageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config imageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter := imagesv1.ImageType_IMAGE_TYPE_UNSPECIFIED
	if !config.Type.IsNull() {
		parsed, message := imageTypeFromString(config.Type.ValueString())
		if message != "" {
			resp.Diagnostics.AddError("Invalid image type", message)
			return
		}
		filter = parsed
	}

	organizationID := config.OrganizationID.ValueString()
	images, err := d.client.ListImages(ctx, organizationID, filter)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list images", err.Error())
		return
	}

	image, diagnostic := selectImageByName(images, config.Name.ValueString(), organizationID)
	if diagnostic != "" {
		resp.Diagnostics.AddError("Unable to find image", diagnostic)
		return
	}

	tags, err := d.client.ListVersionTags(ctx, image.GetMeta().GetId())
	if err != nil {
		resp.Diagnostics.AddError("Unable to list image versions", err.Error())
		return
	}
	versions, listDiagnostics := types.ListValueFrom(ctx, types.StringType, tags)
	resp.Diagnostics.Append(listDiagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ID = types.StringValue(image.GetMeta().GetId())
	config.Type = types.StringValue(imageTypeToString(image.GetType()))
	config.Repository = types.StringValue(image.GetRepository())
	config.Description = types.StringValue(image.GetDescription())
	config.Visibility = types.StringValue(imageVisibilityToString(image.GetVisibility()))
	config.Versions = versions

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// selectImageByName resolves the name the way the platform scopes the list: an
// image the organization owns shadows a public one carrying the same name.
func selectImageByName(images []*imagesv1.Image, name, organizationID string) (*imagesv1.Image, string) {
	var matches []*imagesv1.Image
	for _, image := range images {
		if image.GetName() == name {
			matches = append(matches, image)
		}
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Sprintf("no image named %q is visible to organization %s", name, organizationID)
	case 1:
		return matches[0], ""
	}
	for _, image := range matches {
		if image.GetOrganizationId() == organizationID {
			return image, ""
		}
	}
	ids := make([]string, 0, len(matches))
	for _, image := range matches {
		ids = append(ids, image.GetMeta().GetId())
	}
	sort.Strings(ids)
	return nil, fmt.Sprintf("%d public images are named %q (%s); none belongs to organization %s, so the name is ambiguous. Set type, or reference the image by id.",
		len(matches), name, strings.Join(ids, ", "), organizationID)
}
