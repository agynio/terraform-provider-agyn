package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	runnersv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/runners/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type runnerDataSource struct {
	client *agentapi.Client
}

var _ datasource.DataSource = &runnerDataSource{}

type runnerDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Labels         types.Map    `tfsdk:"labels"`
	Capabilities   types.List   `tfsdk:"capabilities"`
}

func NewRunnerDataSource() datasource.DataSource { return &runnerDataSource{} }

func (d *runnerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runner"
}

func (d *runnerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a runner by name, so an environment can name the runner it places workloads on instead of carrying its UUID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the runner.",
			},
			"organization_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Organization to look in. Omit to search every runner the token can see.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Runner name.",
			},
			"labels": schema.MapAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Runner labels.",
			},
			"capabilities": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Capabilities the runner supports.",
			},
		},
	}
}

func (d *runnerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *runnerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config runnerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	organizationID := stringValue(config.OrganizationID)
	runners, err := d.client.ListRunners(ctx, organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list runners", err.Error())
		return
	}

	name := config.Name.ValueString()
	var matches []*runnersv1.Runner
	for _, runner := range runners {
		if runner.GetName() == name {
			matches = append(matches, runner)
		}
	}
	switch len(matches) {
	case 0:
		resp.Diagnostics.AddError("Unable to find runner", fmt.Sprintf("no runner named %q is visible to this token", name))
		return
	case 1:
	default:
		resp.Diagnostics.AddError("Unable to find runner", fmt.Sprintf("%d runners are named %q; set organization_id to narrow the lookup", len(matches), name))
		return
	}
	runner := matches[0]

	labels, labelDiagnostics := types.MapValueFrom(ctx, types.StringType, runner.GetLabels())
	resp.Diagnostics.Append(labelDiagnostics...)
	capabilities, capabilityDiagnostics := types.ListValueFrom(ctx, types.StringType, runner.GetCapabilities())
	resp.Diagnostics.Append(capabilityDiagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ID = types.StringValue(runner.GetMeta().GetId())
	config.Labels = labels
	config.Capabilities = capabilities
	if runner.OrganizationId != nil {
		config.OrganizationID = types.StringValue(runner.GetOrganizationId())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
