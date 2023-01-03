package serializer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mojixcoder/kid/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewXMLSerializer(t *testing.T) {
	serializer := NewXMLSerializer()

	assert.NotNil(t, serializer)
	assert.IsType(t, defaultXMLSerializer{}, serializer)
}

func TestDefaultXMLSerializerWrite(t *testing.T) {
	res := httptest.NewRecorder()

	serializer := defaultXMLSerializer{}

	p := person{Name: "Mojix", Age: 22}

	err := serializer.Write(res, p, "")

	assert.NoError(t, err)
	assert.Equal(t, "<person><name>Mojix</name><age>22</age></person>", res.Body.String())

	res = httptest.NewRecorder()

	err = serializer.Write(res, p, "    ")

	assert.NoError(t, err)
	assert.Equal(t, "<person>\n    <name>Mojix</name>\n    <age>22</age>\n</person>", res.Body.String())

	// Unsupported type.
	err = serializer.Write(res, make(chan bool), "")
	assert.Error(t, err)

	httpErr := err.(*errors.HTTPError)

	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.Equal(t, httpErr.Err.Error(), httpErr.Message)
	assert.Error(t, httpErr.Err)
}

func TestDefaultXMLSerializerRead(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("<person><name>Mojix</name><age>22</age></person>"))

	serializer := defaultXMLSerializer{}

	var p person
	err := serializer.Read(req, &p)

	assert.NoError(t, err)
	assert.Equal(t, "Mojix", p.Name)
	assert.Equal(t, 22, p.Age)

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader("<person><name>Mojix</name><age>22</age></person>"))

	// Invalid argument passed to unmarshal.
	var p2 person
	err = serializer.Read(req, p2)
	assert.Error(t, err)

	httpErr := err.(*errors.HTTPError)

	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.Equal(t, httpErr.Err.Error(), httpErr.Message)
	assert.Error(t, httpErr.Err)

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader("<person><name>Mojix</name><age>22.5</age></person>"))

	err = serializer.Read(req, &p2)
	assert.Error(t, err)

	httpErr = err.(*errors.HTTPError)

	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Equal(t, httpErr.Err.Error(), httpErr.Message)
	assert.Error(t, httpErr.Err)
}
