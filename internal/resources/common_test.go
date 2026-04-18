package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBoolPointer(t *testing.T) {
	pointer := boolPointer(types.BoolValue(true))
	if pointer == nil || *pointer != true {
		t.Fatalf("expected pointer with value, got %v", pointer)
	}

	if ptr := boolPointer(types.BoolNull()); ptr != nil {
		t.Fatalf("expected nil pointer for null bool")
	}

	if ptr := boolPointer(types.BoolUnknown()); ptr != nil {
		t.Fatalf("expected nil pointer for unknown bool")
	}
}

func TestOptionalString(t *testing.T) {
	if v := optionalString(""); !v.IsNull() {
		t.Fatalf("expected null value for empty string")
	}

	v := optionalString("hello")
	if v.IsNull() || v.ValueString() != "hello" {
		t.Fatalf("unexpected optional string value: %v", v)
	}
}

func TestUpdateStringPointer(t *testing.T) {
	if ptr := updateStringPointer(types.StringUnknown(), types.StringValue("prior")); ptr != nil {
		t.Fatalf("expected nil pointer for unknown plan")
	}

	if ptr := updateStringPointer(types.StringNull(), types.StringNull()); ptr != nil {
		t.Fatalf("expected nil pointer for null plan with null prior")
	}

	if ptr := updateStringPointer(types.StringNull(), types.StringValue("prior")); ptr == nil || *ptr != "" {
		t.Fatalf("expected empty string pointer for cleared value, got %v", ptr)
	}

	if ptr := updateStringPointer(types.StringValue("value"), types.StringValue("prior")); ptr == nil || *ptr != "value" {
		t.Fatalf("expected pointer with value, got %v", ptr)
	}
}

func TestPreserveSensitiveString(t *testing.T) {
	fallback := types.StringValue("fallback")
	if got := preserveSensitiveString(fallback, ""); got.ValueString() != "fallback" {
		t.Fatalf("expected fallback value, got %v", got)
	}

	if got := preserveSensitiveString(fallback, "secret"); got.ValueString() != "secret" {
		t.Fatalf("expected api value, got %v", got)
	}
}

func TestStringListFromPlan(t *testing.T) {
	ctx := context.Background()
	if values, diags := stringListFromPlan(ctx, types.ListNull(types.StringType)); diags.HasError() || values != nil {
		t.Fatalf("expected nil values for null list, got %v (%v)", values, diags)
	}
	if values, diags := stringListFromPlan(ctx, types.ListUnknown(types.StringType)); diags.HasError() || values != nil {
		t.Fatalf("expected nil values for unknown list, got %v (%v)", values, diags)
	}
	list, diags := types.ListValueFrom(ctx, types.StringType, []string{"alpha", "beta"})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	values, diags := stringListFromPlan(ctx, list)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(values) != 2 || values[0] != "alpha" || values[1] != "beta" {
		t.Fatalf("unexpected values: %#v", values)
	}
}

func TestStringListToState(t *testing.T) {
	ctx := context.Background()
	list, diags := stringListToState(ctx, nil, types.ListNull(types.StringType))
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !list.IsNull() {
		t.Fatalf("expected null list state, got %v", list)
	}

	prior, diags := types.ListValueFrom(ctx, types.StringType, []string{"prior"})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	list, diags = stringListToState(ctx, nil, prior)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if list.IsNull() {
		t.Fatalf("expected empty list state, got null")
	}
	values := []string{}
	if diags := list.ElementsAs(ctx, &values, false); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(values) != 0 {
		t.Fatalf("expected empty list, got %#v", values)
	}

	list, diags = stringListToState(ctx, []string{"one", "two"}, types.ListNull(types.StringType))
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	values = []string{}
	if diags := list.ElementsAs(ctx, &values, false); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(values) != 2 || values[0] != "one" || values[1] != "two" {
		t.Fatalf("unexpected list values: %#v", values)
	}
}
