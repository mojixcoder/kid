package serializer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mojixcoder/kid/errors"
	"github.com/stretchr/testify/assert"
)

type person struct {
	Name string `json:"name" xml:"name"`
	Age  int    `json:"age" xml:"age"`
}

func TestNewJSONSerializer(t *testing.T) {
	serializer := NewJSONSerializer()

	assert.NotNil(t, serializer)
	assert.IsType(t, defaultJSONSerializer{}, serializer)
}

func TestDefaultJSONSerializerWrite(t *testing.T) {
	serializer := defaultJSONSerializer{}

	res := httptest.NewRecorder()
	p := person{Name: "Mojix", Age: 22}

	err := serializer.Write(res, p, "")
	assert.NoError(t, err)

	assert.Equal(t, "{\"name\":\"Mojix\",\"age\":22}\n", res.Body.String())

	res = httptest.NewRecorder()

	err = serializer.Write(res, p, "    ")
	assert.NoError(t, err)

	assert.Equal(t, "{\n    \"name\": \"Mojix\",\n    \"age\": 22\n}\n", res.Body.String())

	res = httptest.NewRecorder()

	// Channel type cannot be converted to JSON.
	err = serializer.Write(res, make(chan bool), "")
	assert.Error(t, err)

	httpErr := err.(*errors.HTTPError)

	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.Equal(t, httpErr.Err.Error(), httpErr.Message)
	assert.Error(t, httpErr.Err)
}

func TestDefaultJSONSerializerRead(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("{\"name\":\"Mojix\",\"age\":22}"))

	serializer := defaultJSONSerializer{}

	var p person
	err := serializer.Read(req, &p)

	assert.NoError(t, err)
	assert.Equal(t, "Mojix", p.Name)
	assert.Equal(t, 22, p.Age)

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader("{\"name\":\"Mojix\",\"age\":22}"))

	// Invalid argument passed to unmarshal.
	var p2 person
	err = serializer.Read(req, p2)
	assert.Error(t, err)

	httpErr := err.(*errors.HTTPError)
	_, ok := httpErr.Err.(*json.InvalidUnmarshalError)

	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.Equal(t, httpErr.Err.Error(), httpErr.Message)
	assert.Error(t, httpErr.Err)
	assert.True(t, ok)

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader("{\"name\":\"Mojix\",\"age\":22.5}"))

	err = serializer.Read(req, &p2)

	assert.Error(t, err)
	assert.Error(t, err.(*errors.HTTPError).Err)
	assert.Equal(t, http.StatusBadRequest, err.(*errors.HTTPError).Code)
}
