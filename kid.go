package kid

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"runtime"
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

// Version of Kid.
const Version string = "0.1.0"

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
		debug:                   true,
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
	addr := resolveAddress(address, runtime.GOOS)

	k.printDebug(os.Stdout, "Kid version %s\n", Version)
	k.printDebug(os.Stdout, "Starting server at %s\n", addr)
	k.printDebug(os.Stdout, "Quit the server with CONTROL-C\n")

	return http.ListenAndServe(addr, k)
}

// Use registers a new middleware. The middleware will be applied to all of the routes.
func (k *Kid) Use(middleware MiddlewareFunc) {
	panicIfNil(middleware, "middleware cannot be nil")

	k.middlewares = append(k.middlewares, middleware)
}

// Get registers a new handler for the given path for GET method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) Get(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodGet}, middlewares)
}

// Post registers a new handler for the given path for POST method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) Post(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodPost}, middlewares)
}

// Put registers a new handler for the given path for PUT method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) Put(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodPut}, middlewares)
}

// Patch registers a new handler for the given path for PATCH method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) Patch(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodPatch}, middlewares)
}

// Delete registers a new handler for the given path for DELETE method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) Delete(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodDelete}, middlewares)
}

// Head registers a new handler for the given path for HEAD method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) Head(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodHead}, middlewares)
}

// Options registers a new handler for the given path for OPTIONS method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) Options(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodOptions}, middlewares)
}

// Connect registers a new handler for the given path for CONNECT method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) Connect(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodConnect}, middlewares)
}

// Trace registers a new handler for the given path for TRACE method.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) Trace(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodTrace}, middlewares)
}

// Any registers a new handler for the given path for all of the HTTP methods.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) Any(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	methods := []string{
		http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodHead,
		http.MethodOptions, http.MethodConnect, http.MethodTrace,
	}
	k.router.add(path, handler, methods, middlewares)
}

// Group creates a new router group.
//
// Specifying middlewares is optional. Middlewares will be applied to all of the group routes.
func (k *Kid) Group(prefix string, middlewares ...MiddlewareFunc) Group {
	return newGroup(k, prefix, middlewares...)
}

// Add registers a new handler for the given path for the given methods.
// Specifying at least one method is required.
//
// Specifying middlewares is optional. Middlewares will only be applied to this route.
func (k *Kid) Add(path string, handler HandlerFunc, methods []string, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, methods, middlewares)
}

// Static registers a new route for serving static files.
//
// It uses http.Dir as its file system.
func (k *Kid) Static(urlPath, staticRoot string, middlewares ...MiddlewareFunc) {
	k.StaticFS(urlPath, FS{http.Dir(staticRoot)})
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

	if !c.Response().Written() {
		c.Response().WriteHeaderNow()
	}

	k.pool.Put(c)
}

// applyMiddlewaresToHandler applies middlewares to the handler and returns the handler.
func (k *Kid) applyMiddlewaresToHandler(handler HandlerFunc, middlewares ...MiddlewareFunc) HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// Debug returns whether we are in debug mode or not.
func (k *Kid) Debug() bool {
	return k.debug
}

// ApplyOptions applies the given options.
func (k *Kid) ApplyOptions(opts ...Option) {
	for _, opt := range opts {
		panicIfNil(opt, "option cannot be nil")

		opt.apply(k)
	}
}

// printDebug prints logs only in debug mode.
func (k *Kid) printDebug(w io.Writer, format string, values ...any) {
	if k.Debug() {
		fmt.Fprintf(w, "[DEBUG] "+format, values...)
	}
}

// getPath returns request's path.
func getPath(u *url.URL) string {
	if u.RawPath != "" {
		return u.RawPath
	}
	return u.Path
}

// resolveAddress returns the address which server will run on.
func resolveAddress(addresses []string, goos string) string {
	if len(addresses) == 0 {
		if goos == "windows" {
			return "127.0.0.1:2376"
		}
		return "0.0.0.0:2376"
	}
	return addresses[0]
}

// panicIfNil panics if the given parameter is nil.
func panicIfNil(x any, message string) {
	if x == nil {
		panic(message)
	}

	switch reflect.TypeOf(x).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice, reflect.Interface, reflect.Func:
		if reflect.ValueOf(x).IsNil() {
			panic(message)
		}
	}
}
