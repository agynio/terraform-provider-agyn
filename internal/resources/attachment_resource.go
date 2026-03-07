package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-agyn/internal/teamapi"
)

type attachmentResource struct {
	client *teamapi.Client
}

type attachmentModel struct {
	ID         types.String `tfsdk:"id"`
	Kind       types.String `tfsdk:"kind"`
	SourceID   types.String `tfsdk:"source_id"`
	SourceType types.String `tfsdk:"source_type"`
	TargetID   types.String `tfsdk:"target_id"`
	TargetType types.String `tfsdk:"target_type"`
}

func NewAttachmentResource() resource.Resource { return &attachmentResource{} }

func (r *attachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_attachment"
}

func (r *attachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"kind": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Attachment kind specifying the relation type.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"source_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Source entity ID (UUID).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"source_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Source entity type as reported by the API.",
			},
			"target_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Target entity ID (UUID).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"target_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Target entity type as reported by the API.",
			},
		},
	}
}

func (r *attachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *attachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing attachments.")
		return
	}

	var plan attachmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Kind.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(path.Root("kind"), "Missing Kind", "Attachment kind must be provided.")
		return
	}
	if plan.SourceID.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(path.Root("source_id"), "Missing Source", "Attachment source_id must be provided.")
		return
	}
	if plan.TargetID.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(path.Root("target_id"), "Missing Target", "Attachment target_id must be provided.")
		return
	}

	kind := plan.Kind.ValueString()
	sourceID := plan.SourceID.ValueString()
	targetID := plan.TargetID.ValueString()

	if isGraphAttachmentKind(kind) {
		attachment, err := r.createGraphAttachment(ctx, kind, sourceID, targetID)
		if err != nil {
			resp.Diagnostics.AddError("Create Attachment Failed", err.Error())
			return
		}

		setAttachmentState(&plan, attachment)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	apiSourceID, apiTargetID := attachmentRequestIDs(kind, sourceID, targetID)
	create := teamapi.AttachmentCreate{
		Kind:     kind,
		SourceID: apiSourceID,
		TargetID: apiTargetID,
	}

	attachment, err := r.client.CreateAttachment(ctx, create)
	if err != nil {
		resp.Diagnostics.AddError("Create Attachment Failed", err.Error())
		return
	}

	setAttachmentState(&plan, attachment)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *attachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing attachments.")
		return
	}

	var state attachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() || state.ID.IsUnknown() || state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	if isGraphAttachmentKind(state.Kind.ValueString()) {
		attachment, err := r.readGraphAttachment(ctx, state.Kind.ValueString(), state.ID.ValueString())
		if err != nil {
			if errors.Is(err, teamapi.ErrGraphEdgeNotFound) {
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError("Read Attachment Failed", err.Error())
			return
		}

		setAttachmentState(&state, attachment)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	attachment, err := r.client.GetAttachment(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, teamapi.ErrAttachmentNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Attachment Failed", err.Error())
		return
	}

	setAttachmentState(&state, attachment)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *attachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing attachments.")
		return
	}

	var state attachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if isGraphAttachmentKind(state.Kind.ValueString()) {
		attachment, err := r.readGraphAttachment(ctx, state.Kind.ValueString(), state.ID.ValueString())
		if err != nil {
			if errors.Is(err, teamapi.ErrGraphEdgeNotFound) {
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError("Refresh Attachment Failed", err.Error())
			return
		}

		setAttachmentState(&state, attachment)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	attachment, err := r.client.GetAttachment(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, teamapi.ErrAttachmentNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Refresh Attachment Failed", err.Error())
		return
	}

	setAttachmentState(&state, attachment)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *attachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider must be configured before managing attachments.")
		return
	}

	var state attachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if isGraphAttachmentKind(state.Kind.ValueString()) {
		if state.ID.IsNull() || state.ID.IsUnknown() || state.ID.ValueString() == "" {
			return
		}
		if err := r.client.DeleteGraphEdge(ctx, state.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Delete Attachment Failed", err.Error())
		}
		return
	}

	if err := r.client.DeleteAttachment(ctx, state.ID.ValueString()); err != nil {
		var apiErr *teamapi.APIError
		if errors.As(err, &apiErr) && apiErr.Status == httpStatusNotFound {
			return
		}
		resp.Diagnostics.AddError("Delete Attachment Failed", err.Error())
	}
}

func (r *attachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *attachmentResource) createGraphAttachment(ctx context.Context, kind, sourceID, targetID string) (*teamapi.Attachment, error) {
	apiSourceID, apiTargetID := attachmentRequestIDs(kind, sourceID, targetID)
	sourceHandle, targetHandle, ok := graphAttachmentHandles(kind)
	if !ok {
		return nil, fmt.Errorf("unsupported graph attachment kind %q", kind)
	}

	edgeID := graphEdgeID(apiSourceID, sourceHandle, apiTargetID, targetHandle)
	edge := teamapi.GraphEdge{
		ID:           edgeID,
		Source:       apiSourceID,
		SourceHandle: sourceHandle,
		Target:       apiTargetID,
		TargetHandle: targetHandle,
	}
	if err := r.client.UpsertGraphEdge(ctx, edge); err != nil {
		return nil, err
	}

	sourceType, targetType, ok := graphAttachmentTypes(kind)
	if !ok {
		return nil, fmt.Errorf("unsupported graph attachment kind %q", kind)
	}

	return &teamapi.Attachment{
		ID:         edgeID,
		Kind:       kind,
		SourceID:   apiSourceID,
		SourceType: sourceType,
		TargetID:   apiTargetID,
		TargetType: targetType,
	}, nil
}

func (r *attachmentResource) readGraphAttachment(ctx context.Context, kind, edgeID string) (*teamapi.Attachment, error) {
	edge, err := r.client.FindGraphEdge(ctx, edgeID)
	if err != nil {
		return nil, err
	}

	sourceHandle, targetHandle, ok := graphAttachmentHandles(kind)
	if !ok {
		return nil, fmt.Errorf("unsupported graph attachment kind %q", kind)
	}
	if edge.SourceHandle != sourceHandle || edge.TargetHandle != targetHandle {
		return nil, teamapi.ErrGraphEdgeNotFound
	}

	sourceType, targetType, ok := graphAttachmentTypes(kind)
	if !ok {
		return nil, fmt.Errorf("unsupported graph attachment kind %q", kind)
	}

	return &teamapi.Attachment{
		ID:         edge.ID,
		Kind:       kind,
		SourceID:   edge.Source,
		SourceType: sourceType,
		TargetID:   edge.Target,
		TargetType: targetType,
	}, nil
}

func isGraphAttachmentKind(kind string) bool {
	return kind == mcpServerWorkspaceAttachmentKind
}

func graphAttachmentHandles(kind string) (string, string, bool) {
	if kind == mcpServerWorkspaceAttachmentKind {
		return graphWorkspaceSourceHandle, graphMcpServerTargetHandle, true
	}
	return "", "", false
}

func graphAttachmentTypes(kind string) (string, string, bool) {
	if kind == mcpServerWorkspaceAttachmentKind {
		return graphWorkspaceSourceType, graphMcpServerTargetType, true
	}
	return "", "", false
}

func graphEdgeID(sourceID, sourceHandle, targetID, targetHandle string) string {
	return fmt.Sprintf("%s-%s__%s-%s", sourceID, sourceHandle, targetID, targetHandle)
}

const (
	mcpServerWorkspaceAttachmentKind = "mcpServer_workspaceConfiguration"
	graphWorkspaceSourceHandle       = "$self"
	graphMcpServerTargetHandle       = "workspace"
	graphWorkspaceSourceType         = "workspaceConfiguration"
	graphMcpServerTargetType         = "mcpServer"
)

func attachmentRequestIDs(kind, sourceID, targetID string) (string, string) {
	if kind == mcpServerWorkspaceAttachmentKind {
		return targetID, sourceID
	}
	return sourceID, targetID
}

func setAttachmentState(state *attachmentModel, attachment *teamapi.Attachment) {
	state.ID = types.StringValue(attachment.ID)
	state.Kind = types.StringValue(attachment.Kind)

	sourceID := attachment.SourceID
	targetID := attachment.TargetID
	sourceType := attachment.SourceType
	targetType := attachment.TargetType
	if attachment.Kind == mcpServerWorkspaceAttachmentKind {
		sourceID, targetID = targetID, sourceID
		sourceType, targetType = targetType, sourceType
	}

	state.SourceID = types.StringValue(sourceID)
	state.SourceType = types.StringValue(sourceType)
	state.TargetID = types.StringValue(targetID)
	state.TargetType = types.StringValue(targetType)
}
