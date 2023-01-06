package htmlrenderer

import "net/http"

// HTMLRenderer is the interface for rendering
type HTMLRenderer interface {
	// RenderHTML renders html template
	RenderHTML(res http.ResponseWriter, path string, data any) error
}
