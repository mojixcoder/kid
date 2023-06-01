package kid

import (
	"net/http"
)

// WrapHandlerFunc wraps a http.HandlerFunc and returns a kid.HandlerFunc.
func WrapHandlerFunc(f http.HandlerFunc) HandlerFunc {
	return func(c *Context) {
		f(c.Response(), c.Request())
	}
}

// WrapHandler wraps a http.Handler and returns a kid.HandlerFunc.
func WrapHandler(h http.Handler) HandlerFunc {
	return func(c *Context) {
		h.ServeHTTP(c.Response(), c.Request())
	}
}
