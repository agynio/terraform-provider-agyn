package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStringPointer(t *testing.T) {
	pointer := stringPointer(types.StringValue("value"))
	if pointer == nil || *pointer != "value" {
		t.Fatalf("expected pointer with value, got %v", pointer)
	}

	if ptr := stringPointer(types.StringNull()); ptr != nil {
		t.Fatalf("expected nil pointer for null string")
	}

	if ptr := stringPointer(types.StringUnknown()); ptr != nil {
		t.Fatalf("expected nil pointer for unknown string")
	}
}

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
	if v := optionalString(nil); !v.IsNull() {
		t.Fatalf("expected null value for nil pointer")
	}

	str := "hello"
	v := optionalString(&str)
	if v.IsNull() || v.ValueString() != "hello" {
		t.Fatalf("unexpected optional string value: %v", v)
	}
}

func TestOptionalBool(t *testing.T) {
	if v := optionalBool(nil); !v.IsNull() {
		t.Fatalf("expected null value for nil pointer")
	}

	value := true
	v := optionalBool(&value)
	if v.IsNull() || !v.ValueBool() {
		t.Fatalf("unexpected optional bool value: %v", v)
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
	if got := preserveSensitiveString(fallback, nil); got.ValueString() != "fallback" {
		t.Fatalf("expected fallback value, got %v", got)
	}

	apiValue := "secret"
	if got := preserveSensitiveString(fallback, &apiValue); got.ValueString() != "secret" {
		t.Fatalf("expected api value, got %v", got)
	}
}
