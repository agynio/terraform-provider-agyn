package resources

import (
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func normalizeJSONState(config types.String, apiValue *string) (types.String, diag.Diagnostics) {
	var diags diag.Diagnostics
	if apiValue == nil {
		if config.IsNull() || config.IsUnknown() {
			return types.StringNull(), diags
		}
		return config, diags
	}

	if config.IsNull() || config.IsUnknown() {
		return types.StringValue(*apiValue), diags
	}

	original := config.ValueString()
	if jsonSemanticallyEqual(original, *apiValue) {
		return config, diags
	}

	projected, err := projectJSONKeys(original, *apiValue)
	if err == nil && jsonSemanticallyEqual(original, projected) {
		return config, diags
	}

	return types.StringValue(*apiValue), diags
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

func projectJSONKeys(reference, source string) (string, error) {
	var refObj map[string]any
	var srcObj map[string]any
	if err := json.Unmarshal([]byte(reference), &refObj); err != nil {
		return "", err
	}
	if err := json.Unmarshal([]byte(source), &srcObj); err != nil {
		return "", err
	}
	projected := make(map[string]any, len(refObj))
	for key, refValue := range refObj {
		if value, ok := srcObj[key]; ok {
			projected[key] = value
			continue
		}
		projected[key] = refValue
	}
	raw, err := json.Marshal(projected)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
