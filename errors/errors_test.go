package errors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPError(t *testing.T) {
	err := NewHTTPError(http.StatusCreated)

	assert.Error(t, err)
	assert.Equal(t, http.StatusCreated, err.Code)
	assert.Equal(t, http.StatusText(http.StatusCreated), err.Message)
	assert.Nil(t, err.Err)
}

func TestWithMessage(t *testing.T) {
	err := NewHTTPError(http.StatusOK).WithMessage("new message")

	assert.Equal(t, http.StatusOK, err.Code)
	assert.Equal(t, "new message", err.Message)
	assert.Nil(t, err.Err)
}

func TestWithError(t *testing.T) {
	someErr := errors.New("some error")
	err := NewHTTPError(http.StatusOK).WithError(someErr)

	assert.Equal(t, http.StatusOK, err.Code)
	assert.Equal(t, http.StatusText(http.StatusOK), err.Message)
	assert.ErrorIs(t, err.Err, someErr)
}

func TestError(t *testing.T) {
	err := NewHTTPError(http.StatusOK)

	assert.Equal(t, `{"code": 200, "message": "OK"}`, err.Error())

	err.WithMessage("something went wrong").WithError(errors.New("some error"))

	assert.Equal(t, `{"code": 200, "message": "something went wrong", "error": "some error"}`, err.Error())
}

func TestUnwrap(t *testing.T) {
	someErr := errors.New("some error")
	err := NewHTTPError(http.StatusForbidden).WithError(someErr)

	newErr := fmt.Errorf("some new error: %w", err)

	unwrapedErr := errors.Unwrap(newErr)
	assert.ErrorIs(t, unwrapedErr, err)

	unwrapedErr = errors.Unwrap(unwrapedErr)
	assert.ErrorIs(t, unwrapedErr, someErr)

	err = NewHTTPError(http.StatusForbidden)
	assert.NoError(t, errors.Unwrap(err))
}
