package serializer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewXMLSerializer(t *testing.T) {
	serializer := NewXMLSerializer()

	assert.NotNil(t, serializer)
	assert.IsType(t, defaultXMLSerializer{}, serializer)
}

func TestDefaultXMLSerializer_Write(t *testing.T) {
	res := httptest.NewRecorder()

	serializer := defaultXMLSerializer{}

	p := person{Name: "Mojix", Age: 22}

	serializer.Write(res, p, "")
	assert.Equal(t, "<person><name>Mojix</name><age>22</age></person>", res.Body.String())

	res = httptest.NewRecorder()

	serializer.Write(res, p, "    ")
	assert.Equal(t, "<person>\n    <name>Mojix</name>\n    <age>22</age>\n</person>", res.Body.String())

	// Unsupported type.
	assert.Panics(t, func() {
		serializer.Write(res, make(chan bool), "")
	})
}

func TestDefaultXMLSerializer_Read(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("<person><name>Mojix</name><age>22</age></person>"))

	serializer := defaultXMLSerializer{}

	var p person
	err := serializer.Read(req, &p)

	assert.NoError(t, err)
	assert.Equal(t, "Mojix", p.Name)
	assert.Equal(t, 22, p.Age)

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader("<person><name>Mojix</name><age>22</age></person>"))

	var p2 person

	// Invalid argument passed to unmarshal.
	assert.Panics(t, func() {
		serializer.Read(req, p2)
	})

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader("<person><name>Mojix</name><age>22.5</age></person>"))

	err = serializer.Read(req, &p2)
	assert.Error(t, err)
}
