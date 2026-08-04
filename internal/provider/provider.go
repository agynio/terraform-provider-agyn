package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
	"github.com/agynio/terraform-provider-agyn/internal/resources"
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
		Description:         "Provider for Agyn Agents API via Gateway.",
		MarkdownDescription: "Provider for Agyn Agents API via Gateway.",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Base URL for the Gateway (e.g., https://gateway.example.com).",
			},
			"api_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Bearer token for Gateway authentication. Required for app management operations. Can also be set via the `AGYN_API_TOKEN` environment variable.",
			},
		},
	}
}

func (p *agynProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data struct {
		APIURL   types.String `tfsdk:"api_url"`
		APIToken types.String `tfsdk:"api_token"`
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.APIURL.IsNull() || data.APIURL.IsUnknown() || data.APIURL.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(path.Root("api_url"), "Missing API URL", "The provider requires api_url to be set to the Gateway base URL.")
		return
	}

	apiToken := ""
	if !data.APIToken.IsNull() && !data.APIToken.IsUnknown() && data.APIToken.ValueString() != "" {
		apiToken = data.APIToken.ValueString()
	} else if envToken := os.Getenv("AGYN_API_TOKEN"); envToken != "" {
		apiToken = envToken
	}

	client, err := agentapi.NewClient(agentapi.Config{
		BaseURL:  data.APIURL.ValueString(),
		APIToken: apiToken,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure client", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *agynProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewAppResource,
		resources.NewAppInstallationResource,
		resources.NewOrganizationResource,
		resources.NewUserResource,
		resources.NewLLMProviderResource,
		resources.NewEnvironmentResource,
		resources.NewImageResource,
		resources.NewModelResource,
		resources.NewRunnerResource,
		resources.NewAgentResource,
		resources.NewVolumeResource,
		resources.NewVolumeAttachmentResource,
		resources.NewSecretProviderResource,
		resources.NewSecretResource,
		resources.NewMcpResource,
		resources.NewSkillResource,
		resources.NewEnvResource,
		resources.NewInitScriptResource,
		resources.NewEgressRuleResource,
		resources.NewEgressRuleAttachmentResource,
		resources.NewNetworkResource,
		resources.NewTunnelCredentialResource,
		resources.NewPrivateResourceResource,
		resources.NewPrivateResourceAccessResource,
		resources.NewGroupResource,
		resources.NewGroupMembershipResource,
	}
}

func (p *agynProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
