package security

import "bytes"
import "testing"
import "net/http/httptest"

func Test_AnyOrigin(suite *testing.T) {
	req := httptest.NewRequest("GET", "/anything", bytes.NewBuffer([]byte{}))

	if v := AnyOrigin(req); v != true {
		suite.Fatalf("expected true but got false")
	}
}
