package middlewares

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/mojixcoder/kid"
)

// RecoveryConfig is the config used to build a Recovery middleware.
type RecoveryConfig struct {
	// LogRecovers logs when a recovery happens, only in debug mode.
	LogRecovers bool

	// PrintStacktrace prints the entire stacktrace if true, only in debug mode.
	PrintStacktrace bool

	// Writer is the writer for logging recoveries and stacktraces.
	Writer io.Writer

	// OnRecovery is the function which will be called when a recovery occurs.
	OnRecovery func(c *kid.Context, err any)
}

// DefaultRecoverConfig is the default Recovery config.
var DefaultRecoveryConfig = RecoveryConfig{
	LogRecovers: true,
	Writer:      os.Stdout,
	OnRecovery: func(c *kid.Context, err any) {
		c.JSON(http.StatusInternalServerError, kid.Map{"message": http.StatusText(http.StatusInternalServerError)})
	},
}

// NewRecovery returns a new Recovery middleware.
func NewRecovery() kid.MiddlewareFunc {
	return NewRecoveryWithConfig(DefaultRecoveryConfig)
}

// NewRecoveryWithConfig returns a new Recovery middleware with the given config.
func NewRecoveryWithConfig(cfg RecoveryConfig) kid.MiddlewareFunc {
	return func(next kid.HandlerFunc) kid.HandlerFunc {
		return func(c *kid.Context) {
			defer func() {
				if err := recover(); err != nil {
					if c.Debug() {
						if cfg.LogRecovers {
							fmt.Fprintf(cfg.Writer, "[RECOVERY] panic recovered: %v\n", err)
						}

						if cfg.PrintStacktrace {
							stack := debug.Stack()
							fmt.Fprintf(cfg.Writer, "%s", string(stack))
						}
					}

					if cfg.OnRecovery != nil {
						cfg.OnRecovery(c, err)
					}
				}
			}()

			next(c)
		}
	}
}
