package resources

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/agynio/terraform-provider-agyn/internal/teamapi"
)

type envValueRefModel struct {
	Kind  types.String `tfsdk:"kind"`
	Mount types.String `tfsdk:"mount"`
	Path  types.String `tfsdk:"path"`
	Key   types.String `tfsdk:"key"`
}

type envVarModel struct {
	Name     types.String      `tfsdk:"name"`
	Value    types.String      `tfsdk:"value"`
	ValueRef *envValueRefModel `tfsdk:"value_ref"`
}

func envVarsFromModels(models []envVarModel) ([]teamapi.EnvVar, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(models) == 0 {
		return nil, diags
	}

	vars := make([]teamapi.EnvVar, 0, len(models))
	for index, model := range models {
		if model.Name.IsNull() || model.Name.IsUnknown() {
			diags.AddError("Missing Environment Variable Name", fmt.Sprintf("Env entry %d is missing a name.", index))
			continue
		}

		entry := teamapi.EnvVar{Name: model.Name.ValueString()}
		if !model.Value.IsNull() && !model.Value.IsUnknown() {
			value := model.Value.ValueString()
			entry.Value = &value
		}
		if model.ValueRef != nil {
			ref, refDiags := envValueRefFromModel(*model.ValueRef)
			diags.Append(refDiags...)
			if refDiags.HasError() {
				continue
			}
			entry.ValueRef = ref
		}

		if entry.Value == nil && entry.ValueRef == nil {
			diags.AddError("Missing Environment Variable Value", fmt.Sprintf("Env entry %d must set value or value_ref.", index))
			continue
		}
		if entry.Value != nil && entry.ValueRef != nil {
			diags.AddError("Conflicting Environment Variable Values", fmt.Sprintf("Env entry %d cannot set both value and value_ref.", index))
			continue
		}

		vars = append(vars, entry)
	}

	return vars, diags
}

func envValueRefFromModel(model envValueRefModel) (*teamapi.EnvValueRef, diag.Diagnostics) {
	var diags diag.Diagnostics
	if model.Kind.IsNull() || model.Kind.IsUnknown() {
		diags.AddError("Missing Environment Value Reference Kind", "Env value_ref.kind must be set.")
		return nil, diags
	}
	if model.Mount.IsNull() || model.Mount.IsUnknown() {
		diags.AddError("Missing Environment Value Reference Mount", "Env value_ref.mount must be set.")
		return nil, diags
	}
	if model.Path.IsNull() || model.Path.IsUnknown() {
		diags.AddError("Missing Environment Value Reference Path", "Env value_ref.path must be set.")
		return nil, diags
	}
	if model.Key.IsNull() || model.Key.IsUnknown() {
		diags.AddError("Missing Environment Value Reference Key", "Env value_ref.key must be set.")
		return nil, diags
	}

	return &teamapi.EnvValueRef{
		Kind:  model.Kind.ValueString(),
		Mount: stringPointer(model.Mount),
		Path:  stringPointer(model.Path),
		Key:   stringPointer(model.Key),
	}, diags
}

func envVarModelsFromAPI(vars []teamapi.EnvVar) []envVarModel {
	if len(vars) == 0 {
		return nil
	}

	models := make([]envVarModel, 0, len(vars))
	for _, env := range vars {
		model := envVarModel{
			Name:  types.StringValue(env.Name),
			Value: optionalString(env.Value),
		}
		if env.ValueRef != nil {
			model.ValueRef = &envValueRefModel{
				Kind:  types.StringValue(env.ValueRef.Kind),
				Mount: optionalString(env.ValueRef.Mount),
				Path:  optionalString(env.ValueRef.Path),
				Key:   optionalString(env.ValueRef.Key),
			}
		}
		models = append(models, model)
	}

	return models
}
