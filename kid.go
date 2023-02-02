package kid

import (
	"net/http"
	"sync"

	htmlrenderer "github.com/mojixcoder/kid/html_renderer"
	"github.com/mojixcoder/kid/serializer"
)

type (
	// HandlerFunc is the type which serves HTTP requests.
	HandlerFunc func(c *Context) error

	// MiddlewareFunc is the type of middlewares.
	MiddlewareFunc func(next HandlerFunc) HandlerFunc

	// ErrorHandler is the functions that handles errors when a handler returns an error.
	ErrorHandler func(c *Context, err error)

	// Map is a generic map to make it easier to send responses.
	Map map[string]any

	// Kid is the struct that holds everything together.
	//
	// It's a framework instance.
	Kid struct {
		router                  Router
		middlewares             []MiddlewareFunc
		notFoundHandler         HandlerFunc
		methodNotAllowedHandler HandlerFunc
		errorHandler            ErrorHandler
		jsonSerializer          serializer.Serializer
		xmlSerializer           serializer.Serializer
		htmlRenderer            htmlrenderer.HTMLRenderer
		debug                   bool
		pool                    sync.Pool
	}
)

// New returns a new instance of Kid.
func New() *Kid {
	kid := Kid{
		router:                  newRouter(),
		middlewares:             make([]MiddlewareFunc, 0),
		notFoundHandler:         defaultNotFoundHandler,
		methodNotAllowedHandler: defaultMethodNotAllowedHandler,
		errorHandler:            defaultErrorHandler,
		jsonSerializer:          serializer.NewJSONSerializer(),
		xmlSerializer:           serializer.NewXMLSerializer(),
		htmlRenderer:            htmlrenderer.Default(false),
	}

	kid.pool.New = func() any {
		return newContext(&kid)
	}

	return &kid
}

// Run runs HTTP server.
//
// Specifying an address is optional. Default address is :2376.
func (k *Kid) Run(address ...string) error {
	addr := resolveAddress(address)
	return http.ListenAndServe(addr, k)
}

// Use registers a new middleware. The middleware will be applied to all of the routes.
func (k *Kid) Use(middleware MiddlewareFunc) {
	if middleware == nil {
		panic("middleware cannot be nil")
	}

	k.middlewares = append(k.middlewares, middleware)
}

// GET registers a new handler for the given path for GET method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) GET(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodGet}, middlewares)
}

// POST registers a new handler for the given path for POST method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) POST(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodPost}, middlewares)
}

// PUT registers a new handler for the given path for PUT method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) PUT(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodPut}, middlewares)
}

// PATCH registers a new handler for the given path for PATCH method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) PATCH(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodPatch}, middlewares)
}

// DELETE registers a new handler for the given path for DELETE method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) DELETE(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodDelete}, middlewares)
}

// HEAD registers a new handler for the given path for HEAD method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) HEAD(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodHead}, middlewares)
}

// OPTIONS registers a new handler for the given path for OPTIONS method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) OPTIONS(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodOptions}, middlewares)
}

// CONNECT registers a new handler for the given path for CONNECT method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) CONNECT(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodConnect}, middlewares)
}

// TRACE registers a new handler for the given path for TRACE method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) TRACE(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodTrace}, middlewares)
}

// ANY registers a new handler for the given path for all of the HTTP methods.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) ANY(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	methods := []string{
		http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodHead,
		http.MethodOptions, http.MethodConnect, http.MethodTrace,
	}
	k.router.add(path, handler, methods, middlewares)
}

// ADD registers a new handler for the given path for the given methods.
// Specifying at least one method is required.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) ADD(path string, handler HandlerFunc, methods []string, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, methods, middlewares)
}

// Static registers a new route for serving static files.
//
// It uses http.Dir as its file system.
func (k *Kid) Static(urlPath, staticRoot string, middlewares ...MiddlewareFunc) {
	fileServer := newFileServer(urlPath, FS{http.Dir(staticRoot)})

	methods := []string{http.MethodGet}
	path := appendSlash(urlPath) + "{*filePath}"

	k.router.add(path, WrapHandler(fileServer), methods, middlewares)
}

// StaticFS registers a new route for serving static files.
//
// It uses the given file system to serve static files.
func (k *Kid) StaticFS(urlPath string, fs http.FileSystem, middlewares ...MiddlewareFunc) {
	fileServer := newFileServer(urlPath, fs)

	methods := []string{http.MethodGet}
	path := appendSlash(urlPath) + "{*filePath}"

	k.router.add(path, WrapHandler(fileServer), methods, middlewares)
}

// ServeHTTP implements the http.HandlerFunc interface.
func (k *Kid) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := k.pool.Get().(*Context)
	c.reset(r, w)

	route, params, err := k.router.find(getPath(r.URL), r.Method)

	c.setParams(params)

	var handler HandlerFunc

	if err == errNotFound {
		handler = k.applyMiddlewaresToHandler(k.notFoundHandler, k.middlewares...)
	} else if err == errMethodNotAllowed {
		handler = k.applyMiddlewaresToHandler(k.methodNotAllowedHandler, k.middlewares...)
	} else {
		handler = k.applyMiddlewaresToHandler(route.handler, route.middlewares...)
		handler = k.applyMiddlewaresToHandler(handler, k.middlewares...)
	}

	if err := handler(c); err != nil {
		k.errorHandler(c, err)
	}

	k.pool.Put(c)
}

// Debug returns whether we are in debug mode or not.
func (k *Kid) Debug() bool {
	return k.debug
}

// applyMiddlewaresToHandler applies middlewares to the handler and returns the handler.
func (k *Kid) applyMiddlewaresToHandler(handler HandlerFunc, middlewares ...MiddlewareFunc) HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
