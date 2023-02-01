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

// Path parameters prefix and suffix.
const (
	paramPrefix     = "{"
	paramSuffix     = "}"
	plusParamPrefix = paramPrefix + "+"
	starParamPrefix = paramPrefix + "*"
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
		isPlus  bool
		isStar  bool
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

	for i, segment := range segments {
		if (strings.HasPrefix(segment, plusParamPrefix) || strings.HasPrefix(segment, starParamPrefix)) && strings.HasSuffix(segment, paramSuffix) {
			isPlus := router.isPlus(segment)
			isStar := !isPlus

			if i == len(segments)-1 {
				routeSegments = append(
					routeSegments,
					Segment{isParam: true, isPlus: isPlus, isStar: isStar, tpl: segment[2 : len(segment)-1]},
				)
			} else if i == len(segments)-2 {
				if segments[i+1] != "" {
					panic("plus/star path parameters can only be the last part of a path")
				}
				routeSegments = append(
					routeSegments,
					Segment{isParam: true, isPlus: isPlus, isStar: isStar, tpl: segment[2 : len(segment)-1]},
				)
				break
			} else {
				panic("plus/star path parameters can only be the last part of a path")
			}
		} else if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
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
	totalSegments := len(route.segments)
	var end bool

	for segmentIndex, segment := range route.segments {
		i := strings.IndexByte(path, '/')
		j := i + 1

		if i == -1 {
			i = len(path)
			j = i
			end = true

			// No slashes are left but there are still more segments.
			if segmentIndex != totalSegments-1 {
				// It means /api/v1 will be matched to /api/v1/{*param}
				lastSegment := route.segments[totalSegments-1]
				if segmentIndex == totalSegments-2 && lastSegment.isStar {
					end = true
					params[lastSegment.tpl] = ""
				} else {
					return nil, errNotFound
				}
			}
		}

		if segment.isParam {
			if segment.isPlus || segment.isStar {
				if len(path) == 0 && segment.isPlus {
					return nil, errNotFound
				}

				end = true
				params[segment.tpl] = path

				// Break because it's always the last part of the path.
				break
			}

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

	// We have no routes, so anything won't be found.
	if len(router.routes) == 0 {
		return Route{}, nil, errNotFound
	}

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

// isPlus returns true if path parameter is plus path parameter.
func (router *Router) isPlus(segment string) bool {
	var isPlus bool
	if strings.HasPrefix(segment, plusParamPrefix) {
		isPlus = true
	}
	return isPlus
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
