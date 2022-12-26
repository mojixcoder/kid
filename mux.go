package kid

import (
	"bytes"
	"errors"
	"strings"
)

// Errors.
var (
	errNotFound         = errors.New("match not found")
	errMethodNotAllowed = errors.New("method is not allowed")
)

type (
	// Router is the struct which holds all of the routes.
	Router struct {
		routes []Route
	}

	// Route is a route with its contents.
	Route struct {
		segments    []Segment
		methods     []string
		handler     HandlerFunc
		middlewares []MiddlewareFunc
	}

	// Segment is the type of each path segment.
	Segment struct {
		isParam bool
		tpl     string
	}

	// Params is the type of path parameters.
	Params map[string]string
)

// newRouter returns a new router.
func newRouter() Router {
	return Router{routes: make([]Route, 0)}
}

// add adds a route to the router.
func (router *Router) add(path string, handler HandlerFunc, methods []string, middlewares []MiddlewareFunc) {
	if len(methods) == 0 {
		panic("providing at least one method is required")
	}

	if handler == nil {
		panic("handler cannot be nil")
	}

	path = cleanPath(path, false)

	segments := strings.Split(path, "/")[1:]

	routeSegments := make([]Segment, 0, len(segments))

	for _, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			routeSegments = append(routeSegments, Segment{isParam: true, tpl: segment[1 : len(segment)-1]})
		} else {
			routeSegments = append(routeSegments, Segment{isParam: false, tpl: segment})
		}
	}

	router.routes = append(router.routes, Route{segments: routeSegments, methods: methods, handler: handler, middlewares: middlewares})
}

// match determines if the given path and method matches the route.
func (route *Route) match(path, method string) (Params, error) {
	params := make(Params)
	var end bool

	for segmentIndex, segment := range route.segments {
		i := strings.IndexByte(path, '/')
		j := i + 1

		if i == -1 {
			i = len(path)
			j = i
			end = true

			// No slashes are left but there are still more segments.
			if segmentIndex != len(route.segments)-1 {
				return nil, errNotFound
			}
		}

		if segment.isParam {
			params[segment.tpl] = path[:i]

			// Empty parameter
			if len(path[:i]) == 0 {
				return nil, errNotFound
			}
		} else {
			if segment.tpl != path[:i] {
				return nil, errNotFound
			}
		}

		path = path[j:]
	}

	// Segments are ended but there are still more slashes.
	if !end {
		return nil, errNotFound
	}

	if !methodExists(method, route.methods) {
		return nil, errMethodNotAllowed
	}

	return params, nil
}

// find finds a route which matches the given path and method.
func (router *Router) find(path string, method string) (Route, Params, error) {
	path = cleanPath(path, true)[1:]

	var returnedErr error

	for _, route := range router.routes {
		params, err := route.match(path, method)
		if err == nil {
			return route, params, nil
		}

		if err == errMethodNotAllowed {
			returnedErr = err
		} else if returnedErr == nil {
			returnedErr = err
		}
	}

	return Route{}, nil, returnedErr

}

// cleanPath normalizes the path.
//
// If soft is false it also removes duplicate slashes.
func cleanPath(s string, soft bool) string {
	if s == "" {
		return "/"
	}

	if s[0] != '/' {
		s = "/" + s
	}

	if soft {
		return s
	}

	// Removing repeated slashes.
	var buff bytes.Buffer
	for i := 0; i < len(s); i++ {
		if i != 0 && s[i] == '/' && s[i-1] == '/' {
			continue
		}
		buff.WriteByte(s[i])
	}

	return buff.String()
}

// methodExists checks whether a method exists in a slice of methods.
func methodExists(method string, methods []string) bool {
	for _, v := range methods {
		if v == method {
			return true
		}
	}

	return false
}
