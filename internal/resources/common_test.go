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

func TestPreserveOrApplyString(t *testing.T) {
	str := "hello"
	if v := preserveOrApplyString(types.StringNull(), &str); v.ValueString() != "hello" {
		t.Fatalf("expected hello for null prior, got %v", v)
	}
	if v := preserveOrApplyString(types.StringUnknown(), &str); v.ValueString() != "hello" {
		t.Fatalf("expected hello for unknown prior, got %v", v)
	}
	v := preserveOrApplyString(types.StringValue("old"), &str)
	if v.ValueString() != "old" {
		t.Fatalf("expected old, got %s", v.ValueString())
	}
	v = preserveOrApplyString(types.StringValue("old"), nil)
	if v.ValueString() != "old" {
		t.Fatalf("expected old for nil API, got %s", v.ValueString())
	}
	v = preserveOrApplyString(types.StringNull(), nil)
	if !v.IsNull() {
		t.Fatalf("expected null for null prior and API, got %v", v)
	}
}

func TestPreserveOrApplyInt64(t *testing.T) {
	val := int64(42)
	if v := preserveOrApplyInt64(types.Int64Null(), &val); v.ValueInt64() != 42 {
		t.Fatalf("expected 42 for null prior, got %d", v.ValueInt64())
	}
	if v := preserveOrApplyInt64(types.Int64Unknown(), &val); v.ValueInt64() != 42 {
		t.Fatalf("expected 42 for unknown prior, got %d", v.ValueInt64())
	}
	v := preserveOrApplyInt64(types.Int64Value(10), &val)
	if v.ValueInt64() != 10 {
		t.Fatalf("expected 10, got %d", v.ValueInt64())
	}
	v = preserveOrApplyInt64(types.Int64Value(10), nil)
	if v.ValueInt64() != 10 {
		t.Fatalf("expected 10 for nil API, got %d", v.ValueInt64())
	}
	v = preserveOrApplyInt64(types.Int64Null(), nil)
	if !v.IsNull() {
		t.Fatalf("expected null for null prior and API")
	}
}

func TestPreserveOrApplyBool(t *testing.T) {
	val := true
	if v := preserveOrApplyBool(types.BoolNull(), &val); !v.ValueBool() {
		t.Fatalf("expected true for null prior")
	}
	if v := preserveOrApplyBool(types.BoolUnknown(), &val); !v.ValueBool() {
		t.Fatalf("expected true for unknown prior")
	}
	v := preserveOrApplyBool(types.BoolValue(false), &val)
	if v.ValueBool() {
		t.Fatalf("expected false, got true")
	}
	v = preserveOrApplyBool(types.BoolValue(false), nil)
	if v.ValueBool() {
		t.Fatalf("expected false for nil API, got true")
	}
	v = preserveOrApplyBool(types.BoolNull(), nil)
	if !v.IsNull() {
		t.Fatalf("expected null for null prior and API")
	}
}
