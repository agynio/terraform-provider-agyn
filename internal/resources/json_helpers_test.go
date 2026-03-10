package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMarshalConfig(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	value, diags := marshalConfig(payload{Name: "agent"})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if value != `{"name":"agent"}` {
		t.Fatalf("unexpected value: %s", value)
	}

	_, diags = marshalConfig(make(chan int))
	if !diags.HasError() {
		t.Fatalf("expected marshal error")
	}
}

func TestConfigStateValue(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	value, diags := configStateValue(types.StringNull(), payload{Name: "agent"})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !value.IsNull() {
		t.Fatalf("expected null value")
	}

	value, diags = configStateValue(types.StringUnknown(), payload{Name: "agent"})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !value.IsUnknown() {
		t.Fatalf("expected unknown value")
	}

	value, diags = configStateValue(types.StringValue("set"), payload{Name: "agent"})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if value.ValueString() != `{"name":"agent"}` {
		t.Fatalf("unexpected config value: %s", value.ValueString())
	}

	value, diags = configStateValue(types.StringValue("set"), make(chan int))
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics")
	}
	if value.ValueString() != "set" {
		t.Fatalf("expected original config value on error")
	}
}
