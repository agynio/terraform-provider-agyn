package resources

import (
	agentsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/agents/v1"
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

func computeResourcesFromModel(model *computeResourcesModel) *agentsv1.ComputeResources {
	if model == nil {
		return nil
	}
	resources := &agentsv1.ComputeResources{
		RequestsCpu:    stringValue(model.RequestsCPU),
		RequestsMemory: stringValue(model.RequestsMemory),
		LimitsCpu:      stringValue(model.LimitsCPU),
		LimitsMemory:   stringValue(model.LimitsMemory),
	}
	if resources.RequestsCpu == "" && resources.RequestsMemory == "" && resources.LimitsCpu == "" && resources.LimitsMemory == "" {
		return nil
	}
	return resources
}

func updateComputeResources(plan *computeResourcesModel, prior *computeResourcesModel) *agentsv1.ComputeResources {
	if plan == nil {
		if prior == nil {
			return nil
		}
		return &agentsv1.ComputeResources{}
	}

	priorModel := computeResourcesModel{
		RequestsCPU:    types.StringNull(),
		RequestsMemory: types.StringNull(),
		LimitsCPU:      types.StringNull(),
		LimitsMemory:   types.StringNull(),
	}
	if prior != nil {
		priorModel = *prior
	}

	resources := &agentsv1.ComputeResources{
		RequestsCpu:    updateStringValue(plan.RequestsCPU, priorModel.RequestsCPU),
		RequestsMemory: updateStringValue(plan.RequestsMemory, priorModel.RequestsMemory),
		LimitsCpu:      updateStringValue(plan.LimitsCPU, priorModel.LimitsCPU),
		LimitsMemory:   updateStringValue(plan.LimitsMemory, priorModel.LimitsMemory),
	}
	if resources.RequestsCpu == "" && resources.RequestsMemory == "" && resources.LimitsCpu == "" && resources.LimitsMemory == "" {
		if prior != nil {
			return &agentsv1.ComputeResources{}
		}
		return nil
	}
	return resources
}

func computeResourcesToModel(resources *agentsv1.ComputeResources) *computeResourcesModel {
	if resources == nil {
		return nil
	}
	return &computeResourcesModel{
		RequestsCPU:    optionalString(resources.RequestsCpu),
		RequestsMemory: optionalString(resources.RequestsMemory),
		LimitsCPU:      optionalString(resources.LimitsCpu),
		LimitsMemory:   optionalString(resources.LimitsMemory),
	}
}

func updateStringValue(plan types.String, prior types.String) string {
	if plan.IsUnknown() {
		if prior.IsNull() || prior.IsUnknown() {
			return ""
		}
		return prior.ValueString()
	}
	if plan.IsNull() {
		return ""
	}
	return plan.ValueString()
}
