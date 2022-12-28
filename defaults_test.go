package kid

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupKid() *Kid {
	k := New()

	k.POST("/post", func(c *Context) error {
		return c.JSON(http.StatusOK, Map{"method": c.Request().Method})
	})

	k.GET("/http-error", func(c *Context) error {
		return NewHTTPError(http.StatusBadRequest)
	})

	k.GET("/error", func(c *Context) error {
		return errors.New("something went wrong")
	})

	k.HEAD("/error-head", func(c *Context) error {
		return NewHTTPError(http.StatusBadRequest)
	})

	return k
}

func TestDefaultNotFoundHandler(t *testing.T) {
	k := setupKid()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/not_found", nil)
	assert.NoError(t, err)

	k.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "{\"message\":\"Not Found\"}\n", w.Body.String())
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestDefaultMethodNotAllowedHandler(t *testing.T) {
	k := setupKid()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/post", nil)
	assert.NoError(t, err)

	k.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, "{\"message\":\"Method Not Allowed\"}\n", w.Body.String())
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestDefaultErrorHandler(t *testing.T) {
	k := setupKid()

	// Retuns a proper response based on return HTTP error.
	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/http-error", nil)
	assert.NoError(t, err)

	k.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "{\"message\":\"Bad Request\"}\n", w.Body.String())
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Always internal server error is returned when returned error is not a HTTP error.
	w = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodGet, "/error", nil)
	assert.NoError(t, err)

	k.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "{\"message\":\"Internal Server Error\"}\n", w.Body.String())
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// No body is returned when request method is HEAD.
	w = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodHead, "/error-head", nil)
	assert.NoError(t, err)

	k.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "", w.Body.String())
	assert.Equal(t, "", w.Header().Get("Content-Type"))
}
