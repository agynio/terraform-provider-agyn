package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type mcpServerResource struct{}

type mcpServerModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Config      types.String `tfsdk:"config"`
}

func NewMCPServerResource() resource.Resource { return &mcpServerResource{} }

func (r *mcpServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_server"
}

func (r *mcpServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"title": schema.StringAttribute{Required: true},
		"description": schema.StringAttribute{Optional: true},
		"config": schema.StringAttribute{Optional: true, Description: "JSON-encoded configuration"},
	}}
}

func (r *mcpServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {}
func (r *mcpServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) { resp.Diagnostics.AddError("Not Implemented", "TODO") }
func (r *mcpServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {}
func (r *mcpServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) { resp.Diagnostics.AddError("Not Implemented", "TODO") }
func (r *mcpServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {}
func (r *mcpServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) { resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp) }
