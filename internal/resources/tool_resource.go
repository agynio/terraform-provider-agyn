package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type toolResource struct{}

type toolModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Config      types.String `tfsdk:"config"`
}

func NewToolResource() resource.Resource { return &toolResource{} }

func (r *toolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tool"
}

func (r *toolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"title": schema.StringAttribute{Required: true},
		"description": schema.StringAttribute{Optional: true},
		"type": schema.StringAttribute{Required: true},
		"config": schema.StringAttribute{Optional: true, Description: "JSON-encoded configuration"},
	}}
}

func (r *toolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {}
func (r *toolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) { resp.Diagnostics.AddError("Not Implemented", "TODO") }
func (r *toolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {}
func (r *toolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) { resp.Diagnostics.AddError("Not Implemented", "TODO") }
func (r *toolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {}
func (r *toolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) { resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp) }
