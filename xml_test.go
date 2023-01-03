package kid

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultXMLSerializerWrite(t *testing.T) {
	res := httptest.NewRecorder()
	c := newContext(New())
	c.reset(nil, res)

	serializer := defaultXMLSerializer{}

	p := person{Name: "Mojix", Age: 22}

	err := serializer.Write(c, p, "")

	assert.NoError(t, err)
	assert.Equal(t, "<person><name>Mojix</name><age>22</age></person>", res.Body.String())

	res = httptest.NewRecorder()
	c.reset(nil, res)

	err = serializer.Write(c, p, "    ")

	assert.NoError(t, err)
	assert.Equal(t, "<person>\n    <name>Mojix</name>\n    <age>22</age>\n</person>", res.Body.String())

	// Unsupported type.
	err = serializer.Write(c, make(chan bool), "")
	assert.Error(t, err)

	httpErr := err.(*HTTPError)

	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.Equal(t, httpErr.Err.Error(), httpErr.Message)
	assert.Error(t, httpErr.Err)
}

func TestDefaultXMLSerializerRead(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("<person><name>Mojix</name><age>22</age></person>"))
	c := newContext(New())
	c.reset(req, nil)

	serializer := defaultXMLSerializer{}

	var p person
	err := serializer.Read(c, &p)

	assert.NoError(t, err)
	assert.Equal(t, "Mojix", p.Name)
	assert.Equal(t, 22, p.Age)

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader("<person><name>Mojix</name><age>22</age></person>"))
	c.reset(req, nil)

	// Invalid argument passed to unmarshal.
	var p2 person
	err = serializer.Read(c, p2)
	assert.Error(t, err)

	httpErr := err.(*HTTPError)

	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.Equal(t, httpErr.Err.Error(), httpErr.Message)
	assert.Error(t, httpErr.Err)

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader("<person><name>Mojix</name><age>22.5</age></person>"))
	c.reset(req, nil)

	err = serializer.Read(c, &p2)
	assert.Error(t, err)

	httpErr = err.(*HTTPError)

	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Equal(t, httpErr.Err.Error(), httpErr.Message)
	assert.Error(t, httpErr.Err)
}
