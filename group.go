package kid

import "net/http"

// Group is for creating groups of routes.
//
// It doesn't actually group them but
// it's kind of an abstraction to make it easier to make a group of routes.
type Group struct {
	kid         *Kid
	prefix      string
	middlewares []MiddlewareFunc
}

// newGroup returns a new group.
func newGroup(k *Kid, prefix string, middlewares ...MiddlewareFunc) Group {
	return Group{kid: k, prefix: prefix, middlewares: middlewares}
}

// Get registers a new handler for the given path for GET method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (g *Group) Get(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	g.Add(path, handler, []string{http.MethodGet}, middlewares...)
}

// Post registers a new handler for the given path for POST method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (g *Group) Post(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	g.Add(path, handler, []string{http.MethodPost}, middlewares...)
}

// Put registers a new handler for the given path for PUT method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (g *Group) Put(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	g.Add(path, handler, []string{http.MethodPut}, middlewares...)
}

// Patch registers a new handler for the given path for PATCH method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (g *Group) Patch(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	g.Add(path, handler, []string{http.MethodPatch}, middlewares...)
}

// Delete registers a new handler for the given path for DELETE method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (g *Group) Delete(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	g.Add(path, handler, []string{http.MethodDelete}, middlewares...)
}

// Head registers a new handler for the given path for HEAD method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (g *Group) Head(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	g.Add(path, handler, []string{http.MethodHead}, middlewares...)
}

// Options registers a new handler for the given path for OPTIONS method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (g *Group) Options(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	g.Add(path, handler, []string{http.MethodOptions}, middlewares...)
}

// Connect registers a new handler for the given path for CONNECT method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (g *Group) Connect(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	g.Add(path, handler, []string{http.MethodConnect}, middlewares...)
}

// Trace registers a new handler for the given path for TRACE method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (g *Group) Trace(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	g.Add(path, handler, []string{http.MethodTrace}, middlewares...)
}

// Any registers a new handler for the given path for all of the HTTP methods.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (g *Group) Any(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	methods := []string{
		http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodHead,
		http.MethodOptions, http.MethodConnect, http.MethodTrace,
	}
	g.Add(path, handler, methods, middlewares...)
}

// Add adds a route to the group routes.
func (g *Group) Add(path string, handler HandlerFunc, methods []string, middlewares ...MiddlewareFunc) {
	path = g.prefix + path
	middlewares = g.combineMiddlewares(middlewares)

	g.kid.Add(path, handler, methods, middlewares...)
}

// Group creates a sub-group for that group.
func (g *Group) Group(prefix string, middlewares ...MiddlewareFunc) Group {
	prefix = g.prefix + prefix
	gMiddlewares := g.combineMiddlewares(middlewares)
	return newGroup(g.kid, prefix, gMiddlewares...)
}

// combineMiddlewares combines the given middlewares with the group middlewares and returns the combined middlewares.
func (g *Group) combineMiddlewares(middlewares []MiddlewareFunc) []MiddlewareFunc {
	gMiddlewares := make([]MiddlewareFunc, 0, len(g.middlewares)+len(middlewares))
	if cap(gMiddlewares) == 0 {
		return nil
	}

	gMiddlewares = append(gMiddlewares, g.middlewares...)
	gMiddlewares = append(gMiddlewares, middlewares...)

	return gMiddlewares
}
