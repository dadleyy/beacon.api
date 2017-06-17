package security

import "net/http"

func AnyOrigin(r *http.Request) bool {
	return true
}
