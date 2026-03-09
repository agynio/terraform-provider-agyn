package teamclient

import "encoding/json"

func marshalWorkspaceLimit(raw json.RawMessage) ([]byte, error) {
	if len(raw) == 0 {
		return []byte("null"), nil
	}
	return raw, nil
}

func unmarshalWorkspaceLimit(raw *json.RawMessage, data []byte) error {
	*raw = append((*raw)[:0], data...)
	return nil
}

func (t *PostWorkspaceConfigurationsJSONBody_Config_CpuLimit) UnmarshalJSON(b []byte) error {
	return unmarshalWorkspaceLimit(&t.union, b)
}

func (t PostWorkspaceConfigurationsJSONBody_Config_CpuLimit) MarshalJSON() ([]byte, error) {
	return marshalWorkspaceLimit(t.union)
}

func (t *PostWorkspaceConfigurationsJSONBody_Config_MemoryLimit) UnmarshalJSON(b []byte) error {
	return unmarshalWorkspaceLimit(&t.union, b)
}

func (t PostWorkspaceConfigurationsJSONBody_Config_MemoryLimit) MarshalJSON() ([]byte, error) {
	return marshalWorkspaceLimit(t.union)
}

func (t *PatchWorkspaceConfigurationsIdJSONBody_Config_CpuLimit) UnmarshalJSON(b []byte) error {
	return unmarshalWorkspaceLimit(&t.union, b)
}

func (t PatchWorkspaceConfigurationsIdJSONBody_Config_CpuLimit) MarshalJSON() ([]byte, error) {
	return marshalWorkspaceLimit(t.union)
}

func (t *PatchWorkspaceConfigurationsIdJSONBody_Config_MemoryLimit) UnmarshalJSON(b []byte) error {
	return unmarshalWorkspaceLimit(&t.union, b)
}

func (t PatchWorkspaceConfigurationsIdJSONBody_Config_MemoryLimit) MarshalJSON() ([]byte, error) {
	return marshalWorkspaceLimit(t.union)
}

func (t *GetWorkspaceConfigurations_200_Items_Config_CpuLimit) UnmarshalJSON(b []byte) error {
	return unmarshalWorkspaceLimit(&t.union, b)
}

func (t GetWorkspaceConfigurations_200_Items_Config_CpuLimit) MarshalJSON() ([]byte, error) {
	return marshalWorkspaceLimit(t.union)
}

func (t *GetWorkspaceConfigurations_200_Items_Config_MemoryLimit) UnmarshalJSON(b []byte) error {
	return unmarshalWorkspaceLimit(&t.union, b)
}

func (t GetWorkspaceConfigurations_200_Items_Config_MemoryLimit) MarshalJSON() ([]byte, error) {
	return marshalWorkspaceLimit(t.union)
}

func (t *PostWorkspaceConfigurations_201_Config_CpuLimit) UnmarshalJSON(b []byte) error {
	return unmarshalWorkspaceLimit(&t.union, b)
}

func (t PostWorkspaceConfigurations_201_Config_CpuLimit) MarshalJSON() ([]byte, error) {
	return marshalWorkspaceLimit(t.union)
}

func (t *PostWorkspaceConfigurations_201_Config_MemoryLimit) UnmarshalJSON(b []byte) error {
	return unmarshalWorkspaceLimit(&t.union, b)
}

func (t PostWorkspaceConfigurations_201_Config_MemoryLimit) MarshalJSON() ([]byte, error) {
	return marshalWorkspaceLimit(t.union)
}

func (t *GetWorkspaceConfigurationsId_200_Config_CpuLimit) UnmarshalJSON(b []byte) error {
	return unmarshalWorkspaceLimit(&t.union, b)
}

func (t GetWorkspaceConfigurationsId_200_Config_CpuLimit) MarshalJSON() ([]byte, error) {
	return marshalWorkspaceLimit(t.union)
}

func (t *GetWorkspaceConfigurationsId_200_Config_MemoryLimit) UnmarshalJSON(b []byte) error {
	return unmarshalWorkspaceLimit(&t.union, b)
}

func (t GetWorkspaceConfigurationsId_200_Config_MemoryLimit) MarshalJSON() ([]byte, error) {
	return marshalWorkspaceLimit(t.union)
}

func (t *PatchWorkspaceConfigurationsId_200_Config_CpuLimit) UnmarshalJSON(b []byte) error {
	return unmarshalWorkspaceLimit(&t.union, b)
}

func (t PatchWorkspaceConfigurationsId_200_Config_CpuLimit) MarshalJSON() ([]byte, error) {
	return marshalWorkspaceLimit(t.union)
}

func (t *PatchWorkspaceConfigurationsId_200_Config_MemoryLimit) UnmarshalJSON(b []byte) error {
	return unmarshalWorkspaceLimit(&t.union, b)
}

func (t PatchWorkspaceConfigurationsId_200_Config_MemoryLimit) MarshalJSON() ([]byte, error) {
	return marshalWorkspaceLimit(t.union)
}
