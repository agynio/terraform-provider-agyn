package resources

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEgressRulePartsRejectsNoopRule(t *testing.T) {
	plan := egressRuleModel{
		DomainPattern: types.StringValue("*.github.com"),
		Action:        types.StringNull(),
	}

	var summary, detail string
	_, _, stop := egressRuleParts(plan, func(gotSummary, gotDetail string) {
		summary = gotSummary
		detail = gotDetail
	})

	if !stop {
		t.Fatalf("expected no-op rule to stop")
	}
	if summary != "Invalid egress rule" || !strings.Contains(detail, "requires an action or at least one injected header") {
		t.Fatalf("unexpected diagnostic: %q %q", summary, detail)
	}
}

func TestEgressRulePartsAllowsInjectionOnlyRule(t *testing.T) {
	plan := egressRuleModel{
		DomainPattern: types.StringValue("api.github.com"),
		Action:        types.StringNull(),
		Headers: []egressHeaderModel{{
			Name:  types.StringValue("Authorization"),
			Value: types.StringValue("token"),
		}},
	}

	_, effect, stop := egressRuleParts(plan, func(summary, detail string) {
		t.Fatalf("unexpected diagnostic: %s %s", summary, detail)
	})

	if stop {
		t.Fatalf("expected injection-only rule to be valid")
	}
	if effect.GetAction() != 0 {
		t.Fatalf("expected no action, got %s", effect.GetAction())
	}
	if len(effect.GetInject()) != 1 || effect.GetInject()[0].GetValue() != "token" {
		t.Fatalf("unexpected inject headers: %#v", effect.GetInject())
	}
}

func TestEgressRulePartsRejectsSecretAndValue(t *testing.T) {
	plan := egressRuleModel{
		DomainPattern: types.StringValue("api.github.com"),
		Action:        types.StringValue("allow"),
		Headers: []egressHeaderModel{{
			Name:     types.StringValue("Authorization"),
			Value:    types.StringValue("token"),
			SecretID: types.StringValue("secret-id"),
		}},
	}

	var summary, detail string
	_, _, stop := egressRuleParts(plan, func(gotSummary, gotDetail string) {
		summary = gotSummary
		detail = gotDetail
	})

	if !stop {
		t.Fatalf("expected header with value and secret_id to stop")
	}
	if summary != "Invalid injected header" || !strings.Contains(detail, "exactly one of value or secret_id") {
		t.Fatalf("unexpected diagnostic: %q %q", summary, detail)
	}
}
