package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stringListFromPlan(ctx context.Context, list types.List) ([]string, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	values := []string{}
	diags := list.ElementsAs(ctx, &values, false)
	return values, diags
}

func stringListToState(ctx context.Context, values []string, prior types.List) (types.List, diag.Diagnostics) {
	if len(values) == 0 {
		if prior.IsNull() || prior.IsUnknown() {
			return types.ListNull(types.StringType), nil
		}
		values = []string{}
	}
	return types.ListValueFrom(ctx, types.StringType, values)
}
