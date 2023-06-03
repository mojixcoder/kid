package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/mojixcoder/kid"
	"github.com/stretchr/testify/assert"
)

func TestSetHeader(t *testing.T) {
	testCases := []struct {
		key         string
		val         string
		emptyVal    string
		expectedRes string
	}{
		{key: "empty_max_age", val: strconv.Itoa(int(time.Duration(0).Seconds())), emptyVal: "0", expectedRes: ""},
		{key: "max_age", val: strconv.Itoa(int((time.Hour).Seconds())), emptyVal: "0", expectedRes: "3600"},
		{key: "empty_headers", val: "", emptyVal: "", expectedRes: ""},
		{key: "headers", val: "headers", emptyVal: "", expectedRes: "headers"},
		{key: "empty_creds", val: "false", emptyVal: "false", expectedRes: ""},
		{key: "creds", val: "true", emptyVal: "false", expectedRes: "true"},
	}

	header := make(http.Header)

	for _, testCase := range testCases {
		t.Run(testCase.key, func(t *testing.T) {
			setHeader(header, testCase.key, testCase.val, testCase.emptyVal)
			assert.Equal(t, testCase.expectedRes, header.Get(testCase.key))
		})
	}
}

func TestSetCorsDefaults(t *testing.T) {
	cors := &CorsConfig{}

	setCorsDefaults(cors)
	assert.Equal(t, DefaultCorsConfig.AllowedOrigins, cors.AllowedOrigins)
	assert.Equal(t, DefaultCorsConfig.AllowedMethods, cors.AllowedMethods)

	cors = &CorsConfig{AllowedMethods: []string{http.MethodConnect}}

	setCorsDefaults(cors)
	assert.Equal(t, []string{http.MethodConnect}, cors.AllowedMethods)
}

func TestIsPreflight(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	assert.False(t, isPreflight(req))

	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	assert.True(t, isPreflight(req))
}

func TestCorsConfig_isAllowedOrigin(t *testing.T) {
	cfg := CorsConfig{AllowedOrigins: []string{"http://localhost:2376"}}

	assert.True(t, cfg.isAllowedOrigin(nil, "http://localhost:2376"))
	assert.False(t, cfg.isAllowedOrigin(nil, "http://localhost:2377"))
	assert.False(t, cfg.allowAllOrigins)

	cfg.AllowedOrigins = []string{"*"}

	assert.True(t, cfg.isAllowedOrigin(nil, "http://localhost:2376"))
	assert.True(t, cfg.allowAllOrigins)
	assert.True(t, cfg.isAllowedOrigin(nil, "http://localhost:2377"))

	cfg.AllowOriginFunc = func(c *kid.Context, origin string) bool {
		return false
	}

	assert.False(t, cfg.isAllowedOrigin(nil, "http://localhost:2376"))
}

func TestNewCors(t *testing.T) {
	k := kid.New()
	k.Use(NewCors())

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Add("Access-Control-Request-Method", http.MethodPost)
	req.Header.Add("Origin", "http://localhost:2376")

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNoContent, res.Code)
	assert.Equal(t, "Origin", res.Header().Get("Vary"))
	assert.Equal(t, "*", res.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", res.Header().Get("Access-Control-Allow-Methods"))

	res = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Add("Origin", "http://localhost:2376")

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Equal(t, "Origin", res.Header().Get("Vary"))
	assert.Equal(t, "*", res.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, res.Header().Get("Access-Control-Allow-Methods"))
}

func TestNewCorsWithConfig(t *testing.T) {
	cfg := CorsConfig{
		AllowedOrigins:      []string{"http://localhost:2376"},
		AllowedMethods:      []string{http.MethodGet, http.MethodPost},
		AllowedHeaders:      []string{"Content-Type", "Accept"},
		ExposedHeaders:      []string{"User-Agent"},
		AllowCredentials:    true,
		AllowPrivateNetwork: true,
		MaxAge:              24 * time.Hour,
	}

	k := kid.New()
	k.Use(NewCorsWithConfig(cfg))

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Add("Access-Control-Request-Method", http.MethodPost)
	req.Header.Add("Origin", "http://localhost:2376")
	req.Header.Add("Access-Control-Request-Private-Network", "true")

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNoContent, res.Code)
	assert.Equal(t, "Origin", res.Header().Get("Vary"))
	assert.Equal(t, "http://localhost:2376", res.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST", res.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Accept", res.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "User-Agent", res.Header().Get("Access-Control-Expose-Headers"))
	assert.Equal(t, "true", res.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "true", res.Header().Get("Access-Control-Allow-Private-Network"))
	assert.Equal(t, "86400", res.Header().Get("Access-Control-Max-Age"))

	res = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Add("Access-Control-Request-Method", http.MethodPost)

	k.ServeHTTP(res, req)
	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Empty(t, res.Header().Get("Access-Control-Allow-Origin"))

	res = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Add("Access-Control-Request-Method", http.MethodPost)
	req.Header.Add("Origin", "http://localhost:4000")

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Empty(t, res.Header().Get("Access-Control-Allow-Origin"))
}
