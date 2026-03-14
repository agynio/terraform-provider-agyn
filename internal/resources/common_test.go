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
	if v := preserveOrApplyString(types.StringNull(), &str); !v.IsNull() {
		t.Fatalf("expected null for null prior, got %v", v)
	}
	if v := preserveOrApplyString(types.StringUnknown(), &str); !v.IsNull() {
		t.Fatalf("expected null for unknown prior, got %v", v)
	}
	v := preserveOrApplyString(types.StringValue("old"), &str)
	if v.ValueString() != "hello" {
		t.Fatalf("expected hello, got %s", v.ValueString())
	}
	v = preserveOrApplyString(types.StringValue("old"), nil)
	if !v.IsNull() {
		t.Fatalf("expected null for nil API, got %v", v)
	}
}

func TestPreserveOrApplyInt64(t *testing.T) {
	val := int64(42)
	if v := preserveOrApplyInt64(types.Int64Null(), &val); !v.IsNull() {
		t.Fatalf("expected null for null prior")
	}
	v := preserveOrApplyInt64(types.Int64Value(0), &val)
	if v.ValueInt64() != 42 {
		t.Fatalf("expected 42, got %d", v.ValueInt64())
	}
	zero := int64(0)
	v = preserveOrApplyInt64(types.Int64Value(10), &zero)
	if v.ValueInt64() != 0 {
		t.Fatalf("expected 0, got %d", v.ValueInt64())
	}
}

func TestPreserveOrApplyBool(t *testing.T) {
	val := true
	if v := preserveOrApplyBool(types.BoolNull(), &val); !v.IsNull() {
		t.Fatalf("expected null for null prior")
	}
	v := preserveOrApplyBool(types.BoolValue(false), &val)
	if !v.ValueBool() {
		t.Fatalf("expected true, got false")
	}
}
