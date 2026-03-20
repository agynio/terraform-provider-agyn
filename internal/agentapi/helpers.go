package agentapi

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func parseUUID(id string) (openapi_types.UUID, error) {
	value, err := uuid.Parse(id)
	if err != nil {
		return openapi_types.UUID{}, fmt.Errorf("invalid UUID: %w", err)
	}
	return openapi_types.UUID(value), nil
}

func uuidToString(id openapi_types.UUID) string {
	return uuid.UUID(id).String()
}

func decodePayload(source any, target any) error {
	switch value := source.(type) {
	case []byte:
		return json.Unmarshal(value, target)
	case json.RawMessage:
		return json.Unmarshal(value, target)
	default:
		raw, err := json.Marshal(source)
		if err != nil {
			return err
		}
		return json.Unmarshal(raw, target)
	}
}
