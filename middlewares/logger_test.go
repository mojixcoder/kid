//go:build go1.21

package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mojixcoder/kid"
	"github.com/stretchr/testify/assert"
)

type logRecord struct {
	Msg       string    `json:"msg"`
	Time      time.Time `json:"time"`
	LatenyMS  int64     `json:"latency_ms"`
	Latency   string    `json:"latency"`
	Status    int       `json:"status"`
	Route     string    `json:"route"`
	Path      string    `json:"path"`
	Method    string    `json:"method"`
	UserAgent string    `json:"user_agent"`
}

func TestNewLogger(t *testing.T) {
	middleware := NewLogger()

	assert.NotNil(t, middleware)
}

func TestSetLoggerDefaults(t *testing.T) {
	var cfg LoggerConfig

	setLoggerDefaults(&cfg)

	assert.Equal(t, DefaultLoggerConfig.Out, cfg.Out)
	assert.Equal(t, DefaultLoggerConfig.Logger, cfg.Logger)
	assert.Equal(t, DefaultLoggerConfig.Level, cfg.Level)
	assert.Equal(t, DefaultLoggerConfig.ServerErrorLevel, cfg.ServerErrorLevel)
	assert.Equal(t, DefaultLoggerConfig.ClientErrorLevel, cfg.ClientErrorLevel)
	assert.Equal(t, DefaultLoggerConfig.SuccessLevel, cfg.SuccessLevel)
	assert.Equal(t, DefaultLoggerConfig.Type, cfg.Type)
}

func TestLoggerConfig_getLogger(t *testing.T) {
	var cfg LoggerConfig
	setLoggerDefaults(&cfg)

	logger := cfg.getLogger()
	assert.IsType(t, &slog.JSONHandler{}, logger.Handler())

	cfg.Type = TypeText

	logger = cfg.getLogger()
	assert.IsType(t, &slog.TextHandler{}, logger.Handler())

	cfg.Logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	assert.Equal(t, cfg.Logger, cfg.getLogger())

	assert.PanicsWithValue(t, "invalid logger type", func() {
		cfg.Logger = nil
		cfg.Type = ""
		cfg.getLogger()
	})
}

func TestNewLoggerWithConfig(t *testing.T) {
	var buf bytes.Buffer

	cfg := DefaultLoggerConfig
	cfg.Out = &buf

	k := kid.New()
	k.Use(NewLoggerWithConfig(cfg))

	k.Get("/", func(c *kid.Context) {
		time.Sleep(time.Millisecond)
		c.String(http.StatusOK, "Ok")
	})

	k.Get("/server-error", func(c *kid.Context) {
		time.Sleep(time.Millisecond)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	})

	k.Get("/not-found", func(c *kid.Context) {
		time.Sleep(time.Millisecond)
		c.String(http.StatusNotFound, "Not Found")
	})

	testCases := []struct {
		path   string
		msg    string
		status int
	}{
		{path: "/not-found", msg: "CLIENT ERROR", status: http.StatusNotFound},
		{path: "/", msg: "SUCCESS", status: http.StatusOK},
		{path: "/server-error", msg: "SERVER ERROR", status: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.msg, func(t *testing.T) {
			res := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			req.Header.Set("User-Agent", "Go Test")

			k.ServeHTTP(res, req)

			var logRecord logRecord
			err := json.Unmarshal(buf.Bytes(), &logRecord)
			assert.NoError(t, err)

			buf.Reset()

			assert.Equal(t, testCase.status, logRecord.Status)
			assert.Equal(t, testCase.path, logRecord.Path)
			assert.Equal(t, testCase.path, logRecord.Route)
			assert.Equal(t, http.MethodGet, logRecord.Method)
			assert.Equal(t, "Go Test", logRecord.UserAgent)
			assert.NotZero(t, logRecord.Time)
			assert.NotEmpty(t, logRecord.Latency)
			assert.NotEmpty(t, logRecord.LatenyMS)
			assert.Equal(t, testCase.msg, logRecord.Msg)
		})
	}
}

func TestLogger_Skipper(t *testing.T) {
	var buf bytes.Buffer

	cfg := DefaultLoggerConfig
	cfg.Out = &buf
	cfg.Skipper = func(c *kid.Context) bool {
		return true
	}

	k := kid.New()
	k.Use(NewLoggerWithConfig(cfg))

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	k.ServeHTTP(res, req)

	assert.Empty(t, buf.Bytes())
}
