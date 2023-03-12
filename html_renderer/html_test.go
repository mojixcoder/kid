package htmlrenderer

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/mojixcoder/kid/errors"
	"github.com/stretchr/testify/assert"
)

func newTestHTMLRenderer() *defaultHTMLRenderer {
	htmlRenderer := New("../testdata/templates/", "layouts/", ".html", false)
	return htmlRenderer
}

func getNewLineStr() string {
	if filepath.Separator != rune('/') {
		return "\r\n"
	}
	return "\n"
}

func TestNew(t *testing.T) {
	htmlRenderer := New("templates/", "layouts/", ".html", true)

	assert.Equal(t, "templates/", htmlRenderer.rootDir)
	assert.Equal(t, "layouts/", htmlRenderer.layoutDir)
	assert.Equal(t, ".html", htmlRenderer.extension)
	assert.True(t, htmlRenderer.debug)
	assert.Empty(t, htmlRenderer.funcMap)
	assert.Empty(t, htmlRenderer.templates)
}

func TestDefault(t *testing.T) {
	htmlRenderer := Default(false)

	assert.Equal(t, DefaultRootDir, htmlRenderer.rootDir)
	assert.Equal(t, DefaultLayoutsDir, htmlRenderer.layoutDir)
	assert.Equal(t, DefaultExtension, htmlRenderer.extension)
	assert.False(t, htmlRenderer.debug)
	assert.Empty(t, htmlRenderer.funcMap)
	assert.Empty(t, htmlRenderer.templates)
}

func TestDefaultHTMLRenderer_AddFunc(t *testing.T) {
	htmlRenderer := Default(false)

	assert.PanicsWithValue(t, "function cannot be nil", func() {
		htmlRenderer.AddFunc("func1", nil)
	})

	htmlRenderer.AddFunc("func1", func() {})

	assert.Equal(t, 1, len(htmlRenderer.funcMap))
}

func TestDefaultHTMLRenderer_getTemplateAndLayoutFiles(t *testing.T) {
	htmlRenderer := newTestHTMLRenderer()

	templateFiles, layoutFiles, err := htmlRenderer.getTemplateAndLayoutFiles()
	assert.NoError(t, err)
	assert.Equal(
		t,
		[]string{filepath.Join("..", "testdata", "templates", "layouts", "base.html")},
		layoutFiles,
	)
	assert.Equal(
		t,
		[]string{
			filepath.Join("..", "testdata", "templates", "index.html"),
			filepath.Join("..", "testdata", "templates", "pages", "page.html"),
			filepath.Join("..", "testdata", "templates", "pages", "page2.html"),
		},
		templateFiles,
	)

	htmlRenderer.rootDir = "invalid_path"

	templateFiles, layoutFiles, err = htmlRenderer.getTemplateAndLayoutFiles()
	assert.Error(t, err)
	assert.Nil(t, layoutFiles)
	assert.Nil(t, templateFiles)
}

func TestDefaultHTMLRenderer_loadTemplates(t *testing.T) {
	htmlRenderer := newTestHTMLRenderer()
	htmlRenderer.rootDir = "invalid_path"

	httpErr := htmlRenderer.loadTemplates().(*errors.HTTPError)

	assert.Error(t, httpErr)
	assert.Error(t, httpErr.Err)
	assert.Equal(t, httpErr.Err.Error(), httpErr.Message)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.False(t, htmlRenderer.isInitialized)

	htmlRenderer = newTestHTMLRenderer()
	htmlRenderer.AddFunc("greet", func(name string) string {
		return "Hello " + name
	})

	err := htmlRenderer.loadTemplates()

	assert.NoError(t, err)
	assert.True(t, htmlRenderer.isInitialized)
	assert.Equal(t, 3, len(htmlRenderer.templates))
	assert.IsType(t, &template.Template{}, htmlRenderer.templates["index.html"])
	assert.IsType(t, &template.Template{}, htmlRenderer.templates["pages/page.html"])
	assert.IsType(t, &template.Template{}, htmlRenderer.templates["pages/page2.html"])
}

func TestDefaultHTMLRenderer_shouldntLoadTemplates(t *testing.T) {
	htmlRenderer := newTestHTMLRenderer()

	htmlRenderer.debug = false
	htmlRenderer.isInitialized = false
	assert.False(t, htmlRenderer.shouldntLoadTemplates())

	htmlRenderer.debug = true
	htmlRenderer.isInitialized = false
	assert.False(t, htmlRenderer.shouldntLoadTemplates())

	htmlRenderer.debug = true
	htmlRenderer.isInitialized = true
	assert.False(t, htmlRenderer.shouldntLoadTemplates())

	htmlRenderer.debug = false
	htmlRenderer.isInitialized = true
	assert.True(t, htmlRenderer.shouldntLoadTemplates())
}

func TestDefaultHTMLRenderer_isLayout(t *testing.T) {
	htmlRenderer := newTestHTMLRenderer()

	assert.False(t, htmlRenderer.isLayout("../testdata/templates/index.html"))
	assert.True(t, htmlRenderer.isLayout("../testdata/templates/layouts/base.html"))
}

func TestDefaultHTMLRenderer_getTemplateName(t *testing.T) {
	htmlRenderer := newTestHTMLRenderer()

	assert.Equal(t, "index.html", htmlRenderer.getTemplateName("../testdata/templates/index.html"))
	assert.Equal(t, "pages/page.html", htmlRenderer.getTemplateName("../testdata/templates/pages/page.html"))
}

func TestDefaultHTMLRenderer_getFilesToParse(t *testing.T) {
	layouts := []string{"base.html"}
	file := "index.html"

	files := getFilesToParse(file, layouts)

	assert.Equal(t, []string{file, layouts[0]}, files)
}

func TestDefaultHTMLRenderer_RenderHTML(t *testing.T) {
	htmlRenderer := newTestHTMLRenderer()
	htmlRenderer.rootDir = "invalid_path"

	res := httptest.NewRecorder()

	err := htmlRenderer.RenderHTML(res, "index.html", nil)
	assert.Error(t, err)

	htmlRenderer = newTestHTMLRenderer()
	htmlRenderer.AddFunc("greet", func(name string) string {
		return "Hello " + name
	})

	httpErr := htmlRenderer.RenderHTML(res, "doesn't_exists.html", nil).(*errors.HTTPError)

	assert.Error(t, httpErr)
	assert.ErrorIs(t, ErrTemplateNotFound, httpErr.Err)
	assert.Equal(t, "template doesn't_exists.html not found", httpErr.Message)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)

	newline := getNewLineStr()

	res = httptest.NewRecorder()
	err = htmlRenderer.RenderHTML(res, "index.html", nil)
	assert.NoError(t, err)
	assert.Equal(
		t,
		fmt.Sprintf("%s<html><body>%s<p>content</p>%s</body></html>%s", newline, newline, newline, newline),
		res.Body.String(),
	)

	res = httptest.NewRecorder()
	err = htmlRenderer.RenderHTML(res, "pages/page.html", map[string]string{"key": "page contents"})
	assert.NoError(t, err)
	assert.Equal(t,
		fmt.Sprintf("%s<html><body>%s<p>page contents</p>%s</body></html>%s", newline, newline, newline, newline),
		res.Body.String(),
	)

	res = httptest.NewRecorder()
	err = htmlRenderer.RenderHTML(res, "pages/page2.html", nil)
	assert.NoError(t, err)
	assert.Equal(t,
		fmt.Sprintf("%s<html><body>%s<p>Hello Tom</p>%s</body></html>%s", newline, newline, newline, newline),
		res.Body.String(),
	)
}

func TestNewInternalServerHTTPError(t *testing.T) {
	err := errors.NewHTTPError(http.StatusBadRequest)
	httpErr := newInternalServerHTTPError(err, err.Error()).(*errors.HTTPError)

	assert.Error(t, httpErr)
	assert.ErrorIs(t, httpErr.Err, err)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.Equal(t, err.Error(), httpErr.Message)
}
