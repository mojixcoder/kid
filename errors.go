package kid

import (
	"fmt"
	"net/http"
)

type HTTPError struct {
	Code    int   `json:"-"`
	Message any   `json:"message"`
	Err     error `json:"-"`
}

// Verifying interface compliance.
var _ error = (*HTTPError)(nil)

func (e *HTTPError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf(`{"code": %d, "message": %q}`, e.Code, e.Message)
	}
	return fmt.Sprintf(`{"code": %d, "message": %q, "error": %q}`, e.Code, e.Message, e.Err)
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

func (e *HTTPError) WithError(err error) *HTTPError {
	e.Err = err
	return e
}

func (e *HTTPError) WithMessage(message any) *HTTPError {
	e.Message = message
	return e
}

func NewHTTPError(code int) *HTTPError {
	err := HTTPError{Code: code, Message: http.StatusText(code)}
	return &err
}
