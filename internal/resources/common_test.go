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
