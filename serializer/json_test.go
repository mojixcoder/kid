package serializer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

func TestDefaultJSONSerializer_Write(t *testing.T) {
	serializer := defaultJSONSerializer{}

	res := httptest.NewRecorder()
	p := person{Name: "Mojix", Age: 22}

	serializer.Write(res, p, "")
	assert.Equal(t, "{\"name\":\"Mojix\",\"age\":22}\n", res.Body.String())

	res = httptest.NewRecorder()

	serializer.Write(res, p, "    ")
	assert.Equal(t, "{\n    \"name\": \"Mojix\",\n    \"age\": 22\n}\n", res.Body.String())

	res = httptest.NewRecorder()

	// Channel type cannot be converted to JSON.
	assert.Panics(t, func() {
		serializer.Write(res, make(chan bool), "")
	})
}

func TestDefaultJSONSerializer_Read(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("{\"name\":\"Mojix\",\"age\":22}"))

	serializer := defaultJSONSerializer{}

	var p person
	err := serializer.Read(req, &p)

	assert.NoError(t, err)
	assert.Equal(t, "Mojix", p.Name)
	assert.Equal(t, 22, p.Age)

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader("{\"name\":\"Mojix\",\"age\":22}"))

	var p2 person

	// Invalid argument passed to unmarshal.
	assert.Panics(t, func() {
		serializer.Read(req, p2)
	})

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader("{\"name\":\"Mojix\",\"age\":22.5}"))

	err = serializer.Read(req, &p2)
	assert.Error(t, err)
}
