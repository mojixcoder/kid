package kid

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestDefaultJSONSerializerWrite(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := newContext(New())
	c.reset(req, res)

	serializer := defaultJSONSerializer{}

	p := person{Name: "Mojix", Age: 22}

	err := serializer.Write(c, p, "")
	assert.NoError(t, err)

	assert.Equal(t, "{\"name\":\"Mojix\",\"age\":22}\n", res.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	res = httptest.NewRecorder()
	c.reset(req, res)

	err = serializer.Write(c, p, "    ")
	assert.NoError(t, err)

	assert.Equal(t, "{\n    \"name\": \"Mojix\",\n    \"age\": 22\n}\n", res.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	res = httptest.NewRecorder()
	c.reset(req, res)

	// Channel type cannot be converted to JSON.
	err = serializer.Write(c, make(chan bool), "")
	assert.Error(t, err)

	httpErr := err.(*HTTPError)

	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.Equal(t, httpErr.Err.Error(), httpErr.Message)
	assert.Error(t, httpErr.Err)
}

func TestDefaultJSONSerializerRead(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("{\"name\":\"Mojix\",\"age\":22}"))
	res := httptest.NewRecorder()
	c := newContext(New())
	c.reset(req, res)

	serializer := defaultJSONSerializer{}

	var p person
	err := serializer.Read(c, &p)

	assert.NoError(t, err)
	assert.Equal(t, "Mojix", p.Name)
	assert.Equal(t, 22, p.Age)

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader("{\"name\":\"Mojix\",\"age\":22}"))
	res = httptest.NewRecorder()
	c.reset(req, res)

	// Invalid argument passed to unmarshal.
	var p2 person
	err = serializer.Read(c, p2)
	assert.Error(t, err)

	httpErr := err.(*HTTPError)
	_, ok := httpErr.Err.(*json.InvalidUnmarshalError)

	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.Equal(t, httpErr.Err.Error(), httpErr.Message)
	assert.Error(t, httpErr.Err)
	assert.True(t, ok)

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader("{\"name\":\"Mojix\",\"age\":22.5}"))
	res = httptest.NewRecorder()
	c.reset(req, res)

	err = serializer.Read(c, &p2)

	assert.Error(t, err)
	assert.Error(t, err.(*HTTPError).Err)
	assert.Equal(t, http.StatusBadRequest, err.(*HTTPError).Code)
}
