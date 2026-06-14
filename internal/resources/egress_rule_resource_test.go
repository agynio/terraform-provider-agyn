package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	egressv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/egress/v1"
)

func TestEgressMatcherFromModel(t *testing.T) {
	ports := types.ListValueMust(types.Int32Type, []attr.Value{types.Int32Value(443)})
	methods := newEgressMethodsListValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("get")}))

	matcher, err := egressMatcherFromModel(egressRuleModel{
		DomainPattern: types.StringValue("api.example.com"),
		Ports:         ports,
		Methods:       methods,
		PathPattern:   types.StringValue("/v1/*"),
	})
	if err != nil {
		t.Fatalf("unexpected matcher error: %v", err)
	}
	if matcher.GetDomainPattern() != "api.example.com" {
		t.Fatalf("unexpected domain pattern %q", matcher.GetDomainPattern())
	}
	if len(matcher.GetPorts()) != 1 || matcher.GetPorts()[0] != 443 {
		t.Fatalf("unexpected ports %#v", matcher.GetPorts())
	}
	if len(matcher.GetMethods()) != 1 || matcher.GetMethods()[0] != "GET" {
		t.Fatalf("unexpected methods %#v", matcher.GetMethods())
	}
	if matcher.GetPathPattern() != "/v1/*" {
		t.Fatalf("unexpected path pattern %q", matcher.GetPathPattern())
	}
}

func TestEgressMatcherFromModelIgnoresUnknownDefaultLists(t *testing.T) {
	matcher, err := egressMatcherFromModel(egressRuleModel{
		DomainPattern: types.StringValue("api.example.com"),
		Ports:         types.ListUnknown(types.Int32Type),
		Methods:       newEgressMethodsListValue(types.ListUnknown(types.StringType)),
	})
	if err != nil {
		t.Fatalf("unexpected matcher error: %v", err)
	}
	if len(matcher.GetPorts()) != 0 {
		t.Fatalf("unexpected ports %#v", matcher.GetPorts())
	}
	if len(matcher.GetMethods()) != 0 {
		t.Fatalf("unexpected methods %#v", matcher.GetMethods())
	}
}

func TestEgressEffectFromModel(t *testing.T) {
	effect, err := egressEffectFromModel(egressRuleModel{
		Action: types.StringValue("deny"),
		Headers: []egressHeaderModel{{
			Name:     types.StringValue("Authorization"),
			Scheme:   types.StringValue("bearer"),
			Value:    types.StringNull(),
			SecretID: types.StringValue("secret-id"),
		}},
	})
	if err != nil {
		t.Fatalf("unexpected effect error: %v", err)
	}
	if effect.GetAction() != egressv1.EgressRuleAction_EGRESS_RULE_ACTION_DENY {
		t.Fatalf("unexpected action %s", effect.GetAction())
	}
	if len(effect.GetInject()) != 1 {
		t.Fatalf("unexpected headers %#v", effect.GetInject())
	}
	header := effect.GetInject()[0]
	if header.GetName() != "Authorization" {
		t.Fatalf("unexpected header name %q", header.GetName())
	}
	if header.GetScheme() != egressv1.HeaderAuthScheme_HEADER_AUTH_SCHEME_BEARER {
		t.Fatalf("unexpected scheme %s", header.GetScheme())
	}
	if header.GetSecretId() != "secret-id" {
		t.Fatalf("unexpected secret id %q", header.GetSecretId())
	}
}

func TestEgressHeaderFromModelRequiresExactlyOneCredential(t *testing.T) {
	_, err := egressHeaderFromModel(egressHeaderModel{Name: types.StringValue("Authorization"), Value: types.StringNull(), SecretID: types.StringNull()})
	if err == nil {
		t.Fatalf("expected missing credential error")
	}
	_, err = egressHeaderFromModel(egressHeaderModel{Name: types.StringValue("Authorization"), Value: types.StringValue("literal"), SecretID: types.StringValue("secret-id")})
	if err == nil {
		t.Fatalf("expected conflicting credential error")
	}
}

func TestEgressEffectFromModelRejectsDuplicateHeaders(t *testing.T) {
	_, err := egressEffectFromModel(egressRuleModel{Headers: []egressHeaderModel{
		{Name: types.StringValue("Authorization"), Value: types.StringValue("one"), SecretID: types.StringNull()},
		{Name: types.StringValue(" authorization "), Value: types.StringValue("two"), SecretID: types.StringNull()},
	}})
	if err == nil {
		t.Fatalf("expected duplicate header error")
	}
}

func TestEgressMethodsStateNormalizesMethods(t *testing.T) {
	state := egressMethodsState([]string{"get", " post "})

	elements := state.Elements()
	if len(elements) != 2 {
		t.Fatalf("unexpected method count %d", len(elements))
	}
	if elements[0].(types.String).ValueString() != "GET" || elements[1].(types.String).ValueString() != "POST" {
		t.Fatalf("unexpected normalized methods %#v", elements)
	}
}

func TestEgressHeadersStatePreservesLiteralValuesByHeaderKey(t *testing.T) {
	state := egressHeadersState([]*egressv1.EgressRuleHeader{
		{Name: "X-Second", Credential: &egressv1.EgressRuleHeader_Value{}},
		{Name: "X-First", Scheme: egressv1.HeaderAuthScheme_HEADER_AUTH_SCHEME_BEARER, Credential: &egressv1.EgressRuleHeader_Value{}},
	}, []egressHeaderModel{
		{Name: types.StringValue("X-First"), Scheme: types.StringValue("bearer"), Value: types.StringValue("first"), SecretID: types.StringNull()},
		{Name: types.StringValue("X-Second"), Scheme: types.StringNull(), Value: types.StringValue("second"), SecretID: types.StringNull()},
	})
	if state[0].Value.ValueString() != "second" || state[1].Value.ValueString() != "first" {
		t.Fatalf("unexpected preserved values: %#v", state)
	}
}

func TestEgressActionFromStringRejectsInvalidValue(t *testing.T) {
	if _, err := egressActionFromString("block"); err == nil {
		t.Fatalf("expected invalid action error")
	}
}

func TestEgressMethodsListSemanticEquals(t *testing.T) {
	prior := newEgressMethodsListValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("get")}))
	state := newEgressMethodsListValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("GET")}))

	equal, diagnostics := prior.ListSemanticEquals(context.Background(), state)
	if diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if !equal {
		t.Fatalf("expected method lists to be semantically equal")
	}
}

func TestEgressMethodsListSemanticEqualsRejectsDifferentMethods(t *testing.T) {
	prior := newEgressMethodsListValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("post")}))
	state := newEgressMethodsListValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("GET")}))

	equal, diagnostics := prior.ListSemanticEquals(context.Background(), state)
	if diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if equal {
		t.Fatalf("expected method lists to differ")
	}
}
