package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type agynProvider struct {
	version string
	commit  string
}

func New(version, commit string) func() provider.Provider {
	return func() provider.Provider { return &agynProvider{version: version, commit: commit} }
}

func (p *agynProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "agyn"
	resp.Version = p.version
}

func (p *agynProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Provider for Agyn Team API via Gateway.",
		MarkdownDescription: "Provider for Agyn Team API via Gateway.",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Base URL for the Gateway (e.g., https://gateway.example.com).",
			},
		},
	}
}

type providerData struct {
	cfg Config
	client interface{}
}

func (p *agynProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data struct {
		APIURL types.String `tfsdk:"api_url"`
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.APIURL.IsNull() || data.APIURL.IsUnknown() || data.APIURL.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(path.Root("api_url"), "Missing API URL", "The provider requires api_url to be set to the Gateway base URL.")
		return
	}
	pd := providerData{cfg: Config{APIURL: data.APIURL.ValueString()}, client: nil}
	resp.DataSourceData = &pd
	resp.ResourceData = &pd
}

func (p *agynProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAgentResource,
		NewToolResource,
		NewMCPServerResource,
		NewWorkspaceConfigurationResource,
		NewMemoryBucketResource,
		NewAttachmentResource,
	}
}

func (p *agynProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
