package resources

import "github.com/hashicorp/terraform-plugin-framework/types"

func stringPointer(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	value := v.ValueString()
	return &value
}

func updateStringPointer(plan types.String, prior types.String) *string {
	if plan.IsUnknown() {
		return nil
	}
	if plan.IsNull() {
		if prior.IsNull() || prior.IsUnknown() {
			return nil
		}
		empty := ""
		return &empty
	}
	value := plan.ValueString()
	return &value
}

func boolPointer(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	value := v.ValueBool()
	return &value
}

func optionalString(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

func preserveSensitiveString(fallback types.String, apiValue *string) types.String {
	if apiValue == nil {
		return fallback
	}
	return types.StringValue(*apiValue)
}

func optionalBool(v *bool) types.Bool {
	if v == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*v)
}

const httpStatusNotFound = 404
