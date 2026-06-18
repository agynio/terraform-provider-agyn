package agentapi

import (
	"errors"

	"connectrpc.com/connect"
)

var ErrNotFound = errors.New("not found")

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) || connect.CodeOf(err) == connect.CodeNotFound
}

func IsConflict(err error) bool {
	code := connect.CodeOf(err)
	return code == connect.CodeAlreadyExists || code == connect.CodeAborted
}
