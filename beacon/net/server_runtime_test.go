package net

import "bytes"
import "net/url"
import "testing"
import "net/http"
import "encoding/json"
import "net/http/httptest"
import "github.com/franela/goblin"
import "github.com/dadleyy/beacon.api/beacon/defs"

type testRouteMatcher struct {
	matches []Handler
}

func (m *testRouteMatcher) MatchRequest(*http.Request) (bool, url.Values, Handler) {
	values := make(url.Values)

	if len(m.matches) >= 1 {
		match := m.matches[0]
		m.matches = m.matches[1:]
		return true, values, match
	}

	return false, values, nil
}

type serverRuntimeScaffold struct {
	upgrader       *testUpgrader
	runtime        *ServerRuntime
	publisher      *testPublisher
	request        *http.Request
	body           *bytes.Buffer
	responseWriter *httptest.ResponseRecorder
	routes         *testRouteMatcher
}

func (s *serverRuntimeScaffold) Reset() {
	s.body = new(bytes.Buffer)

	s.request = httptest.NewRequest("GET", "/path", s.body)

	s.responseWriter = httptest.NewRecorder()

	s.publisher = &testPublisher{}

	s.upgrader = &testUpgrader{}

	s.routes = &testRouteMatcher{}

	s.runtime = &ServerRuntime{
		Multiplexer:       s.routes,
		WebsocketUpgrader: s.upgrader,
		ChannelPublisher:  s.publisher,
		Logger:            newTestLogger(),
	}
}

func Test_ServerRuntime(t *testing.T) {
	g := goblin.Goblin(t)

	s := &serverRuntimeScaffold{}

	g.Describe("ServerRuntime", func() {

		g.BeforeEach(s.Reset)

		g.Describe("#ServeHTTP", func() {

			g.It("returns a 404 status code having no routes in it's route list", func() {
				s.runtime.ServeHTTP(s.responseWriter, s.request)
				g.Assert(s.responseWriter.Result().StatusCode).Equal(404)
			})

			g.It("renders the not-found error message if no route was found", func() {
				s.runtime.ServeHTTP(s.responseWriter, s.request)
				de := json.NewDecoder(s.responseWriter.Body)
				jsonOut := struct {
					Errors []string `json:"errors"`
				}{}

				if e := de.Decode(&jsonOut); e != nil {
					g.Fail(e.Error())
					return
				}

				g.Assert(jsonOut.Errors[0]).Equal(defs.ErrNotFound)
			})

			g.Describe("with a matching handler in the multiplexer", func() {

				var result HandlerResult

				g.BeforeEach(func() {
					s.routes.matches = append(s.routes.matches, func(runtime *RequestRuntime) HandlerResult {
						return result
					})
				})

				g.It("redirects if the result returns a redirect", func() {
					result = HandlerResult{Redirect: "http://example.com"}
					s.runtime.ServeHTTP(s.responseWriter, s.request)
					g.Assert(s.responseWriter.Result().Header.Get("Location")).Equal("http://example.com")
				})

				g.It("does nothing if the result explicitly delares itself a render-less operation", func() {
					result = HandlerResult{NoRender: true}
					s.runtime.ServeHTTP(s.responseWriter, s.request)
					g.Assert(s.responseWriter.Body.Len()).Equal(0)
				})

			})

		})

	})
}
