package kid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResponse(t *testing.T) {
	w := httptest.NewRecorder()
	res := newResponse(w).(*response)

	assert.Equal(t, w, res.ResponseWriter)
	assert.Equal(t, http.StatusOK, res.status)
	assert.Zero(t, res.Size())
	assert.False(t, res.Written())
}

func TestResponseWriterWriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	res := newResponse(w).(*response)

	res.WriteHeader(http.StatusAccepted)

	assert.Equal(t, http.StatusAccepted, res.status)
	assert.False(t, res.Written())

	res.WriteHeaderNow()

	// Won't write again because header is already written.
	res.WriteHeader(http.StatusBadRequest)

	assert.Equal(t, http.StatusAccepted, res.status)
}

func TestResponseWriterWriteHeaderNow(t *testing.T) {
	w := httptest.NewRecorder()
	res := newResponse(w).(*response)

	res.WriteHeader(http.StatusAccepted)
	res.WriteHeaderNow()

	assert.True(t, res.Written())
}

func TestResponseWriterSize(t *testing.T) {
	w := httptest.NewRecorder()
	res := newResponse(w)

	n1, err := res.Write([]byte("Hello"))
	assert.NoError(t, err)

	n2, err := res.Write([]byte("Bye"))
	assert.NoError(t, err)

	assert.Equal(t, 8, n1+n2)
	assert.Equal(t, n1+n2, res.Size())
}

func TestResponseWriterWritten(t *testing.T) {
	w := httptest.NewRecorder()
	res := newResponse(w)

	assert.False(t, res.Written())

	res.WriteHeaderNow()

	assert.True(t, res.Written())
}

func TestResponseWriterFlush(t *testing.T) {
	k := New()

	k.GET("/", func(c *Context) error {
		c.Response().WriteHeader(http.StatusBadGateway)
		c.Response().Flush()
		return nil
	})

	srv := httptest.NewServer(k)
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

func TestResponseWriterHijack(t *testing.T) {
	w := httptest.NewRecorder()
	res := newResponse(w)

	assert.Panics(t, func() {
		_, _, _ = res.Hijack()
	})
	assert.True(t, res.Written())
}
