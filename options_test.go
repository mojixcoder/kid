package kid

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockEverything struct{}

func (*mockEverything) RenderHTML(res http.ResponseWriter, path string, data any) error {
	return nil
}

func (*mockEverything) Write(w http.ResponseWriter, in any, indent string) {}

func (*mockEverything) Read(req *http.Request, out any) error {
	return nil
}

func TestWithDebug(t *testing.T) {
	k := New()

	opt := WithDebug(true)
	opt.apply(k)

	assert.True(t, k.Debug())
}

func TestWithHTMLRenderer(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "renderer cannot be nil", func() {
		WithHTMLRenderer(nil)
	})

	renderer := &mockEverything{}

	opt := WithHTMLRenderer(renderer)
	opt.apply(k)

	assert.Equal(t, renderer, k.htmlRenderer)
}

func TestWithXMLSerializer(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "xml serializer cannot be nil", func() {
		WithXMLSerializer(nil)
	})

	serializer := &mockEverything{}

	opt := WithXMLSerializer(serializer)
	opt.apply(k)

	assert.Equal(t, serializer, k.xmlSerializer)
}

func TestWithJSONSerializer(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "json serializer cannot be nil", func() {
		WithJSONSerializer(nil)
	})

	serializer := &mockEverything{}

	opt := WithJSONSerializer(serializer)
	opt.apply(k)

	assert.Equal(t, serializer, k.jsonSerializer)
}

func TestWithNotFoundHandler(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "not found handler cannot be nil", func() {
		WithNotFoundHandler(nil)
	})

	hanlder := func(c *Context) {}

	opt := WithNotFoundHandler(hanlder)
	opt.apply(k)

	assert.True(t, funcsAreEqual(hanlder, k.notFoundHandler))
}

func TestWithMethodNotAllowedHandler(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "method not allowed handler cannot be nil", func() {
		WithMethodNotAllowedHandler(nil)
	})

	hanlder := func(c *Context) {}

	opt := WithMethodNotAllowedHandler(hanlder)
	opt.apply(k)

	assert.True(t, funcsAreEqual(hanlder, k.methodNotAllowedHandler))
}
