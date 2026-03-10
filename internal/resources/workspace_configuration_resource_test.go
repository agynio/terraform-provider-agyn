package resources

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRawMessageFromString(t *testing.T) {
	raw, diags := rawMessageFromString(types.StringValue("2"), "cpu_limit")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if raw == nil || string(*raw) != "2" {
		t.Fatalf("unexpected raw message: %v", raw)
	}

	raw, diags = rawMessageFromString(types.StringValue("4Gi"), "memory_limit")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if raw == nil || string(*raw) != `"4Gi"` {
		t.Fatalf("unexpected raw message: %v", raw)
	}

	_, diags = rawMessageFromString(types.StringValue(`{"value":1}`), "cpu_limit")
	if !diags.HasError() {
		t.Fatalf("expected diagnostics for object JSON")
	}
}

func TestStringFromRawMessage(t *testing.T) {
	numeric := json.RawMessage("2")
	value, diags := stringFromRawMessage(&numeric, "cpu_limit")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if value.ValueString() != "2" {
		t.Fatalf("unexpected value: %s", value.ValueString())
	}

	text := json.RawMessage(`"4Gi"`)
	value, diags = stringFromRawMessage(&text, "memory_limit")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if value.ValueString() != "4Gi" {
		t.Fatalf("unexpected value: %s", value.ValueString())
	}

	nullValue := json.RawMessage("null")
	value, diags = stringFromRawMessage(&nullValue, "cpu_limit")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !value.IsNull() {
		t.Fatalf("expected null value")
	}

	invalid := json.RawMessage("{}")
	_, diags = stringFromRawMessage(&invalid, "cpu_limit")
	if !diags.HasError() {
		t.Fatalf("expected diagnostics for object JSON")
	}
}

func TestNormalizedRawMessageRoundTrip(t *testing.T) {
	input := jsontypes.NewNormalizedValue(`{"packages":["go"]}`)
	raw, diags := rawMessageFromNormalized(input)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if raw == nil || string(*raw) != `{"packages":["go"]}` {
		t.Fatalf("unexpected raw message: %v", raw)
	}

	roundTrip := normalizedFromRawMessage(raw)
	if roundTrip.IsNull() {
		t.Fatalf("expected normalized value")
	}
	if roundTrip.ValueString() != `{"packages":["go"]}` {
		t.Fatalf("unexpected normalized value: %s", roundTrip.ValueString())
	}

	if raw, diags = rawMessageFromNormalized(jsontypes.NewNormalizedNull()); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if raw != nil {
		t.Fatalf("expected nil raw message")
	}

	if normalizedFromRawMessage(nil).IsNull() == false {
		t.Fatalf("expected null normalized value")
	}
}
