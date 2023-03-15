package kid

import (
	"net/http"
	"os"
)

// FS is Kid's http.FileSystem implementation to disable directory listing.
type FS struct {
	http.FileSystem
}

// File is Kid's http.File implementation to disable directory listing.
type File struct {
	http.File
}

// Readdir overrides http.File's default implementation.
//
// It disables directory listing.
func (f File) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

// Open overrides http.FileSystem's default implementation to disable directory listing.
func (fs FS) Open(name string) (http.File, error) {
	f, err := fs.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}

	return File{f}, err
}

// newFileServer returns new file server.
func newFileServer(urlPath string, fs http.FileSystem) http.Handler {
	panicIfNil(fs, "file system cannot be nil")

	urlPath = cleanPath(urlPath, false)
	urlPath = appendSlash(urlPath)

	fileServer := http.StripPrefix(urlPath, http.FileServer(fs))

	return fileServer
}

// appendSlash appends slash to a path if the path is not end with slash.
func appendSlash(path string) string {
	if path[len(path)-1] != '/' {
		path = path + "/"
	}
	return path
}
