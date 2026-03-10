package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-agyn/internal/teamapi"
)

func TestEnvVarsFromModels(t *testing.T) {
	vars, diags := envVarsFromModels([]envVarModel{
		{
			Name:  types.StringValue("FOO"),
			Value: types.StringValue("bar"),
		},
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(vars) != 1 {
		t.Fatalf("expected 1 env var, got %d", len(vars))
	}
	if vars[0].Name != "FOO" {
		t.Fatalf("unexpected name: %q", vars[0].Name)
	}
	if vars[0].Value == nil || *vars[0].Value != "bar" {
		t.Fatalf("unexpected value: %v", vars[0].Value)
	}
	if vars[0].ValueRef != nil {
		t.Fatalf("expected nil value_ref")
	}

	vars, diags = envVarsFromModels([]envVarModel{
		{
			Name: types.StringValue("TOKEN"),
			ValueRef: &envValueRefModel{
				Kind:  types.StringValue("vault"),
				Mount: types.StringValue("secret"),
				Path:  types.StringValue("path"),
				Key:   types.StringValue("key"),
			},
		},
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(vars) != 1 {
		t.Fatalf("expected 1 env var, got %d", len(vars))
	}
	if vars[0].Value != nil {
		t.Fatalf("expected nil value")
	}
	if vars[0].ValueRef == nil || vars[0].ValueRef.Kind != "vault" {
		t.Fatalf("unexpected value_ref: %#v", vars[0].ValueRef)
	}
	if vars[0].ValueRef.Mount == nil || *vars[0].ValueRef.Mount != "secret" {
		t.Fatalf("unexpected mount: %#v", vars[0].ValueRef.Mount)
	}

	vars, diags = envVarsFromModels([]envVarModel{{
		Name: types.StringNull(),
	}})
	if !diags.HasError() {
		t.Fatalf("expected missing name error")
	}
	if len(vars) != 0 {
		t.Fatalf("expected no env vars, got %d", len(vars))
	}

	vars, diags = envVarsFromModels([]envVarModel{{
		Name:  types.StringValue("BOTH"),
		Value: types.StringValue("value"),
		ValueRef: &envValueRefModel{
			Kind:  types.StringValue("vault"),
			Mount: types.StringValue("secret"),
			Path:  types.StringValue("path"),
			Key:   types.StringValue("key"),
		},
	}})
	if !diags.HasError() {
		t.Fatalf("expected conflicting value error")
	}
	if len(vars) != 0 {
		t.Fatalf("expected no env vars, got %d", len(vars))
	}

	vars, diags = envVarsFromModels([]envVarModel{{
		Name: types.StringValue("EMPTY"),
	}})
	if !diags.HasError() {
		t.Fatalf("expected missing value error")
	}
	if len(vars) != 0 {
		t.Fatalf("expected no env vars, got %d", len(vars))
	}
}

func TestEnvValueRefFromModel(t *testing.T) {
	ref, diags := envValueRefFromModel(envValueRefModel{
		Kind:  types.StringValue("vault"),
		Mount: types.StringValue("secret"),
		Path:  types.StringValue("path"),
		Key:   types.StringValue("key"),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if ref == nil || ref.Kind != "vault" {
		t.Fatalf("unexpected ref: %#v", ref)
	}
	if ref.Mount == nil || *ref.Mount != "secret" {
		t.Fatalf("unexpected mount: %#v", ref.Mount)
	}

	_, diags = envValueRefFromModel(envValueRefModel{
		Kind:  types.StringNull(),
		Mount: types.StringValue("secret"),
		Path:  types.StringValue("path"),
		Key:   types.StringValue("key"),
	})
	if !diags.HasError() {
		t.Fatalf("expected missing kind error")
	}
}

func TestEnvVarModelsFromAPI(t *testing.T) {
	value := "plain"
	models := envVarModelsFromAPI([]teamapi.EnvVar{
		{
			Name:  "PLAIN",
			Value: &value,
		},
		{
			Name: "REF",
			ValueRef: &teamapi.EnvValueRef{
				Kind:  "vault",
				Mount: stringPointer(types.StringValue("secret")),
				Path:  stringPointer(types.StringValue("path")),
				Key:   stringPointer(types.StringValue("key")),
			},
		},
	})
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}
	if models[0].Value.IsNull() || models[0].Value.ValueString() != "plain" {
		t.Fatalf("unexpected model value: %v", models[0].Value)
	}
	if models[1].ValueRef == nil || models[1].ValueRef.Kind.ValueString() != "vault" {
		t.Fatalf("unexpected value_ref: %#v", models[1].ValueRef)
	}
}
