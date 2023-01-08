package kid

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/mojixcoder/kid/errors"
	htmlrenderer "github.com/mojixcoder/kid/html_renderer"
	"github.com/stretchr/testify/assert"
)

type person struct {
	Name string `json:"name" xml:"name"`
	Age  int    `json:"age" xml:"age"`
}

func TestNewContext(t *testing.T) {
	k := New()

	ctx := newContext(k)

	assert.Equal(t, k, ctx.kid)
	assert.Nil(t, ctx.storage)
	assert.Nil(t, ctx.params)
	assert.Nil(t, ctx.request)
	assert.Nil(t, ctx.response)
}

func TestContextReset(t *testing.T) {
	ctx := newContext(New())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	expectedRes := newResponse(res)

	ctx.reset(req, res)

	assert.Equal(t, req, ctx.request)
	assert.Equal(t, expectedRes, ctx.response)
	assert.Equal(t, make(Map), ctx.storage)
	assert.Equal(t, make(Params), ctx.params)
}

func TestContextRequest(t *testing.T) {
	ctx := newContext(New())

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	ctx.reset(req, nil)

	assert.Equal(t, req, ctx.Request())
}

func TestContextResponse(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()
	expectedRes := newResponse(res)

	ctx.reset(nil, res)

	assert.Equal(t, expectedRes, ctx.Response())
}

func TestContextSetParams(t *testing.T) {
	ctx := newContext(New())

	params := Params{"foo": "bar", "abc": "xyz"}

	ctx.setParams(params)

	assert.Equal(t, params, ctx.params)
}

func TestContextParams(t *testing.T) {
	ctx := newContext(New())

	params := Params{"foo": "bar", "abc": "xyz"}

	ctx.setParams(params)

	assert.Equal(t, params, ctx.Params())
}

func TestContextParam(t *testing.T) {
	ctx := newContext(New())

	params := Params{"foo": "bar", "abc": "xyz"}

	ctx.setParams(params)

	assert.Equal(t, params["foo"], ctx.Param("foo"))
	assert.Equal(t, params["abc"], ctx.Param("abc"))
}

func TestContextQueryParams(t *testing.T) {
	ctx := newContext(New())

	req := httptest.NewRequest(http.MethodGet, "/?foo=bar&abc=xyz&abc=2", nil)

	ctx.reset(req, nil)

	assert.Equal(t, url.Values{"foo": []string{"bar"}, "abc": []string{"xyz", "2"}}, ctx.QueryParams())

	req = httptest.NewRequest(http.MethodGet, "/", nil)

	ctx.reset(req, nil)

	assert.Equal(t, url.Values{}, ctx.QueryParams())
}

func TestContextQueryParam(t *testing.T) {
	ctx := newContext(New())

	req := httptest.NewRequest(http.MethodGet, "/?foo=bar&abc=xyz&abc=2", nil)

	ctx.reset(req, nil)

	assert.Equal(t, "bar", ctx.QueryParam("foo"))
	assert.Equal(t, "xyz", ctx.QueryParam("abc"))
	assert.Equal(t, "", ctx.QueryParam("does_not_exist"))
}

func TestContextQueryParamMultiple(t *testing.T) {
	ctx := newContext(New())

	req := httptest.NewRequest(http.MethodGet, "/?foo=bar&abc=xyz&abc=2", nil)

	ctx.reset(req, nil)

	assert.Equal(t, []string{"bar"}, ctx.QueryParamMultiple("foo"))
	assert.Equal(t, []string{"xyz", "2"}, ctx.QueryParamMultiple("abc"))
	assert.Equal(t, []string{}, ctx.QueryParamMultiple("does_not_exist"))
}

func TestContextSet(t *testing.T) {
	ctx := newContext(New())
	ctx.reset(nil, nil)

	ctx.Set("val", 1)

	val, ok := ctx.storage["val"]

	assert.True(t, ok)
	assert.Equal(t, 1, val)
	assert.Equal(t, 1, len(ctx.storage))
}

func TestContextGet(t *testing.T) {
	ctx := newContext(New())
	ctx.reset(nil, nil)

	ctx.storage["val"] = 12.64

	val, ok := ctx.Get("val")

	assert.True(t, ok)
	assert.Equal(t, 12.64, val)
}

func TestContextGetSetDataRace(t *testing.T) {
	ctx := newContext(New())
	ctx.reset(nil, nil)

	ch := make(chan struct{})

	go func() {
		ctx.Set("foo", "bar")
		close(ch)
	}()

	_, _ = ctx.Get("foo")

	<-ch
}

func TestContextWriteContentType(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()

	ctx.reset(nil, res)

	assert.Equal(t, "", ctx.response.Header().Get("Content-Type"))

	ctx.writeContentType("application/json")

	assert.Equal(t, "application/json", ctx.response.Header().Get("Content-Type"))

	// don't write content type if it's already written.
	ctx.writeContentType("application/javascript")

	assert.Equal(t, "application/json", ctx.response.Header().Get("Content-Type"))
}

func TestNoContent(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()

	ctx.reset(nil, res)

	ctx.NoContent(http.StatusNoContent)
	assert.Equal(t, http.StatusNoContent, res.Code)
}

func TestContextReadJSON(t *testing.T) {
	ctx := newContext(New())

	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("{\"name\":\"Mojix\",\"age\":22}"))
	res := httptest.NewRecorder()

	ctx.reset(req, res)

	var p person
	err := ctx.ReadJSON(&p)
	assert.NoError(t, err)

	assert.Equal(t, person{Name: "Mojix", Age: 22}, p)

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader("{\"name\":\"Mojix\",\"age\":22"))
	res = httptest.NewRecorder()

	ctx.reset(req, res)

	var p2 person
	httpErr := ctx.ReadJSON(&p2).(*errors.HTTPError)

	assert.Error(t, httpErr)
	assert.Error(t, httpErr.Err)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestContextJSON(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()

	ctx.reset(nil, res)

	p := person{Name: "foo", Age: 1999}
	err := ctx.JSON(http.StatusCreated, &p)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"name\":\"foo\",\"age\":1999}\n", res.Body.String())

	res = httptest.NewRecorder()

	ctx.reset(nil, res)

	httpErr := ctx.JSON(http.StatusCreated, make(chan bool)).(*errors.HTTPError)

	assert.Error(t, httpErr)
	assert.Error(t, httpErr.Err)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
}

func TestContextJSONIndent(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()

	ctx.reset(nil, res)

	p := person{Name: "foo", Age: 1999}
	err := ctx.JSONIndent(http.StatusCreated, &p, "    ")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\n    \"name\": \"foo\",\n    \"age\": 1999\n}\n", res.Body.String())

	res = httptest.NewRecorder()

	ctx.reset(nil, res)

	httpErr := ctx.JSONIndent(http.StatusCreated, make(chan bool), "    ").(*errors.HTTPError)

	assert.Error(t, httpErr)
	assert.Error(t, httpErr.Err)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
}

func TestContextJSONByte(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()

	ctx.reset(nil, res)

	p := person{Name: "foo", Age: 1999}

	blob, err := json.Marshal(p)
	assert.NoError(t, err)

	err = ctx.JSONByte(http.StatusOK, blob)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"name\":\"foo\",\"age\":1999}", res.Body.String())
}

func TestContextReadXML(t *testing.T) {
	ctx := newContext(New())

	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("<person><name>Mojix</name><age>22</age></person>"))

	ctx.reset(req, nil)

	var p person
	err := ctx.ReadXML(&p)
	assert.NoError(t, err)

	assert.Equal(t, person{Name: "Mojix", Age: 22}, p)

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader("<person><name>Mojix</name><age>22</age></person"))

	ctx.reset(req, nil)

	var p2 person
	httpErr := ctx.ReadXML(&p2).(*errors.HTTPError)

	assert.Error(t, httpErr)
	assert.Error(t, httpErr.Err)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestContextXML(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()

	ctx.reset(nil, res)

	p := person{Name: "foo", Age: 1999}
	err := ctx.XML(http.StatusCreated, &p)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/xml", res.Header().Get("Content-Type"))
	assert.Equal(t, "<person><name>foo</name><age>1999</age></person>", res.Body.String())

	res = httptest.NewRecorder()

	ctx.reset(nil, res)

	httpErr := ctx.XML(http.StatusCreated, make(chan bool)).(*errors.HTTPError)

	assert.Error(t, httpErr)
	assert.Error(t, httpErr.Err)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
}

func TestContextXMLIndent(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()

	ctx.reset(nil, res)

	p := person{Name: "foo", Age: 1999}
	err := ctx.XMLIndent(http.StatusCreated, &p, "    ")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/xml", res.Header().Get("Content-Type"))
	assert.Equal(t, "<person>\n    <name>foo</name>\n    <age>1999</age>\n</person>", res.Body.String())

	res = httptest.NewRecorder()

	ctx.reset(nil, res)

	httpErr := ctx.XMLIndent(http.StatusCreated, make(chan bool), "    ").(*errors.HTTPError)

	assert.Error(t, httpErr)
	assert.Error(t, httpErr.Err)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
}

func TestContextXMLByte(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()

	ctx.reset(nil, res)

	p := person{Name: "foo", Age: 1999}

	blob, err := xml.Marshal(p)
	assert.NoError(t, err)

	err = ctx.XMLByte(http.StatusOK, blob)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/xml", res.Header().Get("Content-Type"))
	assert.Equal(t, "<person><name>foo</name><age>1999</age></person>", res.Body.String())
}

func TestContextHTML(t *testing.T) {
	k := New()
	renderer := htmlrenderer.New("testdata/templates/", "layouts/", ".html", false)
	renderer.AddFunc("greet", func() int { return 1 })
	k.htmlRenderer = renderer

	ctx := newContext(k)

	res := httptest.NewRecorder()
	ctx.reset(nil, res)

	err := ctx.HTML(http.StatusAccepted, "index.html", nil)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, res.Code)
	assert.Equal(t, "\n<html><body>\n<p>content</p>\n</body></html>\n", res.Body.String())
	assert.Equal(t, "text/html", res.Header().Get("Content-Type"))
}

func TestContextHTMLString(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()
	ctx.reset(nil, res)

	err := ctx.HTMLString(http.StatusAccepted, "<p>Hello</p>")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, res.Code)
	assert.Equal(t, "<p>Hello</p>", res.Body.String())
	assert.Equal(t, "text/html", res.Header().Get("Content-Type"))

}
