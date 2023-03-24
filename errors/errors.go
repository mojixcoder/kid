// Package errors implements error interface.
//
// HTTPError has first class support in Kid and is used for returning proper responses when an error happens.
package errors

import (
	"fmt"
	"net/http"
)

// HTTPError is the struct for returning HTTP errors.
//
// Can be used by Kid's default error handler.
type HTTPError struct {
	Code    int   `json:"-"`
	Message any   `json:"message"`
	Err     error `json:"-"`
}

// Verifying interface compliance.
var _ error = (*HTTPError)(nil)

// Error implements the error interface and returns error as string.
func (e *HTTPError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf(`{"code": %d, "message": %q}`, e.Code, e.Message)
	}
	return fmt.Sprintf(`{"code": %d, "message": %q, "error": %q}`, e.Code, e.Message, e.Err)
}

// Unwrap implements the errors.Unwrap interface.
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// WithError sets the internal error of HTTP error.
func (e *HTTPError) WithError(err error) *HTTPError {
	e.Err = err
	return e
}

// WithMessage sets the error message.
//
// This message will be sent in response in Kid's default error handler. So it should of a type which can be converted to JSON.
func (e *HTTPError) WithMessage(message any) *HTTPError {
	e.Message = message
	return e
}

// NewHTTPError returns a new HTTP error.
func NewHTTPError(code int) *HTTPError {
	err := HTTPError{Code: code, Message: http.StatusText(code)}
	return &err
}
