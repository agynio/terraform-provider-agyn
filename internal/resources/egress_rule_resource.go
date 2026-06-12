package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
	Ports          []types.Int32       `tfsdk:"ports"`
	Methods        []types.String      `tfsdk:"methods"`
	PathPattern    types.String        `tfsdk:"path_pattern"`
	Action         types.String        `tfsdk:"action"`
	Headers        []egressHeaderModel `tfsdk:"header"`
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
	headerCredentialPaths := []path.Expression{
		path.MatchRelative().AtParent().AtName("value"),
		path.MatchRelative().AtParent().AtName("secret_id"),
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Agyn egress rule.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Computed: true, MarkdownDescription: "UUID identifier of the egress rule.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"organization_id": schema.StringAttribute{Required: true, MarkdownDescription: "Organization identifier for the egress rule.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()}},
			"name":            schema.StringAttribute{Required: true, MarkdownDescription: "Rule name."},
			"description":     schema.StringAttribute{Optional: true, MarkdownDescription: "Human-readable description."},
			"domain_pattern":  schema.StringAttribute{Required: true, MarkdownDescription: "Destination host pattern, for example `api.example.com` or `*.example.com`."},
			"ports":           schema.ListAttribute{Optional: true, Computed: true, ElementType: types.Int32Type, MarkdownDescription: "Destination ports to intercept. Empty uses platform defaults.", Validators: []validator.List{listvalidator.ValueInt32sAre(int32validator.Between(1, 65535))}, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
			"methods":         schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, MarkdownDescription: "HTTP methods to match. Empty matches all methods.", PlanModifiers: []planmodifier.List{normalizeEgressMethodsPlan(), listplanmodifier.UseStateForUnknown()}},
			"path_pattern":    schema.StringAttribute{Optional: true, MarkdownDescription: "Request path glob. Empty matches all paths."},
			"action":          schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Rule action. One of `allow` or `deny`.", Validators: []validator.String{stringvalidator.OneOf("allow", "deny")}},
		},
		Blocks: map[string]schema.Block{
			"header": schema.ListNestedBlock{
				MarkdownDescription: "Header to inject for matching requests.",
				NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
					"name":   schema.StringAttribute{Required: true, MarkdownDescription: "Header name."},
					"scheme": schema.StringAttribute{Optional: true, MarkdownDescription: "Credential scheme. One of `bearer` or `basic`.", Validators: []validator.String{stringvalidator.OneOf("bearer", "basic")}},
					"value": schema.StringAttribute{Optional: true, Sensitive: true, MarkdownDescription: "Literal header value.", Validators: []validator.String{
						stringvalidator.AtLeastOneOf(headerCredentialPaths...),
						stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("secret_id")),
					}},
					"secret_id": schema.StringAttribute{Optional: true, MarkdownDescription: "Secret identifier containing the header value.", Validators: []validator.String{
						stringvalidator.AtLeastOneOf(headerCredentialPaths...),
						stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("value")),
					}},
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
	input, err := createEgressRuleRequest(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid egress rule", err.Error())
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
	matcher, err := egressMatcherFromModel(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid egress matcher", err.Error())
		return
	}
	effect, err := egressEffectFromModel(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid egress effect", err.Error())
		return
	}
	rule, err := r.client.UpdateEgressRule(ctx, &egressv1.UpdateEgressRuleRequest{Id: plan.ID.ValueString(), Name: stringPtr(plan.Name.ValueString()), Description: stringPtr(stringValue(plan.Description)), Matcher: matcher, Effect: effect})
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
	if err := r.client.DeleteEgressRule(ctx, state.ID.ValueString()); err != nil {
		if agentapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete egress rule", err.Error())
	}
}

func (r *egressRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func createEgressRuleRequest(plan egressRuleModel) (*egressv1.CreateEgressRuleRequest, error) {
	matcher, err := egressMatcherFromModel(plan)
	if err != nil {
		return nil, err
	}
	effect, err := egressEffectFromModel(plan)
	if err != nil {
		return nil, err
	}
	return &egressv1.CreateEgressRuleRequest{OrganizationId: plan.OrganizationID.ValueString(), Name: plan.Name.ValueString(), Description: stringValue(plan.Description), Matcher: matcher, Effect: effect}, nil
}

type normalizeEgressMethodsPlanModifier struct{}

func normalizeEgressMethodsPlan() planmodifier.List {
	return normalizeEgressMethodsPlanModifier{}
}

func (m normalizeEgressMethodsPlanModifier) Description(context.Context) string {
	return "Normalizes configured egress rule methods to uppercase."
}

func (m normalizeEgressMethodsPlanModifier) MarkdownDescription(context.Context) string {
	return "Normalizes configured egress rule methods to uppercase."
}

func (m normalizeEgressMethodsPlanModifier) PlanModifyList(_ context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	elements := req.ConfigValue.Elements()
	normalized := make([]attr.Value, 0, len(elements))
	for _, element := range elements {
		method, ok := element.(types.String)
		if !ok || method.IsNull() || method.IsUnknown() {
			return
		}
		normalized = append(normalized, types.StringValue(strings.ToUpper(strings.TrimSpace(method.ValueString()))))
	}

	value, diagnostics := types.ListValue(types.StringType, normalized)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.PlanValue = value
}

func egressMatcherFromModel(model egressRuleModel) (*egressv1.EgressRuleMatcher, error) {
	methods := make([]string, 0, len(model.Methods))
	for _, method := range model.Methods {
		if method.IsNull() || method.IsUnknown() || strings.TrimSpace(method.ValueString()) == "" {
			return nil, fmt.Errorf("methods cannot contain empty values")
		}
		methods = append(methods, strings.ToUpper(strings.TrimSpace(method.ValueString())))
	}
	ports := make([]int32, 0, len(model.Ports))
	for _, port := range model.Ports {
		if !port.IsNull() && !port.IsUnknown() {
			ports = append(ports, port.ValueInt32())
		}
	}
	return &egressv1.EgressRuleMatcher{DomainPattern: model.DomainPattern.ValueString(), Ports: ports, Methods: methods, PathPattern: stringValue(model.PathPattern)}, nil
}

func egressEffectFromModel(model egressRuleModel) (*egressv1.EgressRuleEffect, error) {
	action, err := egressActionFromString(stringValue(model.Action))
	if err != nil {
		return nil, err
	}
	effect := &egressv1.EgressRuleEffect{Action: action.Enum()}
	seenHeaders := make(map[string]struct{}, len(model.Headers))
	for _, headerModel := range model.Headers {
		header, err := egressHeaderFromModel(headerModel)
		if err != nil {
			return nil, err
		}
		headerKey := egressHeaderNameKey(header.GetName())
		if _, ok := seenHeaders[headerKey]; ok {
			return nil, fmt.Errorf("duplicate header name %s", header.GetName())
		}
		seenHeaders[headerKey] = struct{}{}
		effect.Inject = append(effect.Inject, header)
	}
	return effect, nil
}

func egressHeaderFromModel(model egressHeaderModel) (*egressv1.EgressRuleHeader, error) {
	header := &egressv1.EgressRuleHeader{Name: model.Name.ValueString()}
	scheme, err := egressSchemeFromString(stringValue(model.Scheme))
	if err != nil {
		return nil, err
	}
	header.Scheme = scheme
	hasValue := !model.Value.IsNull() && !model.Value.IsUnknown()
	hasSecretID := !model.SecretID.IsNull() && !model.SecretID.IsUnknown()
	if hasValue && hasSecretID {
		return nil, fmt.Errorf("header %s requires exactly one of value or secret_id", model.Name.ValueString())
	}
	if hasValue {
		header.Credential = &egressv1.EgressRuleHeader_Value{Value: model.Value.ValueString()}
		return header, nil
	}
	if hasSecretID {
		header.Credential = &egressv1.EgressRuleHeader_SecretId{SecretId: model.SecretID.ValueString()}
		return header, nil
	}
	return nil, fmt.Errorf("header %s requires value or secret_id", model.Name.ValueString())
}

func egressRuleState(rule *egressv1.EgressRule, prior egressRuleModel) egressRuleModel {
	matcher := rule.GetMatcher()
	ports := make([]types.Int32, 0, len(matcher.GetPorts()))
	for _, port := range matcher.GetPorts() {
		ports = append(ports, types.Int32Value(port))
	}
	methods := make([]types.String, 0, len(matcher.GetMethods()))
	for _, method := range matcher.GetMethods() {
		methods = append(methods, types.StringValue(method))
	}
	return egressRuleModel{ID: types.StringValue(rule.GetMeta().GetId()), OrganizationID: types.StringValue(rule.GetOrganizationId()), Name: types.StringValue(rule.GetName()), Description: optionalString(rule.GetDescription()), DomainPattern: types.StringValue(matcher.GetDomainPattern()), Ports: ports, Methods: methods, PathPattern: optionalString(matcher.GetPathPattern()), Action: types.StringValue(egressActionToString(rule.GetEffect().GetAction())), Headers: egressHeadersState(rule.GetEffect().GetInject(), prior.Headers)}
}

func egressHeadersState(headers []*egressv1.EgressRuleHeader, prior []egressHeaderModel) []egressHeaderModel {
	priorValues := make(map[string]types.String, len(prior))
	for _, header := range prior {
		if header.Value.IsNull() || header.Value.IsUnknown() {
			continue
		}
		priorValues[egressHeaderStateKey(header.Name.ValueString(), stringValue(header.Scheme))] = header.Value
	}
	state := make([]egressHeaderModel, 0, len(headers))
	for _, header := range headers {
		scheme := egressSchemeToString(header.GetScheme())
		entry := egressHeaderModel{Name: types.StringValue(header.GetName()), Scheme: optionalString(scheme), Value: types.StringNull(), SecretID: types.StringNull()}
		if header.GetSecretId() != "" {
			entry.SecretID = types.StringValue(header.GetSecretId())
		} else if priorValue, ok := priorValues[egressHeaderStateKey(header.GetName(), scheme)]; ok {
			entry.Value = priorValue
		}
		state = append(state, entry)
	}
	return state
}

func egressHeaderStateKey(name string, scheme string) string {
	return egressHeaderNameKey(name) + "|" + strings.ToLower(strings.TrimSpace(scheme))
}

func egressHeaderNameKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func egressActionFromString(value string) (egressv1.EgressRuleAction, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "allow":
		return egressv1.EgressRuleAction_EGRESS_RULE_ACTION_ALLOW, nil
	case "deny":
		return egressv1.EgressRuleAction_EGRESS_RULE_ACTION_DENY, nil
	default:
		return egressv1.EgressRuleAction_EGRESS_RULE_ACTION_UNSPECIFIED, fmt.Errorf("action must be allow or deny")
	}
}

func egressActionToString(action egressv1.EgressRuleAction) string {
	switch action {
	case egressv1.EgressRuleAction_EGRESS_RULE_ACTION_ALLOW:
		return "allow"
	case egressv1.EgressRuleAction_EGRESS_RULE_ACTION_DENY:
		return "deny"
	case egressv1.EgressRuleAction_EGRESS_RULE_ACTION_UNSPECIFIED:
		return ""
	default:
		panic("unsupported egress rule action " + action.String())
	}
}

func egressSchemeFromString(value string) (egressv1.HeaderAuthScheme, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return egressv1.HeaderAuthScheme_HEADER_AUTH_SCHEME_UNSPECIFIED, nil
	case "bearer":
		return egressv1.HeaderAuthScheme_HEADER_AUTH_SCHEME_BEARER, nil
	case "basic":
		return egressv1.HeaderAuthScheme_HEADER_AUTH_SCHEME_BASIC, nil
	default:
		return egressv1.HeaderAuthScheme_HEADER_AUTH_SCHEME_UNSPECIFIED, fmt.Errorf("scheme must be bearer or basic")
	}
}

func egressSchemeToString(scheme egressv1.HeaderAuthScheme) string {
	switch scheme {
	case egressv1.HeaderAuthScheme_HEADER_AUTH_SCHEME_BEARER:
		return "bearer"
	case egressv1.HeaderAuthScheme_HEADER_AUTH_SCHEME_BASIC:
		return "basic"
	case egressv1.HeaderAuthScheme_HEADER_AUTH_SCHEME_UNSPECIFIED:
		return ""
	default:
		panic("unsupported egress header scheme " + scheme.String())
	}
}

func stringPtr(value string) *string { return &value }
