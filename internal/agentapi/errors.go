package agentapi

import "connectrpc.com/connect"

func IsNotFound(err error) bool {
	return connect.CodeOf(err) == connect.CodeNotFound
}

func IsConflict(err error) bool {
	code := connect.CodeOf(err)
	return code == connect.CodeAlreadyExists || code == connect.CodeAborted
}
