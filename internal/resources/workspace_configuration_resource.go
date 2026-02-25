package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type workspaceConfigurationResource struct{}

type workspaceConfigurationModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Config      types.String `tfsdk:"config"`
}

func NewWorkspaceConfigurationResource() resource.Resource { return &workspaceConfigurationResource{} }

func (r *workspaceConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_configuration"
}

func (r *workspaceConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"title": schema.StringAttribute{Required: true},
		"description": schema.StringAttribute{Optional: true},
		"config": schema.StringAttribute{Optional: true, Description: "JSON-encoded configuration"},
	}}
}

func (r *workspaceConfigurationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {}
func (r *workspaceConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) { resp.Diagnostics.AddError("Not Implemented", "TODO") }
func (r *workspaceConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {}
func (r *workspaceConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) { resp.Diagnostics.AddError("Not Implemented", "TODO") }
func (r *workspaceConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {}
func (r *workspaceConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) { resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp) }
