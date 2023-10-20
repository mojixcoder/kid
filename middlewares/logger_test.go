//go:build go1.21

package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	middleware := NewLogger()

	assert.NotNil(t, middleware)
}
