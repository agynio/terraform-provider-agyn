package agentapi

type ComputeResources struct {
	RequestsCPU    *string `json:"requestsCpu,omitempty"`
	RequestsMemory *string `json:"requestsMemory,omitempty"`
	LimitsCPU      *string `json:"limitsCpu,omitempty"`
	LimitsMemory   *string `json:"limitsMemory,omitempty"`
}
