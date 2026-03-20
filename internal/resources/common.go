package resources

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/proto"
)

func stringValue(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}

func stringPointer(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	value := v.ValueString()
	return proto.String(value)
}

func updateStringPointer(plan types.String, prior types.String) *string {
	if plan.IsUnknown() {
		return nil
	}
	if plan.IsNull() {
		if prior.IsNull() || prior.IsUnknown() {
			return nil
		}
		return proto.String("")
	}
	value := plan.ValueString()
	return proto.String(value)
}

func boolPointer(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	value := v.ValueBool()
	return proto.Bool(value)
}

func optionalString(v string) types.String {
	if v == "" {
		return types.StringNull()
	}
	return types.StringValue(v)
}

func preserveSensitiveString(fallback types.String, apiValue string) types.String {
	if apiValue == "" {
		return fallback
	}
	return types.StringValue(apiValue)
}
