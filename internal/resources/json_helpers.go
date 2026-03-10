package resources

import (
	"encoding/json"

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
	value, diags := marshalConfig(payload)
	if diags.HasError() {
		return config, diags
	}
	return types.StringValue(value), diags
}
