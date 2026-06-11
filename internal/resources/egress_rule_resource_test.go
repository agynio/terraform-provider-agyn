package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	egressv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/egress/v1"
)

func TestEgressMatcherFromModel(t *testing.T) {
	matcher, err := egressMatcherFromModel(egressRuleModel{
		DomainPattern: types.StringValue("api.example.com"),
		Ports:         []types.Int32{types.Int32Value(443)},
		Methods:       []types.String{types.StringValue("get")},
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

func TestEgressActionFromStringRejectsInvalidValue(t *testing.T) {
	if _, err := egressActionFromString("block"); err == nil {
		t.Fatalf("expected invalid action error")
	}
}
