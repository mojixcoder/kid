// Package htmlrenderer provides an interface and its implementations for rendering HTML pages.
package htmlrenderer

import (
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

var (
	// DefaultRootDir is the default root directory.
	DefaultRootDir = filepath.FromSlash("templates/")

	// DefaultLayoutsDir is the default layout directory. Relative to root directory.
	DefaultLayoutsDir = filepath.FromSlash("layouts/")

	// DefaultExtension is the default template file extension.
	DefaultExtension = ".html"

	// ErrTemplateNotFound is the internal error when template is not found.
	ErrTemplateNotFound = errors.New("template not found")
)

// defaultHTMLRenderer is the default implementation of HTMLRenderer.
type defaultHTMLRenderer struct {
	templates     map[string]*template.Template
	funcMap       template.FuncMap
	rootDir       string
	layoutDir     string
	extension     string
	debug         bool
	isInitialized bool
}

// Verifying interface compliance.
var _ HTMLRenderer = (*defaultHTMLRenderer)(nil)

// New returns a new HTML renderer.
func New(templatesDir, layoutsDir, extension string, debug bool) *defaultHTMLRenderer {
	htmlRenderer := defaultHTMLRenderer{
		rootDir:   templatesDir,
		layoutDir: layoutsDir,
		extension: extension,
		debug:     debug,
		templates: make(map[string]*template.Template),
		funcMap:   make(template.FuncMap),
	}
	return &htmlRenderer
}

// Default returns a new default HTML renderer.
func Default(debug bool) *defaultHTMLRenderer {
	return New(
		DefaultRootDir,
		DefaultLayoutsDir,
		DefaultExtension,
		debug,
	)
}

// SetFunc sets a function in the func map.
func (r *defaultHTMLRenderer) SetFunc(name string, f any) {
	if f == nil {
		panic("function cannot be nil")
	}
	r.funcMap[name] = f
}

// RenderHTML implements Kid's HTML renderer.
func (r *defaultHTMLRenderer) RenderHTML(res http.ResponseWriter, path string, data any) {
	if err := r.loadTemplates(); err != nil {
		panic(err)
	}

	if tpl, ok := r.templates[path]; !ok {
		panic(ErrTemplateNotFound)
	} else {
		if err := tpl.Execute(res, data); err != nil {
			panic(err)
		}
	}
}

// getTemplateAndLayoutFiles returns template and layout files.
func (r *defaultHTMLRenderer) getTemplateAndLayoutFiles() ([]string, []string, error) {
	templateFiles := make([]string, 0)
	layoutFiles := make([]string, 0)

	err := filepath.Walk(r.rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !r.isValidExt(path) {
			return nil
		}

		if r.isLayout(path) {
			layoutFiles = append(layoutFiles, path)
		} else {
			templateFiles = append(templateFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return templateFiles, layoutFiles, nil
}

// loadTemplates loads and parses templates.
func (r *defaultHTMLRenderer) loadTemplates() error {
	if r.shouldntLoadTemplates() {
		return nil
	}

	templateFiles, layoutFiles, err := r.getTemplateAndLayoutFiles()
	if err != nil {
		return err
	}

	for _, templateFile := range templateFiles {
		name := r.getTemplateName(templateFile)
		files := getFilesToParse(templateFile, layoutFiles)
		tpl := template.Must(template.New(filepath.Base(name)).Funcs(r.funcMap).ParseFiles(files...))
		r.templates[name] = tpl
	}

	r.isInitialized = true

	return nil
}

// shouldntLoadTemplates if true don't need to load templates otherwise load templates.
func (r *defaultHTMLRenderer) shouldntLoadTemplates() bool {
	return r.isInitialized && !r.debug
}

// isLayout determines if the file is a layout file or not.
func (r *defaultHTMLRenderer) isLayout(file string) bool {
	return strings.HasPrefix(filepath.ToSlash(file), filepath.ToSlash(r.rootDir+r.layoutDir))
}

// isValidExt determines if the file has valid extension.
func (r *defaultHTMLRenderer) isValidExt(file string) bool {
	return strings.HasSuffix(file, r.extension)
}

// getTemplateName extracts template name from file path.
func (r *defaultHTMLRenderer) getTemplateName(filePath string) string {
	return filepath.ToSlash(filePath[len(r.rootDir):])
}

// getFilesToParse merges template path and layouts into a string slice.
func getFilesToParse(templatePath string, layouts []string) []string {
	files := make([]string, 0, len(layouts)+1)
	files = append(files, templatePath)
	files = append(files, layouts...)
	return files
}
