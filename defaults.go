package kid

import (
	"net/http"
)

var (
	// defaultNotFoundHandler is Kid's default not found handler.
	//
	// It will be used when request doesn't match any routes.
	defaultNotFoundHandler HandlerFunc = func(c *Context) {
		c.JSON(http.StatusNotFound, Map{"message": http.StatusText(http.StatusNotFound)})
	}

	// defaultMethodNotAllowedHandler is Kid's default method not allowed handler.
	//
	// It will be used when request matches a route but its method doesn't match route's method.
	defaultMethodNotAllowedHandler HandlerFunc = func(c *Context) {
		c.JSON(http.StatusMethodNotAllowed, Map{"message": http.StatusText(http.StatusMethodNotAllowed)})
	}
)
