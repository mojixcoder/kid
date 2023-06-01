package kid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupKid() *Kid {
	k := New()

	k.Post("/post", func(c *Context) {
		c.JSON(http.StatusOK, Map{"method": c.Request().Method})
	})

	return k
}

func TestDefaultNotFoundHandler(t *testing.T) {
	k := setupKid()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/not-found", nil)
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
