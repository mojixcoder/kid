package kid

import (
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
