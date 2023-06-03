package middlewares

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mojixcoder/kid"
)

// CorsConfig is the config used to build CORS middleware.
type CorsConfig struct {
	// AllowedOrigins specifies which origins can access the resource.
	// If "*" is in the list, all origins will be allowed.
	//
	// Defaults to ["*"]
	AllowedOrigins []string

	// AllowOriginFunc is a custom function for validating the origin.
	// The origin will always be set and you don't need to check that in this function.
	//
	// If you set this function the rest of validation logic will be ignored.
	//
	// Defaults to nil.
	AllowOriginFunc func(c *kid.Context, origin string) bool

	// AllowedMethods is the list of allowed HTTP methods.
	//
	// Defaults to ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"].
	AllowedMethods []string

	// AllowedHeaders is the list of the custom headers which are allowed to be sent.
	//
	// If "*" is in the list, all headers will be allowed.
	AllowedHeaders []string

	// ExposedHeaders a list of headers that clients are allowed to access.
	//
	// Defaults to [].
	ExposedHeaders []string

	// MaxAge is the maximum duration that the response to the preflight request can be cached before another call is made.
	// In second percision.
	//
	// Will not be used if 0.
	// Defaults to 0.
	MaxAge time.Duration

	// AllowCredentials if true, cookies will be allowed to be included in cross-site HTTP requests.
	//
	// defaults to false.
	AllowCredentials bool

	// AllowPrivateNetwork if true, allow requests from sites on “public” IP to this server on a “private” IP.
	//
	// defaults to false.
	AllowPrivateNetwork bool

	// allowAllOrigins will be true if "*" is in allowed origins.
	allowAllOrigins bool
}

// DefaultCorsConfig is the default CORS config.
var DefaultCorsConfig = CorsConfig{
	AllowedOrigins: []string{"*"},
	AllowedMethods: []string{
		http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodOptions,
	},
}

// NewCors returns a new CORS config.
func NewCors() kid.MiddlewareFunc {
	return NewCorsWithConfig(DefaultCorsConfig)
}

// NewCorsWithConfig returns a new CORS middleware with the given config.
func NewCorsWithConfig(cfg CorsConfig) kid.MiddlewareFunc {
	setDefaults(&cfg)

	allowedMethods := strings.Join(cfg.AllowedMethods, ", ")
	allowedHeaders := strings.Join(cfg.AllowedHeaders, ", ")
	exposedHeaders := strings.Join(cfg.ExposedHeaders, ", ")
	maxAge := strconv.Itoa(int(cfg.MaxAge.Seconds()))
	allowCreds := "false"
	if cfg.AllowCredentials {
		allowCreds = "true"
	}

	return func(next kid.HandlerFunc) kid.HandlerFunc {
		return func(c *kid.Context) {
			req := c.Request()
			header := c.Response().Header()
			preflight := isPreflight(req)

			header.Set("Vary", "Origin")

			origin := req.Header.Get("Origin")
			if origin == "" {
				next(c)
				return
			}

			if !cfg.isAllowedOrigin(c, origin) {
				next(c)
				return
			}

			if cfg.allowAllOrigins && !cfg.AllowCredentials {
				header.Set("Access-Control-Allow-Origin", "*")
			} else {
				header.Set("Access-Control-Allow-Origin", origin)
			}

			if cfg.AllowPrivateNetwork && req.Header.Get("Access-Control-Request-Private-Network") == "true" {
				header.Set("Access-Control-Allow-Private-Network", "true")
			}

			setHeader(header, "Access-Control-Allow-Credentials", allowCreds, "false")
			setHeader(header, "Access-Control-Expose-Headers", exposedHeaders, "")

			switch preflight {
			case false:
				next(c)
			case true:
				setHeader(header, "Access-Control-Allow-Methods", allowedMethods, "")
				setHeader(header, "Access-Control-Allow-Headers", allowedHeaders, "")
				setHeader(header, "Access-Control-Max-Age", maxAge, "0")

				c.NoContent(http.StatusNoContent)
			}
		}
	}
}

// isPreflight checks if this is a preflight request.
func isPreflight(req *http.Request) bool {
	return req.Method == http.MethodOptions && req.Header.Get("Access-Control-Request-Method") != ""
}

// isAllowedOrigin validates the origin.
func (cors *CorsConfig) isAllowedOrigin(c *kid.Context, origin string) bool {
	if cors.AllowOriginFunc != nil {
		return cors.AllowOriginFunc(c, origin)
	}

	if cors.allowAllOrigins {
		return true
	}

	for _, v := range cors.AllowedOrigins {
		if v == "*" {
			cors.allowAllOrigins = true
			return true
		}
		if v == origin {
			return true
		}
	}

	return false
}

// setHeader sets the header if not empty.
func setHeader(header http.Header, key, value, emptyValue string) {
	if value != emptyValue {
		header.Set(key, value)
	}
}

// setDefaults sets the default CORS configs.
func setDefaults(cfg *CorsConfig) {
	if len(cfg.AllowedOrigins) == 0 {
		cfg.AllowedOrigins = DefaultCorsConfig.AllowedOrigins
	}

	if len(cfg.AllowedMethods) == 0 {
		cfg.AllowedMethods = DefaultCorsConfig.AllowedMethods
	}
}
