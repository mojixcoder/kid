package kid

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	htmlrenderer "github.com/mojixcoder/kid/html_renderer"
	"github.com/stretchr/testify/assert"
)

type errWriter struct {
	*httptest.ResponseRecorder
}

func (errWriter) Write(blob []byte) (int, error) {
	return 0, errors.New("new err")
}

type person struct {
	Name string `json:"name" xml:"name"`
	Age  int    `json:"age" xml:"age"`
}

func getNewLineStr() string {
	if filepath.Separator != rune('/') {
		return "\r\n"
	}
	return "\n"
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

func TestContext_reset(t *testing.T) {
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

func TestContext_Request(t *testing.T) {
	ctx := newContext(New())

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	ctx.reset(req, nil)

	assert.Equal(t, req, ctx.Request())
}

func TestContext_Response(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()
	expectedRes := newResponse(res)

	ctx.reset(nil, res)

	assert.Equal(t, expectedRes, ctx.Response())
}

func TestContext_setParams(t *testing.T) {
	ctx := newContext(New())

	params := Params{"foo": "bar", "abc": "xyz"}

	ctx.setParams(params)

	assert.Equal(t, params, ctx.params)
}

func TestContext_Params(t *testing.T) {
	ctx := newContext(New())

	params := Params{"foo": "bar", "abc": "xyz"}

	ctx.setParams(params)

	assert.Equal(t, params, ctx.Params())
}

func TestContext_Param(t *testing.T) {
	ctx := newContext(New())

	params := Params{"foo": "bar", "abc": "xyz"}

	ctx.setParams(params)

	assert.Equal(t, params["foo"], ctx.Param("foo"))
	assert.Equal(t, params["abc"], ctx.Param("abc"))
}

func TestContext_QueryParams(t *testing.T) {
	ctx := newContext(New())

	req := httptest.NewRequest(http.MethodGet, "/?foo=bar&abc=xyz&abc=2", nil)

	ctx.reset(req, nil)

	assert.Equal(t, url.Values{"foo": []string{"bar"}, "abc": []string{"xyz", "2"}}, ctx.QueryParams())

	req = httptest.NewRequest(http.MethodGet, "/", nil)

	ctx.reset(req, nil)

	assert.Equal(t, url.Values{}, ctx.QueryParams())
}

func TestContext_QueryParam(t *testing.T) {
	ctx := newContext(New())

	req := httptest.NewRequest(http.MethodGet, "/?foo=bar&abc=xyz&abc=2", nil)

	ctx.reset(req, nil)

	assert.Equal(t, "bar", ctx.QueryParam("foo"))
	assert.Equal(t, "xyz", ctx.QueryParam("abc"))
	assert.Equal(t, "", ctx.QueryParam("does_not_exist"))
}

func TestContext_QueryParamMultiple(t *testing.T) {
	ctx := newContext(New())

	req := httptest.NewRequest(http.MethodGet, "/?foo=bar&abc=xyz&abc=2", nil)

	ctx.reset(req, nil)

	assert.Equal(t, []string{"bar"}, ctx.QueryParamMultiple("foo"))
	assert.Equal(t, []string{"xyz", "2"}, ctx.QueryParamMultiple("abc"))
	assert.Equal(t, []string{}, ctx.QueryParamMultiple("does_not_exist"))
}

func TestContext_Set(t *testing.T) {
	ctx := newContext(New())
	ctx.reset(nil, nil)

	ctx.Set("val", 1)

	val, ok := ctx.storage["val"]

	assert.True(t, ok)
	assert.Equal(t, 1, val)
	assert.Equal(t, 1, len(ctx.storage))
}

func TestContext_Get(t *testing.T) {
	ctx := newContext(New())
	ctx.reset(nil, nil)

	ctx.storage["val"] = 12.64

	val, ok := ctx.Get("val")

	assert.True(t, ok)
	assert.Equal(t, 12.64, val)
}

func TestContext_GetSet_DataRace(t *testing.T) {
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

func TestContext_writeContentType(t *testing.T) {
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

func Test_NoContent(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()

	ctx.reset(nil, res)

	ctx.NoContent(http.StatusNoContent)
	assert.Equal(t, http.StatusNoContent, res.Code)
}

func TestContext_ReadJSON(t *testing.T) {
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
	err = ctx.ReadJSON(&p2)

	assert.Error(t, err)
}

func TestContext_mustWrite(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()
	ctx.reset(nil, res)

	ctx.mustWrite([]byte("byte"))

	assert.Equal(t, "byte", res.Body.String())

	ctx.reset(nil, errWriter{res})

	assert.Panics(t, func() {
		ctx.mustWrite([]byte("byte"))
	})
}

func TestContext_JSON(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()

	ctx.reset(nil, res)

	p := person{Name: "foo", Age: 1999}
	ctx.JSON(http.StatusCreated, &p)

	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"name\":\"foo\",\"age\":1999}\n", res.Body.String())

	res = httptest.NewRecorder()

	ctx.reset(nil, res)

	assert.Panics(t, func() {
		ctx.JSON(http.StatusCreated, make(chan bool))
	})
}

func TestContext_JSONIndent(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()

	ctx.reset(nil, res)

	p := person{Name: "foo", Age: 1999}
	ctx.JSONIndent(http.StatusCreated, &p, "    ")

	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\n    \"name\": \"foo\",\n    \"age\": 1999\n}\n", res.Body.String())

	res = httptest.NewRecorder()

	ctx.reset(nil, res)

	assert.Panics(t, func() {
		ctx.JSONIndent(http.StatusCreated, make(chan bool), "    ")
	})
}

func TestContext_JSONByte(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()

	ctx.reset(nil, res)

	p := person{Name: "foo", Age: 1999}

	blob, err := json.Marshal(p)
	assert.NoError(t, err)

	ctx.JSONByte(http.StatusOK, blob)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"name\":\"foo\",\"age\":1999}", res.Body.String())
}

func TestContext_ReadXML(t *testing.T) {
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
	err = ctx.ReadXML(&p2)

	assert.Error(t, err)
}

func TestContext_XML(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()

	ctx.reset(nil, res)

	p := person{Name: "foo", Age: 1999}
	ctx.XML(http.StatusCreated, &p)

	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/xml", res.Header().Get("Content-Type"))
	assert.Equal(t, "<person><name>foo</name><age>1999</age></person>", res.Body.String())

	res = httptest.NewRecorder()

	ctx.reset(nil, res)

	assert.Panics(t, func() {
		ctx.XML(http.StatusCreated, make(chan bool))
	})
}

func TestContext_XMLIndent(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()

	ctx.reset(nil, res)

	p := person{Name: "foo", Age: 1999}
	ctx.XMLIndent(http.StatusCreated, &p, "    ")

	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/xml", res.Header().Get("Content-Type"))
	assert.Equal(t, "<person>\n    <name>foo</name>\n    <age>1999</age>\n</person>", res.Body.String())

	res = httptest.NewRecorder()

	ctx.reset(nil, res)

	assert.Panics(t, func() {
		ctx.XMLIndent(http.StatusCreated, make(chan bool), "    ")
	})
}

func TestContext_XMLByte(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()

	ctx.reset(nil, res)

	p := person{Name: "foo", Age: 1999}

	blob, err := xml.Marshal(p)
	assert.NoError(t, err)

	ctx.XMLByte(http.StatusOK, blob)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/xml", res.Header().Get("Content-Type"))
	assert.Equal(t, "<person><name>foo</name><age>1999</age></person>", res.Body.String())
}

func TestContext_HTML(t *testing.T) {
	k := New()
	renderer := htmlrenderer.New("testdata/templates/", "layouts/", ".html", false)
	renderer.SetFunc("greet", func() int { return 1 })
	k.htmlRenderer = renderer

	ctx := newContext(k)

	res := httptest.NewRecorder()
	ctx.reset(nil, res)

	ctx.HTML(http.StatusAccepted, "index.html", nil)

	newLine := getNewLineStr()
	expectedRes := fmt.Sprintf(
		"%s<html><body>%s<p>content</p>%s</body></html>%s",
		newLine, newLine, newLine, newLine,
	)

	assert.Equal(t, http.StatusAccepted, res.Code)
	assert.Equal(t, expectedRes, res.Body.String())
	assert.Equal(t, "text/html", res.Header().Get("Content-Type"))
}

func TestContext_HTMLString(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()
	ctx.reset(nil, res)

	ctx.HTMLString(http.StatusAccepted, "<p>Hello</p>")

	assert.Equal(t, http.StatusAccepted, res.Code)
	assert.Equal(t, "<p>Hello</p>", res.Body.String())
	assert.Equal(t, "text/html", res.Header().Get("Content-Type"))
}

func TestContext_Debug(t *testing.T) {
	k := New()
	k.debug = true

	ctx := newContext(k)

	assert.True(t, ctx.Debug())
}

func TestContext_String(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()
	ctx.reset(nil, res)

	ctx.String(http.StatusAccepted, "hello")

	assert.Equal(t, http.StatusAccepted, res.Code)
	assert.Equal(t, "hello", res.Body.String())
	assert.Equal(t, "text/plain", res.Header().Get("Content-Type"))
}

func TestContext_Byte(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()
	ctx.reset(nil, res)

	ctx.Byte(http.StatusOK, []byte("mojix"))

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "mojix", res.Body.String())
	assert.Equal(t, "application/octet-stream", res.Header().Get("Content-Type"))
}

func TestContext_GetResponseHeader(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()
	res.Header().Set("key", "value")

	ctx.reset(nil, res)

	assert.Equal(t, "value", ctx.GetResponseHeader("key"))
}

func TestContext_SetResponseHeader(t *testing.T) {
	ctx := newContext(New())

	res := httptest.NewRecorder()
	ctx.reset(nil, res)

	ctx.SetResponseHeader("key", "value")
	assert.Equal(t, "value", ctx.GetResponseHeader("key"))
}

func TestContext_GetRequestHeader(t *testing.T) {
	ctx := newContext(New())

	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	req.Header.Set("key", "value")

	ctx.reset(req, nil)

	assert.Equal(t, "value", ctx.GetRequestHeader("key"))
}

func TestContext_SetRequestHeader(t *testing.T) {
	ctx := newContext(New())

	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	ctx.reset(req, nil)

	ctx.SetRequestHeader("key", "value")
	assert.Equal(t, "value", ctx.GetRequestHeader("key"))
}
