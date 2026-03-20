package resources

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNormalizeJSONState(t *testing.T) {
	apiConfig := `{"a":2,"b":1}`
	apiConfigExtra := `{"a":1,"b":2,"c":3}`
	apiConfigDifferent := `{"a":3}`

	tests := []struct {
		name     string
		config   types.String
		apiValue string
		want     types.String
	}{
		{
			name:     "api nil config null",
			config:   types.StringNull(),
			apiValue: "",
			want:     types.StringNull(),
		},
		{
			name:     "api nil config unknown",
			config:   types.StringUnknown(),
			apiValue: "",
			want:     types.StringNull(),
		},
		{
			name:     "api nil config set",
			config:   types.StringValue(`{"a":1}`),
			apiValue: "",
			want:     types.StringValue(`{"a":1}`),
		},
		{
			name:     "api value config null",
			config:   types.StringNull(),
			apiValue: apiConfig,
			want:     types.StringValue(apiConfig),
		},
		{
			name:     "api value config unknown",
			config:   types.StringUnknown(),
			apiValue: apiConfig,
			want:     types.StringValue(apiConfig),
		},
		{
			name:     "api semantically equal",
			config:   types.StringValue(`{"b":1,"a":2}`),
			apiValue: apiConfig,
			want:     types.StringValue(`{"b":1,"a":2}`),
		},
		{
			name:     "api adds keys",
			config:   types.StringValue(`{"a":1,"b":2}`),
			apiValue: apiConfigExtra,
			want:     types.StringValue(`{"a":1,"b":2}`),
		},
		{
			name:     "api differs",
			config:   types.StringValue(`{"a":1}`),
			apiValue: apiConfigDifferent,
			want:     types.StringValue(apiConfigDifferent),
		},
	}

	for _, test := range tests {
		result, diags := normalizeJSONState(test.config, test.apiValue)
		if diags.HasError() {
			t.Fatalf("%s: unexpected diagnostics: %v", test.name, diags)
		}
		assertStringValue(t, test.name, result, test.want)
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
			name:      "server added keys",
			reference: `{"a":1,"b":2}`,
			source:    `{"a":1,"b":2,"c":3}`,
			expected:  map[string]any{"a": float64(1), "b": float64(2)},
		},
		{
			name:      "nested objects",
			reference: `{"a":{"b":1}}`,
			source:    `{"a":{"b":1,"c":2},"d":3}`,
			expected:  map[string]any{"a": map[string]any{"b": float64(1), "c": float64(2)}},
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

func assertStringValue(t *testing.T, name string, got types.String, want types.String) {
	if want.IsNull() {
		if !got.IsNull() {
			t.Fatalf("%s: expected null", name)
		}
		return
	}
	if want.IsUnknown() {
		if !got.IsUnknown() {
			t.Fatalf("%s: expected unknown", name)
		}
		return
	}
	if got.IsNull() || got.IsUnknown() {
		t.Fatalf("%s: expected value %q", name, want.ValueString())
	}
	if got.ValueString() != want.ValueString() {
		t.Fatalf("%s: expected %q, got %q", name, want.ValueString(), got.ValueString())
	}
}
