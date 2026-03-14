package resources

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-agyn/internal/teamapi"
)

type memoryBucketResource struct {
	client *teamapi.Client
}

var _ resource.ResourceWithUpgradeState = &memoryBucketResource{}

type memoryBucketModel struct {
	ID               types.String `tfsdk:"id"`
	Title            types.String `tfsdk:"title"`
	Description      types.String `tfsdk:"description"`
	Config           types.String `tfsdk:"config"`
	Scope            types.String `tfsdk:"scope"`
	CollectionPrefix types.String `tfsdk:"collection_prefix"`
}

type memoryBucketModelV0 struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Config      types.String `tfsdk:"config"`
}

func NewMemoryBucketResource() resource.Resource { return &memoryBucketResource{} }

func (r *memoryBucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_memory_bucket"
}

func (r *memoryBucketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn memory bucket.",
		Version:             1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID identifier of the memory bucket.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"title": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable title of the memory bucket.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable description of the memory bucket.",
			},
			"config": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Deprecated JSON-encoded memory bucket configuration. Use structured attributes instead.",
				DeprecationMessage:  "Use structured configuration attributes instead of config.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("scope"),
						path.MatchRoot("collection_prefix"),
					),
				},
			},
			"scope": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Scope for the memory bucket (global or perThread).",
				Validators: []validator.String{
					stringvalidator.OneOf("global", "perThread"),
				},
			},
			"collection_prefix": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Collection prefix for memory entries.",
			},
		},
	}
}

func (r *memoryBucketResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
					},
					"title": schema.StringAttribute{
						Optional: true,
					},
					"description": schema.StringAttribute{
						Optional: true,
					},
					"config": schema.StringAttribute{
						Required: true,
					},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior memoryBucketModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}

				config, diags := memoryBucketConfigFromString(prior.Config)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgraded := memoryBucketModel{
					ID:          prior.ID,
					Title:       prior.Title,
					Description: prior.Description,
					Config:      types.StringNull(),
				}
				applyMemoryBucketConfigToModel(&upgraded, config)

				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
	}
}

func (r *memoryBucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*teamapi.Client)
	if !ok || client == nil {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Unable to obtain configured API client")
		return
	}
	r.client = client
}

func (r *memoryBucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing memory buckets.")
		return
	}

	var plan memoryBucketModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, diags := memoryBucketConfigFromPlan(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	create := teamapi.MemoryBucketCreate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
		Config:      config,
	}

	bucket, err := r.client.CreateMemoryBucket(ctx, create)
	if err != nil {
		resp.Diagnostics.AddError("Create Memory Bucket Failed", err.Error())
		return
	}

	configValue := plan.Config
	plan.ID = types.StringValue(bucket.ID)
	plan.Title = optionalString(bucket.Title)
	plan.Description = optionalString(bucket.Description)
	if configValue.IsNull() || configValue.IsUnknown() {
		applyMemoryBucketConfigToModel(&plan, bucket.Config)
	}
	plan.Config, diags = configStateValue(configValue, bucket.Config)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *memoryBucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing memory buckets.")
		return
	}

	var state memoryBucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucket, err := r.client.GetMemoryBucket(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Memory Bucket Failed", err.Error())
		return
	}

	configValue := state.Config
	state.ID = types.StringValue(bucket.ID)
	state.Title = optionalString(bucket.Title)
	state.Description = optionalString(bucket.Description)
	if configValue.IsNull() || configValue.IsUnknown() {
		applyMemoryBucketConfigToModel(&state, bucket.Config)
	}
	var diags diag.Diagnostics
	state.Config, diags = configStateValue(configValue, bucket.Config)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *memoryBucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing memory buckets.")
		return
	}

	var plan memoryBucketModel
	var state memoryBucketModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, sendConfig, diags := memoryBucketConfigForUpdate(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	update := teamapi.MemoryBucketUpdate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
	}
	if sendConfig {
		update.Config = &config
	}

	bucket, err := r.client.UpdateMemoryBucket(ctx, state.ID.ValueString(), update)
	if err != nil {
		resp.Diagnostics.AddError("Update Memory Bucket Failed", err.Error())
		return
	}

	configValue := plan.Config
	plan.ID = types.StringValue(bucket.ID)
	plan.Title = optionalString(bucket.Title)
	plan.Description = optionalString(bucket.Description)
	if configValue.IsNull() || configValue.IsUnknown() {
		applyMemoryBucketConfigToModel(&plan, bucket.Config)
	}
	plan.Config, diags = configStateValue(configValue, bucket.Config)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *memoryBucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing memory buckets.")
		return
	}

	var state memoryBucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteMemoryBucket(ctx, state.ID.ValueString()); err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			return
		}
		resp.Diagnostics.AddError("Delete Memory Bucket Failed", err.Error())
	}
}

func (r *memoryBucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func memoryBucketConfigFromPlan(plan memoryBucketModel) (teamapi.MemoryBucketConfig, diag.Diagnostics) {
	if !plan.Config.IsNull() && !plan.Config.IsUnknown() {
		return memoryBucketConfigFromString(plan.Config)
	}
	return memoryBucketConfigFromFields(plan), nil
}

func memoryBucketConfigForUpdate(plan memoryBucketModel) (teamapi.MemoryBucketConfig, bool, diag.Diagnostics) {
	if !plan.Config.IsNull() && !plan.Config.IsUnknown() {
		config, diags := memoryBucketConfigFromString(plan.Config)
		return config, true, diags
	}

	if !memoryBucketHasTypedConfig(plan) {
		return teamapi.MemoryBucketConfig{}, false, nil
	}

	return memoryBucketConfigFromFields(plan), true, nil
}

func memoryBucketHasTypedConfig(plan memoryBucketModel) bool {
	return !(plan.Scope.IsNull() || plan.Scope.IsUnknown()) ||
		!(plan.CollectionPrefix.IsNull() || plan.CollectionPrefix.IsUnknown())
}

func memoryBucketConfigFromFields(plan memoryBucketModel) teamapi.MemoryBucketConfig {
	return teamapi.MemoryBucketConfig{
		Scope:            stringPointer(plan.Scope),
		CollectionPrefix: stringPointer(plan.CollectionPrefix),
	}
}

func memoryBucketConfigFromString(value types.String) (teamapi.MemoryBucketConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		diags.AddAttributeError(path.Root("config"), "Missing Config", "Memory bucket config must be provided and cannot be empty.")
		return teamapi.MemoryBucketConfig{}, diags
	}
	configValue := value.ValueString()
	if configValue == "" {
		diags.AddAttributeError(path.Root("config"), "Missing Config", "Memory bucket config must be provided and cannot be empty.")
		return teamapi.MemoryBucketConfig{}, diags
	}
	var config teamapi.MemoryBucketConfig
	if err := json.Unmarshal([]byte(configValue), &config); err != nil {
		diags.AddAttributeError(path.Root("config"), "Invalid JSON", "Memory bucket config must be valid JSON.")
		return teamapi.MemoryBucketConfig{}, diags
	}
	return config, diags
}

func applyMemoryBucketConfigToModel(model *memoryBucketModel, config teamapi.MemoryBucketConfig) {
	model.Scope = optionalString(config.Scope)
	model.CollectionPrefix = optionalString(config.CollectionPrefix)
}
