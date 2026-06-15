package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/proto"

	egressv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/egress/v1"
	"github.com/agynio/terraform-provider-agyn/internal/agentapi"
)

type egressRuleResource struct {
	client *agentapi.Client
}

var _ resource.Resource = &egressRuleResource{}
var _ resource.ResourceWithImportState = &egressRuleResource{}

type egressRuleModel struct {
	ID             types.String        `tfsdk:"id"`
	OrganizationID types.String        `tfsdk:"organization_id"`
	Name           types.String        `tfsdk:"name"`
	Description    types.String        `tfsdk:"description"`
	DomainPattern  types.String        `tfsdk:"domain_pattern"`
	Ports          types.String        `tfsdk:"ports"`
	Methods        types.String        `tfsdk:"methods"`
	PathPattern    types.String        `tfsdk:"path_pattern"`
	Action         types.String        `tfsdk:"action"`
	Headers        []egressHeaderModel `tfsdk:"injected_header"`
}

type egressHeaderModel struct {
	Name     types.String `tfsdk:"name"`
	Scheme   types.String `tfsdk:"scheme"`
	Value    types.String `tfsdk:"value"`
	SecretID types.String `tfsdk:"secret_id"`
}

func NewEgressRuleResource() resource.Resource { return &egressRuleResource{} }

func (r *egressRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_egress_rule"
}

func (r *egressRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	credentialValidators := []validator.String{stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("value"), path.MatchRelative().AtParent().AtName("secret_id"))}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn egress rule.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"organization_id": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()}},
			"name":            schema.StringAttribute{Required: true},
			"description":     schema.StringAttribute{Optional: true},
			"domain_pattern":  schema.StringAttribute{Required: true},
			"ports":           schema.StringAttribute{Optional: true, MarkdownDescription: "Comma-separated destination ports. Empty uses service defaults."},
			"methods":         schema.StringAttribute{Optional: true, MarkdownDescription: "Comma-separated HTTP methods. Empty matches all methods."},
			"path_pattern":    schema.StringAttribute{Optional: true},
			"action": schema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{stringvalidator.OneOf("allow", "deny")},
			},
		},
		Blocks: map[string]schema.Block{
			"injected_header": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
					"name":      schema.StringAttribute{Required: true},
					"scheme":    schema.StringAttribute{Optional: true, Validators: []validator.String{stringvalidator.OneOf("bearer", "basic")}},
					"value":     schema.StringAttribute{Optional: true, Sensitive: true, Validators: credentialValidators},
					"secret_id": schema.StringAttribute{Optional: true, Validators: credentialValidators},
				}},
			},
		},
	}
}

func (r *egressRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*agentapi.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *agentapi.Client")
		return
	}
	r.client = client
}

func (r *egressRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan egressRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input, stop := egressRuleCreateRequest(plan, resp)
	if stop {
		return
	}
	rule, err := r.client.CreateEgressRule(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create egress rule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, egressRuleState(rule, plan))...)
}

func (r *egressRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state egressRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rule, err := r.client.GetEgressRule(ctx, state.ID.ValueString())
	if err != nil {
		if agentapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read egress rule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, egressRuleState(rule, state))...)
}

func (r *egressRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var plan egressRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input, stop := egressRuleUpdateRequest(plan, resp)
	if stop {
		return
	}
	rule, err := r.client.UpdateEgressRule(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update egress rule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, egressRuleState(rule, plan))...)
}

func (r *egressRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing API client", "The provider has not been configured")
		return
	}
	var state egressRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteEgressRule(ctx, state.ID.ValueString()); err != nil && !agentapi.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete egress rule", err.Error())
	}
}

func (r *egressRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func egressRuleCreateRequest(plan egressRuleModel, resp *resource.CreateResponse) (*egressv1.CreateEgressRuleRequest, bool) {
	matcher, effect, stop := egressRuleParts(plan, resp.Diagnostics.AddError)
	if stop {
		return nil, true
	}
	return &egressv1.CreateEgressRuleRequest{OrganizationId: plan.OrganizationID.ValueString(), Name: plan.Name.ValueString(), Description: stringValue(plan.Description), Matcher: matcher, Effect: effect}, false
}

func egressRuleUpdateRequest(plan egressRuleModel, resp *resource.UpdateResponse) (*egressv1.UpdateEgressRuleRequest, bool) {
	matcher, effect, stop := egressRuleParts(plan, resp.Diagnostics.AddError)
	if stop {
		return nil, true
	}
	return &egressv1.UpdateEgressRuleRequest{Id: plan.ID.ValueString(), Name: proto.String(plan.Name.ValueString()), Description: proto.String(stringValue(plan.Description)), Matcher: matcher, Effect: effect}, false
}

func egressRuleParts(plan egressRuleModel, addError func(string, string)) (*egressv1.EgressRuleMatcher, *egressv1.EgressRuleEffect, bool) {
	if len(plan.Headers) == 0 && plan.Action.IsNull() {
		addError("Invalid egress rule", "Egress rule requires an action or at least one injected header.")
		return nil, nil, true
	}
	ports, err := parseCSVPorts(stringValue(plan.Ports))
	if err != nil {
		addError("Invalid ports", err.Error())
		return nil, nil, true
	}
	methods := parseCSVStrings(stringValue(plan.Methods), strings.ToUpper)
	inject := make([]*egressv1.EgressRuleHeader, 0, len(plan.Headers))
	for _, header := range plan.Headers {
		if headerHasLiteralValue(header) == headerHasSecretID(header) {
			addError("Invalid injected header", "Each injected header requires exactly one of value or secret_id.")
			return nil, nil, true
		}
		protoHeader := &egressv1.EgressRuleHeader{Name: header.Name.ValueString(), Scheme: headerScheme(header.Scheme.ValueString())}
		if headerHasLiteralValue(header) {
			protoHeader.Credential = &egressv1.EgressRuleHeader_Value{Value: header.Value.ValueString()}
		} else {
			protoHeader.Credential = &egressv1.EgressRuleHeader_SecretId{SecretId: header.SecretID.ValueString()}
		}
		inject = append(inject, protoHeader)
	}
	effect := &egressv1.EgressRuleEffect{Inject: inject}
	if !plan.Action.IsNull() && !plan.Action.IsUnknown() {
		action := egressAction(plan.Action.ValueString())
		effect.Action = &action
	}
	return &egressv1.EgressRuleMatcher{DomainPattern: plan.DomainPattern.ValueString(), Ports: ports, Methods: methods, PathPattern: stringValue(plan.PathPattern)}, effect, false
}

func egressRuleState(rule *egressv1.EgressRule, plan egressRuleModel) egressRuleModel {
	state := plan
	state.ID = types.StringValue(rule.GetMeta().GetId())
	state.OrganizationID = types.StringValue(rule.GetOrganizationId())
	state.Name = types.StringValue(rule.GetName())
	state.Description = optionalString(rule.GetDescription())
	state.DomainPattern = types.StringValue(rule.GetMatcher().GetDomainPattern())
	state.Ports = optionalString(formatInt32CSV(rule.GetMatcher().GetPorts()))
	state.Methods = optionalString(strings.Join(rule.GetMatcher().GetMethods(), ","))
	state.PathPattern = optionalString(rule.GetMatcher().GetPathPattern())
	state.Action = optionalString(actionFromProto(rule.GetEffect().GetAction()))
	for i, header := range rule.GetEffect().GetInject() {
		if i >= len(state.Headers) {
			state.Headers = append(state.Headers, egressHeaderModel{})
		}
		state.Headers[i].Name = types.StringValue(header.GetName())
		state.Headers[i].Scheme = optionalString(schemeFromProto(header.GetScheme()))
		if secretID := header.GetSecretId(); secretID != "" {
			state.Headers[i].SecretID = types.StringValue(secretID)
			state.Headers[i].Value = types.StringNull()
		}
	}
	return state
}

func headerHasLiteralValue(header egressHeaderModel) bool {
	return !header.Value.IsNull() && !header.Value.IsUnknown()
}

func headerHasSecretID(header egressHeaderModel) bool {
	return !header.SecretID.IsNull() && !header.SecretID.IsUnknown()
}

func parseCSVPorts(value string) ([]int32, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	ports := make([]int32, 0, len(parts))
	for _, part := range parts {
		parsed, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || parsed < 1 || parsed > 65535 {
			return nil, fmt.Errorf("ports must be comma-separated integers between 1 and 65535")
		}
		ports = append(ports, int32(parsed))
	}
	return ports, nil
}

func parseCSVStrings(value string, transform func(string) string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, transform(trimmed))
		}
	}
	return values
}

func formatInt32CSV(values []int32) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = strconv.Itoa(int(value))
	}
	return strings.Join(parts, ",")
}

func egressAction(value string) egressv1.EgressRuleAction {
	if value == "deny" {
		return egressv1.EgressRuleAction_EGRESS_RULE_ACTION_DENY
	}
	return egressv1.EgressRuleAction_EGRESS_RULE_ACTION_ALLOW
}

func actionFromProto(value egressv1.EgressRuleAction) string {
	switch value {
	case egressv1.EgressRuleAction_EGRESS_RULE_ACTION_ALLOW:
		return "allow"
	case egressv1.EgressRuleAction_EGRESS_RULE_ACTION_DENY:
		return "deny"
	default:
		return ""
	}
}

func headerScheme(value string) egressv1.HeaderAuthScheme {
	switch value {
	case "bearer":
		return egressv1.HeaderAuthScheme_HEADER_AUTH_SCHEME_BEARER
	case "basic":
		return egressv1.HeaderAuthScheme_HEADER_AUTH_SCHEME_BASIC
	default:
		return egressv1.HeaderAuthScheme_HEADER_AUTH_SCHEME_UNSPECIFIED
	}
}

func schemeFromProto(value egressv1.HeaderAuthScheme) string {
	switch value {
	case egressv1.HeaderAuthScheme_HEADER_AUTH_SCHEME_BEARER:
		return "bearer"
	case egressv1.HeaderAuthScheme_HEADER_AUTH_SCHEME_BASIC:
		return "basic"
	default:
		return ""
	}
}
