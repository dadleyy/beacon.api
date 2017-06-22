package security

import "net/http"

// AnyOrigin is a no-op to allow any origin to upgrade their request to a websocket
func AnyOrigin(r *http.Request) bool {
	return true
}
