package routes

import "bytes"
import "testing"
import "net/http/httptest"
import "github.com/franela/goblin"
import "github.com/dadleyy/beacon.api/beacon/net"

type systemInfoScaffold struct {
	body    *bytes.Buffer
	runtime *net.RequestRuntime
}

func (s *systemInfoScaffold) Reset() {
	s.body = bytes.NewBuffer([]byte{})
	s.runtime = &net.RequestRuntime{
		Request: httptest.NewRequest("GET", "/system", s.body),
	}
}

func Test_SystemInfoRoute(t *testing.T) {
	g := goblin.Goblin(t)

	scaffold := &systemInfoScaffold{}

	g.Describe("SystemInfo", func() {

		g.BeforeEach(scaffold.Reset)

		g.It("sets the current time", func() {
			r := SystemInfo(scaffold.runtime)
			_, ok := r.Metadata["time"]
			g.Assert(ok).Equal(true)
		})
	})
}
