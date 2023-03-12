package kid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapHandler(t *testing.T) {
	k1 := New()
	k2 := New()

	k2.Get("/test", func(c *Context) error {
		return c.JSONByte(http.StatusOK, []byte(`{"status": "ok"}`))
	})

	k1.Get("/test", WrapHandler(k2))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()

	k1.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"status": "ok"}`, res.Body.String())
}

func TestWrapHandlerFunc(t *testing.T) {
	k := New()

	k.Get("/test", WrapHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status": "ok"}`))
		assert.NoError(t, err)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"status": "ok"}`, res.Body.String())
}
