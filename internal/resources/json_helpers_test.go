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
	type orderingPayload struct {
		A int `json:"a"`
		B int `json:"b"`
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

	original := types.StringValue(`{"b":1,"a":2}`)
	value, diags = configStateValue(original, orderingPayload{A: 2, B: 1})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if value.ValueString() != original.ValueString() {
		t.Fatalf("expected original JSON ordering, got: %s", value.ValueString())
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

func TestJSONSemanticallyEqual(t *testing.T) {
	tests := []struct {
		name     string
		left     string
		right    string
		expected bool
	}{
		{
			name:     "different ordering",
			left:     `{"b":1,"a":2}`,
			right:    `{"a":2,"b":1}`,
			expected: true,
		},
		{
			name:     "same ordering",
			left:     `{"a":2,"b":1}`,
			right:    `{"a":2,"b":1}`,
			expected: true,
		},
		{
			name:     "different content",
			left:     `{"a":1}`,
			right:    `{"a":2}`,
			expected: false,
		},
		{
			name:     "invalid left",
			left:     "{invalid",
			right:    `{"a":1}`,
			expected: false,
		},
		{
			name:     "invalid right",
			left:     `{"a":1}`,
			right:    "{invalid",
			expected: false,
		},
		{
			name:     "both invalid",
			left:     "{invalid",
			right:    "{invalid",
			expected: false,
		},
	}

	for _, test := range tests {
		if jsonSemanticallyEqual(test.left, test.right) != test.expected {
			t.Fatalf("%s: expected %v", test.name, test.expected)
		}
	}
}
