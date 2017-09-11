package net

import "io"
import "fmt"
import "log"
import "bytes"
import "net/url"
import "testing"
import "net/http"
import "net/http/httptest"
import "github.com/franela/goblin"

import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/logging"

type testPublisher struct {
}

func (u *testPublisher) PublishReader(string, io.Reader) error {
	return nil
}

type testUpgrader struct {
	errors    []error
	streamers []defs.Streamer
}

func (u *testUpgrader) UpgradeWebsocket(http.ResponseWriter, *http.Request, http.Header) (defs.Streamer, error) {
	if len(u.streamers) >= 1 {
		return u.streamers[0], nil
	}

	e := fmt.Errorf("bad")

	if len(u.errors) >= 1 {
		e = u.errors[0]
	}

	return nil, e
}

func newTestLogger() *logging.Logger {
	out := log.New(bytes.NewBuffer([]byte{}), "foobar", 0)
	return &logging.Logger{Logger: out}
}

type requestRuntimeScaffold struct {
	body      *bytes.Buffer
	values    url.Values
	upgrader  *testUpgrader
	publisher *testPublisher
	request   *http.Request
	runtime   *RequestRuntime
}

func (s *requestRuntimeScaffold) Reset() {
	s.body = bytes.NewBuffer([]byte{})
	s.values = make(url.Values)
	s.upgrader = &testUpgrader{}
	s.publisher = &testPublisher{}
	s.request = httptest.NewRequest("GET", "/something", s.body)
	s.runtime = &RequestRuntime{
		Values:            s.values,
		WebsocketUpgrader: s.upgrader,
		ChannelPublisher:  s.publisher,
		Logger:            newTestLogger(),
		Request:           s.request,
	}
}

func Test_RequestRuntime(t *testing.T) {
	g := goblin.Goblin(t)

	s := &requestRuntimeScaffold{}

	g.Describe("RequestRuntime", func() {

		g.BeforeEach(s.Reset)

		g.Describe("#Websocket", func() {

			g.It("returns the error rerturned from the upgrader", func() {
				s.upgrader.errors = append(s.upgrader.errors, fmt.Errorf("bad socket"))
				_, e := s.runtime.Websocket()
				g.Assert(e.Error()).Equal("bad socket")
			})
		})

		g.Describe("#ServerError", func() {

			g.It("returns the error string in the appropriate error response", func() {
				r := s.runtime.ServerError()
				g.Assert(r.Errors[0].Error()).Equal(defs.ErrServerError)
			})
		})

		g.Describe("#LogicError", func() {

			g.It("returns the error string in the appropriate error response", func() {
				r := s.runtime.LogicError("bad request")
				g.Assert(r.Errors[0].Error()).Equal("bad request")
			})
		})

		g.Describe("#ReadBody", func() {
			g.It("returns an error if unable to parse the request body into the given interface", func() {
				s.body.Write([]byte("}{"))

				dest := struct {
					Name string `json:"name"`
				}{}

				g.Assert(s.runtime.ReadBody(&dest) != nil).Equal(true)
			})
			g.It("returns the content type from the request header", func() {
				json := []byte(`{"name":"frank reynolds"}`)
				s.body.Write(json)
				s.body.Grow(len(json))
				dest := &struct {
					Name string `json:"name"`
				}{}

				g.Assert(s.runtime.ReadBody(dest)).Equal(nil)
				g.Assert(dest.Name).Equal("frank reynolds")
			})
		})

		g.Describe("#ContentType", func() {
			g.It("returns the content type from the request header", func() {
				s.request.Header.Set(defs.APIContentTypeHeader, "something")
				g.Assert(s.runtime.ContentType()).Equal("something")
			})
		})

		g.Describe("#GetQueryParam", func() {

			g.It("returns an empty string if unable to parse the query string", func() {
				g.Assert(s.runtime.GetQueryParam("something")).Equal("")
			})

			g.It("returns the string set on the http request", func() {
				s.request.URL, _ = url.Parse("http://example.com?something=hi")
				g.Assert(s.runtime.GetQueryParam("something")).Equal("hi")
			})

		})

		g.Describe("#HeaderValue", func() {

			g.It("returns an empty string if header is not set", func() {
				g.Assert(s.runtime.HeaderValue("something")).Equal("")
			})

			g.It("returns the string associated with the header", func() {
				s.request.Header.Set("something", "hello")
				g.Assert(s.runtime.HeaderValue("something")).Equal("hello")
			})

		})

	})
}
