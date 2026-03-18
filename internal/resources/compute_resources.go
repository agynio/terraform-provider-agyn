package resources

import (
	"github.com/agynio/terraform-provider-agyn/internal/teamapi"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type computeResourcesModel struct {
	RequestsCPU    types.String `tfsdk:"requests_cpu"`
	RequestsMemory types.String `tfsdk:"requests_memory"`
	LimitsCPU      types.String `tfsdk:"limits_cpu"`
	LimitsMemory   types.String `tfsdk:"limits_memory"`
}

func computeResourcesSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"requests_cpu": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "CPU requests (for example, 500m).",
		},
		"requests_memory": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Memory requests (for example, 1Gi).",
		},
		"limits_cpu": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "CPU limits (for example, 1).",
		},
		"limits_memory": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Memory limits (for example, 2Gi).",
		},
	}
}

func computeResourcesFromModel(model *computeResourcesModel) *teamapi.ComputeResources {
	if model == nil {
		return nil
	}
	resources := teamapi.ComputeResources{
		RequestsCPU:    stringPointer(model.RequestsCPU),
		RequestsMemory: stringPointer(model.RequestsMemory),
		LimitsCPU:      stringPointer(model.LimitsCPU),
		LimitsMemory:   stringPointer(model.LimitsMemory),
	}
	if resources.RequestsCPU == nil && resources.RequestsMemory == nil && resources.LimitsCPU == nil && resources.LimitsMemory == nil {
		return nil
	}
	return &resources
}

func computeResourcesToModel(resources *teamapi.ComputeResources) *computeResourcesModel {
	if resources == nil {
		return nil
	}
	return &computeResourcesModel{
		RequestsCPU:    optionalString(resources.RequestsCPU),
		RequestsMemory: optionalString(resources.RequestsMemory),
		LimitsCPU:      optionalString(resources.LimitsCPU),
		LimitsMemory:   optionalString(resources.LimitsMemory),
	}
}
