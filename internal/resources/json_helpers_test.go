package resources

import (
	"encoding/json"
	"reflect"
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

func TestProjectJSONKeys(t *testing.T) {
	tests := []struct {
		name      string
		reference string
		source    string
		expected  map[string]any
		wantErr   bool
	}{
		{
			name:      "subset keys",
			reference: `{"a":1}`,
			source:    `{"a":2,"b":3}`,
			expected:  map[string]any{"a": float64(2)},
		},
		{
			name:      "missing source key",
			reference: `{"a":1,"b":2}`,
			source:    `{"a":3}`,
			expected:  map[string]any{"a": float64(3), "b": float64(2)},
		},
		{
			name:      "empty objects",
			reference: `{}`,
			source:    `{}`,
			expected:  map[string]any{},
		},
		{
			name:      "invalid reference",
			reference: "{invalid",
			source:    `{"a":1}`,
			wantErr:   true,
		},
		{
			name:      "invalid source",
			reference: `{"a":1}`,
			source:    "{invalid",
			wantErr:   true,
		},
	}

	for _, test := range tests {
		result, err := projectJSONKeys(test.reference, test.source)
		if test.wantErr {
			if err == nil {
				t.Fatalf("%s: expected error", test.name)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", test.name, err)
		}
		var got map[string]any
		if err := json.Unmarshal([]byte(result), &got); err != nil {
			t.Fatalf("%s: invalid json result: %v", test.name, err)
		}
		if !reflect.DeepEqual(got, test.expected) {
			t.Fatalf("%s: expected %v, got %v", test.name, test.expected, got)
		}
	}
}

func TestConfigStateValueProjection(t *testing.T) {
	type payload struct {
		Name       string `json:"name"`
		Role       string `json:"role"`
		DebounceMs int64  `json:"debounceMs"`
	}

	original := types.StringValue(`{"name":"foo","role":"bar"}`)
	value, diags := configStateValue(original, payload{Name: "foo", Role: "bar"})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if value.ValueString() != original.ValueString() {
		t.Fatalf("expected original JSON, got: %s", value.ValueString())
	}
}
