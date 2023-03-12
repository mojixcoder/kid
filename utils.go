package kid

import (
	"net/http"
)

// WrapHandlerFunc wraps a http.HandlerFunc and returns a kid.HandlerFunc.
func WrapHandlerFunc(f http.HandlerFunc) HandlerFunc {
	return func(c *Context) error {
		f(c.Response(), c.Request())
		return nil
	}
}

// WrapHandler wraps a http.Handler and returns a kid.HandlerFunc.
func WrapHandler(h http.Handler) HandlerFunc {
	return func(c *Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
