package kid

import (
	"net/http"
	"net/url"
	"sync"
)

type (
	HandlerFunc func(c Context) error

	MiddlewareFunc func(next HandlerFunc) HandlerFunc

	ErrorHandler func(c Context, err error)

	Map map[string]interface{}

	Kid struct {
		router                  Router
		middlewares             []MiddlewareFunc
		notFoundHandler         HandlerFunc
		methodNotAllowedHandler HandlerFunc
		errorHandler            ErrorHandler
		jsonSerializer          JSONSerializer
		pool                    sync.Pool
	}
)

func New() *Kid {
	kid := Kid{
		router:                  newRouter(),
		middlewares:             make([]MiddlewareFunc, 0),
		notFoundHandler:         defaultNotFoundHandler,
		methodNotAllowedHandler: defaultMethodNotAllowedHandler,
		errorHandler:            defaultErrorHandler,
		jsonSerializer:          defaultJSONSerializer{},
	}

	kid.pool.New = func() any {
		return kid.newContext()
	}

	return &kid
}

func (k *Kid) newContext() *context {
	ctx := context{
		storage: make(Map),
		params:  make(Params),
		kid:     k,
	}
	return &ctx
}

func (k *Kid) Run(address ...string) error {
	addr := resolveAddress(address)

	return http.ListenAndServe(addr, k)
}

func (k *Kid) Use(middleware MiddlewareFunc) {
	if middleware == nil {
		panic("middleware cannot be nil")
	}

	k.middlewares = append(k.middlewares, middleware)
}

func (k *Kid) GET(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodGet}, middlewares)
}

func (k *Kid) POST(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodPost}, middlewares)
}

func (k *Kid) PUT(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodPut}, middlewares)
}

func (k *Kid) PATCH(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodPatch}, middlewares)
}

func (k *Kid) DELETE(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodDelete}, middlewares)
}

func (k *Kid) HEAD(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodHead}, middlewares)
}

func (k *Kid) OPTIONS(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodOptions}, middlewares)
}

func (k *Kid) CONNECT(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodConnect}, middlewares)
}

func (k *Kid) TRACE(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, []string{http.MethodTrace}, middlewares)
}

func (k *Kid) ANY(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	methods := []string{
		http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodHead,
		http.MethodOptions, http.MethodConnect, http.MethodTrace,
	}
	k.router.add(path, handler, methods, middlewares)
}

func (k *Kid) ADD(path string, handler HandlerFunc, methods []string, middlewares ...MiddlewareFunc) {
	k.router.add(path, handler, methods, middlewares)
}

func (k *Kid) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := k.pool.Get().(*context)
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

func (k *Kid) applyMiddlewaresToHandler(handler HandlerFunc, middlewares ...MiddlewareFunc) HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func getPath(u *url.URL) string {
	if u.RawPath != "" {
		return u.RawPath
	}
	return u.Path
}

func resolveAddress(addresses []string) string {
	if len(addresses) == 0 {
		return ":2376"
	}
	return addresses[0]
}
