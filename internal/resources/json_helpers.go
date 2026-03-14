package resources

import (
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func marshalConfig(config any) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	raw, err := json.Marshal(config)
	if err != nil {
		diags.AddError("Failed to Serialize Config", err.Error())
		return "", diags
	}
	return string(raw), diags
}

func configStateValue(config types.String, payload any) (types.String, diag.Diagnostics) {
	if config.IsNull() || config.IsUnknown() {
		return config, nil
	}

	responseJSON, diags := marshalConfig(payload)
	if diags.HasError() {
		return config, diags
	}

	original := config.ValueString()
	if jsonSemanticallyEqual(original, responseJSON) {
		return config, nil
	}

	return types.StringValue(responseJSON), diags
}

func jsonSemanticallyEqual(a, b string) bool {
	var objA any
	var objB any
	if err := json.Unmarshal([]byte(a), &objA); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &objB); err != nil {
		return false
	}
	return reflect.DeepEqual(objA, objB)
}
