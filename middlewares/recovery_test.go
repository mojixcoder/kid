package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mojixcoder/kid"
	"github.com/stretchr/testify/assert"
)

var flag bool

var recoveryHandler kid.HandlerFunc = func(c *kid.Context) {
	panic("err")
}

func TestNewRecoveryWithConfig(t *testing.T) {
	k := kid.New()
	var buf bytes.Buffer

	recovery := NewRecoveryWithConfig(RecoveryConfig{LogRecovers: true, Writer: &buf})

	ctx := k.NewContext(nil, httptest.NewRecorder())
	recovery(recoveryHandler)(ctx)
	assert.Equal(t, "[RECOVERY] panic recovered: err\n", buf.String())

	buf.Reset()
	recovery = NewRecoveryWithConfig(RecoveryConfig{PrintStacktrace: true, Writer: &buf})

	ctx = k.NewContext(nil, httptest.NewRecorder())
	recovery(recoveryHandler)(ctx)
	assert.NotEmpty(t, buf.String())

	buf.Reset()
	k.ApplyOptions(kid.WithDebug(false))
	recovery(recoveryHandler)(ctx)
	assert.Empty(t, buf.String())

	buf.Reset()
	recovery = NewRecoveryWithConfig(RecoveryConfig{
		OnRecovery: func(c *kid.Context, err any) {
			flag = true
		},
	})

	ctx = k.NewContext(nil, httptest.NewRecorder())
	recovery(recoveryHandler)(ctx)
	assert.True(t, flag)
}

func TestNewRecovery(t *testing.T) {
	k := kid.New()

	recovery := NewRecovery()

	res := httptest.NewRecorder()
	ctx := k.NewContext(nil, res)
	recovery(recoveryHandler)(ctx)

	assert.Equal(t, res.Code, http.StatusInternalServerError)
	assert.Equal(t, "{\"message\":\"Internal Server Error\"}\n", res.Body.String())
}
