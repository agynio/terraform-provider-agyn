package resources

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-agyn/internal/teamapi"
)

type memoryBucketResource struct {
	client *teamapi.Client
}

type memoryBucketModel struct {
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
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"title":       schema.StringAttribute{Optional: true},
			"description": schema.StringAttribute{Optional: true},
			"config": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "JSON-encoded memory bucket configuration.",
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

	configValue := plan.Config.ValueString()
	if configValue == "" {
		resp.Diagnostics.AddAttributeError(path.Root("config"), "Missing Config", "Memory bucket config must be provided and cannot be empty.")
		return
	}
	if !json.Valid([]byte(configValue)) {
		resp.Diagnostics.AddAttributeError(path.Root("config"), "Invalid JSON", "Memory bucket config must be valid JSON.")
		return
	}

	create := teamapi.MemoryBucketCreate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
		Config:      json.RawMessage(configValue),
	}

	bucket, err := r.client.CreateMemoryBucket(ctx, create)
	if err != nil {
		resp.Diagnostics.AddError("Create Memory Bucket Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(bucket.ID)
	plan.Title = optionalString(bucket.Title)
	plan.Description = optionalString(bucket.Description)
	plan.Config = types.StringValue(string(bucket.Config))

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

	state.ID = types.StringValue(bucket.ID)
	state.Title = optionalString(bucket.Title)
	state.Description = optionalString(bucket.Description)
	state.Config = types.StringValue(string(bucket.Config))

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

	update := teamapi.MemoryBucketUpdate{
		Title:       stringPointer(plan.Title),
		Description: stringPointer(plan.Description),
	}

	if !plan.Config.IsUnknown() && !plan.Config.IsNull() {
		configValue := plan.Config.ValueString()
		if configValue == "" || !json.Valid([]byte(configValue)) {
			resp.Diagnostics.AddAttributeError(path.Root("config"), "Invalid JSON", "Memory bucket config must be valid JSON when provided.")
			return
		}
		raw := json.RawMessage(configValue)
		update.Config = &raw
	}

	bucket, err := r.client.UpdateMemoryBucket(ctx, state.ID.ValueString(), update)
	if err != nil {
		resp.Diagnostics.AddError("Update Memory Bucket Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(bucket.ID)
	plan.Title = optionalString(bucket.Title)
	plan.Description = optionalString(bucket.Description)
	plan.Config = types.StringValue(string(bucket.Config))

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
