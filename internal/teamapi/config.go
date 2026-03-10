package teamapi

import "encoding/json"

type AgentConfig struct {
	Name                      *string `json:"name,omitempty"`
	Role                      *string `json:"role,omitempty"`
	Model                     *string `json:"model,omitempty"`
	SystemPrompt              *string `json:"systemPrompt,omitempty"`
	DebounceMs                *int64  `json:"debounceMs,omitempty"`
	WhenBusy                  *string `json:"whenBusy,omitempty"`
	ProcessBuffer             *string `json:"processBuffer,omitempty"`
	SendFinalResponseToThread *bool   `json:"sendFinalResponseToThread,omitempty"`
	RestrictOutput            *bool   `json:"restrictOutput,omitempty"`
	RestrictionMessage        *string `json:"restrictionMessage,omitempty"`
	RestrictionMaxInjections  *int64  `json:"restrictionMaxInjections,omitempty"`
	SummarizationKeepTokens   *int64  `json:"summarizationKeepTokens,omitempty"`
	SummarizationMaxTokens    *int64  `json:"summarizationMaxTokens,omitempty"`
}

type EnvValueRef struct {
	Kind  string  `json:"kind"`
	Mount *string `json:"mount,omitempty"`
	Path  *string `json:"path,omitempty"`
	Key   *string `json:"key,omitempty"`
}

type EnvVar struct {
	Name     string       `json:"name"`
	Value    *string      `json:"value,omitempty"`
	ValueRef *EnvValueRef `json:"valueRef,omitempty"`
}

type RestartPolicy struct {
	MaxAttempts *int64 `json:"maxAttempts,omitempty"`
	BackoffMs   *int64 `json:"backoffMs,omitempty"`
}

type MCPServerConfig struct {
	Namespace           *string        `json:"namespace,omitempty"`
	Command             *string        `json:"command,omitempty"`
	Workdir             *string        `json:"workdir,omitempty"`
	RequestTimeoutMs    *int64         `json:"requestTimeoutMs,omitempty"`
	StartupTimeoutMs    *int64         `json:"startupTimeoutMs,omitempty"`
	HeartbeatIntervalMs *int64         `json:"heartbeatIntervalMs,omitempty"`
	StaleTimeoutMs      *int64         `json:"staleTimeoutMs,omitempty"`
	Restart             *RestartPolicy `json:"restart,omitempty"`
	Env                 []EnvVar       `json:"env,omitempty"`
}

type WorkspaceVolumes struct {
	Enabled   *bool   `json:"enabled,omitempty"`
	MountPath *string `json:"mountPath,omitempty"`
}

type WorkspaceConfigurationConfig struct {
	Image         *string           `json:"image,omitempty"`
	Env           []EnvVar          `json:"env,omitempty"`
	InitialScript *string           `json:"initialScript,omitempty"`
	CpuLimit      *json.RawMessage  `json:"cpu_limit,omitempty"`
	MemoryLimit   *json.RawMessage  `json:"memory_limit,omitempty"`
	Platform      *string           `json:"platform,omitempty"`
	EnableDinD    *bool             `json:"enableDinD,omitempty"`
	TtlSeconds    *int64            `json:"ttlSeconds,omitempty"`
	Volumes       *WorkspaceVolumes `json:"volumes,omitempty"`
	Nix           *json.RawMessage  `json:"nix,omitempty"`
}

type MemoryBucketConfig struct {
	Scope            *string `json:"scope,omitempty"`
	CollectionPrefix *string `json:"collectionPrefix,omitempty"`
}
