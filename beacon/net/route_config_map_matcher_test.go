package net

import "fmt"
import "bytes"
import "regexp"
import "testing"
import "net/http/httptest"
import "github.com/franela/goblin"

func Test_RouteConfigMapMatcher(t *testing.T) {
	g := goblin.Goblin(t)

	first := func(*RequestRuntime) HandlerResult {
		return HandlerResult{Errors: []error{fmt.Errorf("first")}}
	}

	second := func(*RequestRuntime) HandlerResult {
		return HandlerResult{Errors: []error{fmt.Errorf("second")}}
	}

	runtime := &RequestRuntime{}

	r := RouteConfigMapMatcher{
		RouteConfig{"GET", regexp.MustCompile("^/first$")}:                               first,
		RouteConfig{"GET", regexp.MustCompile("^/second$")}:                              second,
		RouteConfig{"GET", regexp.MustCompile("^/obj/(?P<id>\\d+)$")}:                    second,
		RouteConfig{"GET", regexp.MustCompile("^/unnamed/(\\d+)$")}:                      second,
		RouteConfig{"GET", regexp.MustCompile("^/multiple/(?P<id>\\d+)/(?P<two>\\d+)$")}: second,
	}

	g.Describe("RouteConfigMapMatcher", func() {
		g.It("returns false if request matches no routes", func() {
			req := httptest.NewRequest("POST", "/foo", bytes.NewBuffer([]byte("whoa")))
			ok, _, _ := r.MatchRequest(req)
			g.Assert(ok).Equal(false)
		})

		g.It("returns true & first handler if request matches first routes", func() {
			req := httptest.NewRequest("GET", "/first", bytes.NewBuffer([]byte("whoa")))
			_, _, handler := r.MatchRequest(req)
			result := handler(runtime)
			g.Assert(result.Errors[0].Error()).Equal("first")
		})

		g.It("returns true & first handler if request matches first routes", func() {
			req := httptest.NewRequest("GET", "/second", bytes.NewBuffer([]byte("whoa")))
			_, _, handler := r.MatchRequest(req)
			result := handler(runtime)
			g.Assert(result.Errors[0].Error()).Equal("second")
		})

		g.It("returns the parameter based on matches in regex", func() {
			req := httptest.NewRequest("GET", "/obj/123", bytes.NewBuffer([]byte("whoa")))
			_, params, _ := r.MatchRequest(req)
			g.Assert(params.Get("id")).Equal("123")
		})

		g.It("provides access to unnamed params via $0", func() {
			req := httptest.NewRequest("GET", "/unnamed/123", bytes.NewBuffer([]byte("whoa")))
			_, params, _ := r.MatchRequest(req)
			g.Assert(params.Get("$0")).Equal("123")
		})

		g.It("provides access to multiple", func() {
			req := httptest.NewRequest("GET", "/multiple/123/456", bytes.NewBuffer([]byte("whoa")))
			_, params, _ := r.MatchRequest(req)
			g.Assert(params.Get("id")).Equal("123")
			g.Assert(params.Get("two")).Equal("456")
		})

	})
}
