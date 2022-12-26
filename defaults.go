package kid

import "net/http"

var (
	// defaultNotFoundHandler is Kid's default not found handler.
	//
	// It will be used when request doesn't match any routes.
	defaultNotFoundHandler HandlerFunc = func(c *Context) error {
		err := NewHTTPError(http.StatusNotFound)
		return err
	}

	// defaultMethodNotAllowedHandler is Kid's default method not allowed handler.
	//
	// It will be used when request matches a route but its method doesn't match route's method.
	defaultMethodNotAllowedHandler HandlerFunc = func(c *Context) error {
		err := NewHTTPError(http.StatusMethodNotAllowed)
		return err
	}

	// defaultErrorHandler is Kid's default error handler.
	//
	// It will be used when handlers return an error.
	// It can send proper responses when an HTTP error is returned.
	defaultErrorHandler ErrorHandler = func(c *Context, err error) {
		httpErr, ok := err.(*HTTPError)
		if !ok {
			httpErr = NewHTTPError(http.StatusInternalServerError)
		}

		if c.Request().Method == http.MethodHead {
			c.NoContent(httpErr.Code)
		} else {
			c.JSON(httpErr.Code, httpErr)
		}
	}
)
