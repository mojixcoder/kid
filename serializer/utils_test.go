package serializer

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPErrorFromError(t *testing.T) {
	err := errors.New("test error")

	httpErr := newHTTPErrorFromError(http.StatusBadRequest, err)

	assert.Error(t, httpErr)
	assert.Error(t, httpErr.Err)
	assert.ErrorIs(t, httpErr.Err, err)
	assert.Equal(t, err.Error(), httpErr.Message)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}
