package net

import "fmt"
import "bytes"
import "regexp"
import "testing"
import "net/http/httptest"

func Test_Net_RouteList(suite *testing.T) {
	first := func(*RequestRuntime) HandlerResult {
		return HandlerResult{Errors: []error{fmt.Errorf("first")}}
	}

	second := func(*RequestRuntime) HandlerResult {
		return HandlerResult{Errors: []error{fmt.Errorf("second")}}
	}

	runtime := &RequestRuntime{}

	r := RouteList{
		RouteConfig{"GET", regexp.MustCompile("^/first$")}:                               first,
		RouteConfig{"GET", regexp.MustCompile("^/second$")}:                              second,
		RouteConfig{"GET", regexp.MustCompile("^/obj/(?P<id>\\d+)$")}:                    second,
		RouteConfig{"GET", regexp.MustCompile("^/unnamed/(\\d+)$")}:                      second,
		RouteConfig{"GET", regexp.MustCompile("^/multiple/(?P<id>\\d+)/(?P<two>\\d+)$")}: second,
	}

	suite.Run("returns false if request matches no routes", func(test *testing.T) {
		req := httptest.NewRequest("POST", "/foo", bytes.NewBuffer([]byte("whoa")))
		ok, _, _ := r.match(req)

		if ok == true {
			test.Fatalf("expected no match but was told one was found")
		}
	})

	suite.Run("returns true & first handler if request matches first routes", func(test *testing.T) {
		req := httptest.NewRequest("GET", "/first", bytes.NewBuffer([]byte("whoa")))
		_, _, handler := r.match(req)
		result := handler(runtime)
		if len(result.Errors) != 1 || result.Errors[0].Error() != "first" {
			test.Fatalf("expected first handler")
		}
	})

	suite.Run("returns true & first handler if request matches first routes", func(test *testing.T) {
		req := httptest.NewRequest("GET", "/second", bytes.NewBuffer([]byte("whoa")))
		_, _, handler := r.match(req)
		result := handler(runtime)
		if len(result.Errors) != 1 || result.Errors[0].Error() != "second" {
			test.Fatalf("expected second handler")
		}
	})

	suite.Run("returns the parameter based on matches in regex", func(test *testing.T) {
		req := httptest.NewRequest("GET", "/obj/123", bytes.NewBuffer([]byte("whoa")))
		_, params, _ := r.match(req)
		if params.Get("id") != "123" {
			test.Fatalf("expected to have id param \"123\" but did not; %v", params)
			return
		}
	})

	suite.Run("provides access to unnamed params via $0", func(test *testing.T) {
		req := httptest.NewRequest("GET", "/unnamed/123", bytes.NewBuffer([]byte("whoa")))
		_, params, _ := r.match(req)
		if params.Get("$0") != "123" {
			test.Fatalf("expected to have id param \"123\" but did not; %v", params)
			return
		}
	})

	suite.Run("provides access to multiple", func(test *testing.T) {
		req := httptest.NewRequest("GET", "/multiple/123/456", bytes.NewBuffer([]byte("whoa")))
		_, params, _ := r.match(req)
		if params.Get("id") != "123" || params.Get("two") != "456" {
			test.Fatalf("expected to have id param \"123\" but did not; %v", params)
			return
		}
	})
}
