package kid

import (
	"net/http"
	"net/url"
	"sync"
)

type (
	Context interface {
		Request() *http.Request
		Response() http.ResponseWriter
		Set(string, any)
		Get(string) (any, bool)
		Param(name string) string
		Params() Params
		QueryParam(name string) string
		QueryParamMultiple(name string) []string
		QueryParams() url.Values
		JSON(int, any) error
		JSONIndent(int, any, string) error
		NoContent(int)
	}

	context struct {
		request  *http.Request
		response http.ResponseWriter
		params   Params
		storage  Map
		kid      *Kid
		lock     sync.Mutex
	}
)

// Verifying interface compliance.
var _ Context = (*context)(nil)

func (c *context) reset(request *http.Request, response http.ResponseWriter) {
	c.request = request
	c.response = response
	c.storage = make(Map)
	c.params = make(Params)
}

func (c *context) setParams(params Params) {
	c.params = params
}

func (c *context) Request() *http.Request {
	return c.request
}

func (c *context) Response() http.ResponseWriter {
	return c.response
}

func (c *context) Param(name string) string {
	return c.params[name]
}

func (c *context) Params() Params {
	return c.params
}

func (c *context) QueryParam(name string) string {
	queryParam := c.request.URL.Query().Get(name)
	return queryParam
}

func (c *context) QueryParamMultiple(name string) []string {
	return c.request.URL.Query()[name]
}

func (c *context) QueryParams() url.Values {
	return c.request.URL.Query()
}

func (c *context) JSON(code int, obj any) error {
	c.writeContentType("application/json")
	c.response.WriteHeader(code)
	return c.kid.jsonSerializer.Marshal(c, obj, "")
}

func (c *context) JSONIndent(code int, obj any, indent string) error {
	c.writeContentType("application/json")
	c.response.WriteHeader(code)
	return c.kid.jsonSerializer.Marshal(c, obj, indent)
}

func (c *context) NoContent(code int) {
	c.response.WriteHeader(code)
}

func (c *context) writeContentType(contentType string) {
	header := c.response.Header()
	header.Set("Content-Type", contentType)
}

func (c *context) Set(key string, val any) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.storage[key] = val
}

func (c *context) Get(key string) (any, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	val, ok := c.storage[key]
	return val, ok
}
