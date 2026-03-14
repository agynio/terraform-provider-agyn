package resources

import "github.com/hashicorp/terraform-plugin-framework/types"

func stringPointer(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	value := v.ValueString()
	return &value
}

func int64Pointer(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	value := v.ValueInt64()
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

func optionalInt64(v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*v)
}

func optionalBool(v *bool) types.Bool {
	if v == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*v)
}

func preserveOrApplyString(prior types.String, apiValue *string) types.String {
	if prior.IsNull() || prior.IsUnknown() {
		return types.StringNull()
	}
	return optionalString(apiValue)
}

func preserveOrApplyInt64(prior types.Int64, apiValue *int64) types.Int64 {
	if prior.IsNull() || prior.IsUnknown() {
		return types.Int64Null()
	}
	return optionalInt64(apiValue)
}

func preserveOrApplyBool(prior types.Bool, apiValue *bool) types.Bool {
	if prior.IsNull() || prior.IsUnknown() {
		return types.BoolNull()
	}
	return optionalBool(apiValue)
}

const httpStatusNotFound = 404
