package htmlrenderer

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mojixcoder/kid"
	"github.com/mojixcoder/kid/errors"
	"github.com/stretchr/testify/assert"
)

func newTestHTMLRenderer() defaultHTMLRenderer {
	htmlRenderer := New("../testdata/templates/", "layouts/", ".html", false)
	return htmlRenderer
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

func TestDefaultHTMLRendererAddFunc(t *testing.T) {
	htmlRenderer := Default(false)

	assert.PanicsWithValue(t, "function cannot be nil", func() {
		htmlRenderer.AddFunc("func1", nil)
	})

	htmlRenderer.AddFunc("func1", func() {})

	assert.Equal(t, 1, len(htmlRenderer.funcMap))
}

func TestDefaultHTMLRendererGetTemplateAndLayoutFiles(t *testing.T) {
	htmlRenderer := newTestHTMLRenderer()

	templateFiles, layoutFiles, err := htmlRenderer.getTemplateAndLayoutFiles()
	assert.NoError(t, err)
	assert.Equal(
		t,
		[]string{"../testdata/templates/layouts/base.html"},
		layoutFiles,
	)
	assert.Equal(
		t,
		[]string{"../testdata/templates/index.html", "../testdata/templates/pages/page.html", "../testdata/templates/pages/page2.html"},
		templateFiles,
	)

	htmlRenderer.rootDir = "invalid_path"

	templateFiles, layoutFiles, err = htmlRenderer.getTemplateAndLayoutFiles()
	assert.Error(t, err)
	assert.Nil(t, layoutFiles)
	assert.Nil(t, templateFiles)
}

func TestDefaultHTMLRendererLoadTemplates(t *testing.T) {
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

func TestDefaultHTMLRendererShouldntLoadTemplates(t *testing.T) {
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

func TestDefaultHTMLRendererIsLayout(t *testing.T) {
	htmlRenderer := newTestHTMLRenderer()

	assert.False(t, htmlRenderer.isLayout("../testdata/templates/index.html"))
	assert.True(t, htmlRenderer.isLayout("../testdata/templates/layouts/base.html"))
}

func TestDefaultHTMLRendererGetTemplateName(t *testing.T) {
	htmlRenderer := newTestHTMLRenderer()

	assert.Equal(t, "index.html", htmlRenderer.getTemplateName("../testdata/templates/index.html"))
	assert.Equal(t, "pages/page.html", htmlRenderer.getTemplateName("../testdata/templates/pages/page.html"))
}

func TestDefaultHTMLRendererGetFilesToParse(t *testing.T) {
	layouts := []string{"base.html"}
	file := "index.html"

	files := getFilesToParse(file, layouts)

	assert.Equal(t, []string{file, layouts[0]}, files)
}

func TestDefaultHTMLRendererRenderHTML(t *testing.T) {
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

	res = httptest.NewRecorder()
	err = htmlRenderer.RenderHTML(res, "index.html", nil)
	assert.NoError(t, err)
	assert.Equal(t, "\n<html><body>\n<p>content</p>\n</body></html>\n", res.Body.String())

	res = httptest.NewRecorder()
	err = htmlRenderer.RenderHTML(res, "pages/page.html", kid.Map{"key": "page contents"})
	assert.NoError(t, err)
	assert.Equal(t, "\n<html><body>\n<p>page contents</p>\n</body></html>\n", res.Body.String())

	res = httptest.NewRecorder()
	err = htmlRenderer.RenderHTML(res, "pages/page2.html", nil)
	assert.NoError(t, err)
	assert.Equal(t, "\n<html><body>\n<p>Hello Tom</p>\n</body></html>\n", res.Body.String())
}

func TestNewInternalServerHTTPError(t *testing.T) {
	err := errors.NewHTTPError(http.StatusBadRequest)
	httpErr := newInternalServerHTTPError(err, err.Error()).(*errors.HTTPError)

	assert.Error(t, httpErr)
	assert.ErrorIs(t, httpErr.Err, err)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.Equal(t, err.Error(), httpErr.Message)
}
