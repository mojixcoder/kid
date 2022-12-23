package kid

import "net/http"

var (
	defaultNotFoundHandler HandlerFunc = func(c Context) error {
		err := NewHTTPError(http.StatusNotFound)
		return err
	}

	defaultMethodNotAllowedHandler HandlerFunc = func(c Context) error {
		err := NewHTTPError(http.StatusMethodNotAllowed)
		return err
	}

	defaultErrorHandler ErrorHandler = func(c Context, err error) {
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
