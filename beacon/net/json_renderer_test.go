package net

import "fmt"
import "bytes"
import "testing"
import "net/http"
import "encoding/json"
import "net/http/httptest"
import "github.com/franela/goblin"

type jsonRendererScaffold struct {
	recorder *httptest.ResponseRecorder
	renderer *JSONRenderer
}

func (s *jsonRendererScaffold) parsedBody() jsonResponse {
	res := jsonResponse{}
	json.Unmarshal(s.recorder.Body.Bytes(), &res)
	return res
}

func (s *jsonRendererScaffold) Reset() {
	s.recorder = &httptest.ResponseRecorder{
		Body: bytes.NewBuffer([]byte{}),
	}
	s.renderer = &JSONRenderer{version: "testing"}
}

func Test_JSONRenderer(t *testing.T) {
	g := goblin.Goblin(t)

	s := &jsonRendererScaffold{}

	g.Describe("JSONRenderer", func() {

		g.BeforeEach(s.Reset)

		g.It("returns nil if no errors were encountered", func() {
			g.Assert(s.renderer.Render(s.recorder, HandlerResult{})).Equal(nil)
		})

		g.It("successfully sets the content type header", func() {
			s.renderer.Render(s.recorder, HandlerResult{})
			g.Assert(s.recorder.HeaderMap.Get("Content-Type")).Equal("application/json")
		})

		g.It("successfully set the response status code", func() {
			s.renderer.Render(s.recorder, HandlerResult{})
			g.Assert(s.recorder.Result().StatusCode).Equal(http.StatusOK)
		})

		g.Describe("having been given a result with some meta", func() {
			g.BeforeEach(func() {
				result := HandlerResult{
					Metadata: Metadata{
						"hello": "world",
					},
				}
				s.renderer.Render(s.recorder, result)
			})

			g.It("sends the metadata along too", func() {
				v, ok := s.parsedBody().Meta["hello"]
				g.Assert(ok).Equal(true)
				g.Assert(v).Equal("world")
			})
		})

		g.Describe("having been given a result with errors", func() {

			g.BeforeEach(func() {
				result := HandlerResult{
					Errors: []error{fmt.Errorf("bad-mojo")},
				}
				s.renderer.Render(s.recorder, result)
			})

			g.It("successfully set the response status code", func() {
				g.Assert(s.recorder.Result().StatusCode).Equal(http.StatusBadRequest)
			})

			g.It("successfully sets the json status to ERRORED if any errors present", func() {
				g.Assert(s.parsedBody().Status).Equal("ERRORED")
			})

		})

	})
}
