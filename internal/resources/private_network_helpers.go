package resources

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func int32ListFromPlan(list types.List) ([]int32, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	elements := list.Elements()
	values := make([]int32, 0, len(elements))
	for _, element := range elements {
		value, ok := element.(types.Int32)
		if !ok || value.IsNull() || value.IsUnknown() {
			return nil, fmt.Errorf("ports cannot contain empty values")
		}
		values = append(values, value.ValueInt32())
	}
	return values, nil
}

func int32ListState(values []int32) types.List {
	elements := make([]attr.Value, 0, len(values))
	for _, value := range values {
		elements = append(elements, types.Int32Value(value))
	}
	return types.ListValueMust(types.Int32Type, elements)
}

func timestampString(timestamp *timestamppb.Timestamp) types.String {
	if timestamp == nil {
		return types.StringNull()
	}
	return types.StringValue(timestamp.AsTime().UTC().Format(time.RFC3339))
}
