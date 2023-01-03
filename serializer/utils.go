package serializer

import "github.com/mojixcoder/kid/errors"

// newHTTPErrorFromError returns a new HTTP error from an error.
func newHTTPErrorFromError(code int, err error) *errors.HTTPError {
	return errors.NewHTTPError(code).WithMessage(err.Error()).WithError(err)
}
