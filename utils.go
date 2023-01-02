package kid

import (
	"net/http"
	"net/url"
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

// getPath returns request's path.
func getPath(u *url.URL) string {
	if u.RawPath != "" {
		return u.RawPath
	}
	return u.Path
}

// resolveAddress returns the address which server will run on.
func resolveAddress(addresses []string) string {
	if len(addresses) == 0 {
		return ":2376"
	}
	return addresses[0]
}
