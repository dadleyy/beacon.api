package net

import "fmt"
import "net/url"
import "net/http"

// RouteList is simply a map between a route configuration and the handler function
type RouteList map[RouteConfig]Handler

func noop(r *RequestRuntime) HandlerResult {
	return HandlerResult{Errors: []error{fmt.Errorf("not-found")}}
}

func (list *RouteList) match(request *http.Request) (bool, url.Values, Handler) {
	method, path := request.Method, request.URL.EscapedPath()
	pbytes := []byte(path)

	for config, handler := range *list {
		if match := config.Pattern.Match(pbytes); config.Method != method || match != true {
			continue
		}

		if s := config.Pattern.NumSubexp(); s == 0 {
			return true, make(url.Values), handler
		}

		groups := config.Pattern.FindAllStringSubmatch(string(path), -1)
		names := config.Pattern.SubexpNames()

		if groups == nil || len(groups) != 1 {
			return true, make(url.Values), handler
		}

		values := groups[0][1:]
		params := make(url.Values)
		count := len(names)

		if count >= 0 {
			names = names[1:]
			count = len(names)
		}

		for indx, v := range values {
			if indx < count && len(names[indx]) >= 1 {
				params.Set(names[indx], v)
				continue
			}

			params.Set(fmt.Sprintf("$%d", indx), v)
		}

		return true, params, handler
	}

	return false, make(url.Values), noop
}
