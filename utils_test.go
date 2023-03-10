package kid

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveAddress(t *testing.T) {
	addr := resolveAddress([]string{})

	assert.Equal(t, ":2376", addr)

	addr = resolveAddress([]string{":2377", "2378"})
	assert.Equal(t, ":2377", addr)
}

func TestGetPath(t *testing.T) {
	u, err := url.Parse("http://localhost/foo%25fbar")
	assert.NoError(t, err)

	assert.Empty(t, u.RawPath)
	assert.Equal(t, u.Path, getPath(u))

	u, err = url.Parse("http://localhost/foo%fbar")
	assert.NoError(t, err)

	assert.NotEmpty(t, u.RawPath)
	assert.Equal(t, u.RawPath, getPath(u))
}

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
