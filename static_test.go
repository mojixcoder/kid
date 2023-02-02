package kid

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppendSlash(t *testing.T) {
	path := "/static"
	assert.Equal(t, "/static/", appendSlash(path))

	path = path + "/"
	assert.Equal(t, "/static/", appendSlash(path))
}

func TestNewFileServer(t *testing.T) {
	assert.Panics(t, func() {
		newFileServer("/static/", nil)
	})

	fileServer := newFileServer("/static/", http.Dir("/var/www"))
	assert.NotNil(t, fileServer)
}

func TestFileReaddir(t *testing.T) {
	var f File

	info, err := f.Readdir(10)

	assert.Nil(t, info)
	assert.Nil(t, err)
}

func TestFSOpen(t *testing.T) {
	fs := FS{http.Dir("testdata/static")}

	f, err := fs.Open("non-existent.js")
	assert.Error(t, err)
	assert.Nil(t, f)

	f, err = fs.Open("main.html")
	assert.NoError(t, err)
	assert.IsType(t, File{}, f)
}
