//go:build go1.21

package middlewares

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/mojixcoder/kid"
)

type (
	// LoggerConfig is the config used to build logger middleware.
	LoggerConfig struct {
		// Logger is the logger instance.
		// Optional. If set, Out, Level and Type configs won't be used.
		Logger *slog.Logger

		// Out is the writer that logs will be written at.
		// Defaults to os.Stdout.
		Out io.Writer

		// Level is the log level used for initializing a logger instance.
		// Defaults to slog.LevelInfo.
		Level slog.Leveler

		// SuccessLevel is the log level when status code < 400.
		// Defaults to slog.LevelInfo.
		SuccessLevel slog.Leveler

		// ClientErrorLevel is the log level when status code is between 400 and 499.
		// Defaults to slog.LevelWarn.
		ClientErrorLevel slog.Leveler

		// ServerErrorLevel is the log level when status code >= 500.
		// Defaults to slog.LevelError.
		ServerErrorLevel slog.Leveler

		// Type is the logger type.
		// Defaults to JSON.
		Type LoggerType

		// Skipper is a function used for skipping middleware execution.
		// Defaults to nil.
		Skipper func(c *kid.Context) bool
	}

	// LoggerType is the type for specifying logger type.
	LoggerType string
)

const (
	// JSONLogger is the JSON logger type.
	TypeJSON LoggerType = "JSON"

	// TextLogger is the text logger type.
	TypeText LoggerType = "TEXT"
)

// DefaultLoggerConfig is the default logger config.
var DefaultLoggerConfig = LoggerConfig{
	Out:              os.Stdout,
	Level:            slog.LevelInfo,
	SuccessLevel:     slog.LevelInfo,
	ClientErrorLevel: slog.LevelWarn,
	ServerErrorLevel: slog.LevelError,
	Type:             TypeJSON,
}

// NewLogger returns a new logger middleware.
func NewLogger() kid.MiddlewareFunc {
	return NewLoggerWithConfig(DefaultLoggerConfig)
}

// NewLoggerWithConfig returns a new logger middleware with the given config.
func NewLoggerWithConfig(cfg LoggerConfig) kid.MiddlewareFunc {
	setLoggerDefaults(&cfg)

	logger := cfg.getLogger()

	successLvl := cfg.SuccessLevel.Level()
	clientErrLvl := cfg.ClientErrorLevel.Level()
	serverErrLvl := cfg.ServerErrorLevel.Level()

	return func(next kid.HandlerFunc) kid.HandlerFunc {
		return func(c *kid.Context) {
			// Skip if necessary.
			if cfg.Skipper != nil && cfg.Skipper(c) {
				next(c)
				return
			}

			start := time.Now()

			next(c)

			end := time.Now()
			elapsed := end.Sub(start)

			status := c.Response().Status()

			attrs := []slog.Attr{
				slog.Time("time", end),
				slog.Int64("latency_ms", elapsed.Milliseconds()),
				slog.String("latency", elapsed.String()),
				slog.Int("status", status),
				slog.String("route", c.Route()),
				slog.String("path", c.Path()),
				slog.String("method", c.Method()),
				slog.String("user_agent", c.GetRequestHeader("User-Agent")),
			}

			if status < 400 {
				logger.LogAttrs(context.Background(), successLvl, "SUCCESS", attrs...)
			} else if status <= 499 {
				logger.LogAttrs(context.Background(), clientErrLvl, "CLIENT ERROR", attrs...)
			} else { // 5xx status codes.
				logger.LogAttrs(context.Background(), serverErrLvl, "SERVER ERROR", attrs...)
			}
		}
	}
}

// getLogger returns the appropriate logger instance.
func (cfg LoggerConfig) getLogger() *slog.Logger {
	if cfg.Logger != nil {
		return cfg.Logger
	}

	switch cfg.Type {
	case TypeJSON:
		return slog.New(slog.NewJSONHandler(cfg.Out, &slog.HandlerOptions{Level: cfg.Level}))
	case TypeText:
		return slog.New(slog.NewTextHandler(cfg.Out, &slog.HandlerOptions{Level: cfg.Level}))
	default:
		panic("invalid logger type")
	}
}

// setLoggerDefaults sets logger default values.
func setLoggerDefaults(cfg *LoggerConfig) {
	if cfg.Out == nil {
		cfg.Out = DefaultLoggerConfig.Out
	}

	if cfg.Level == nil {
		cfg.Level = DefaultLoggerConfig.Level
	}

	if cfg.SuccessLevel == nil {
		cfg.SuccessLevel = DefaultLoggerConfig.SuccessLevel
	}

	if cfg.ClientErrorLevel == nil {
		cfg.ClientErrorLevel = DefaultLoggerConfig.ClientErrorLevel
	}

	if cfg.ServerErrorLevel == nil {
		cfg.ServerErrorLevel = DefaultLoggerConfig.ServerErrorLevel
	}

	if cfg.Type == "" {
		cfg.Type = DefaultLoggerConfig.Type
	}
}
