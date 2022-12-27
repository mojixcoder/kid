package kid

import (
	"net/http"
	"net/url"
	"sync"
)

// Context is the context of current HTTP request.
// It holds data related to current HTTP request.
type Context struct {
	request  *http.Request
	response http.ResponseWriter
	params   Params
	storage  Map
	kid      *Kid
	lock     sync.Mutex
}

// reset resets the context.
func (c *Context) reset(request *http.Request, response http.ResponseWriter) {
	c.request = request
	c.response = response
	c.storage = make(Map)
	c.params = make(Params)
}

// setParams sets request's path parameters.
func (c *Context) setParams(params Params) {
	c.params = params
}

// Request returns plain request of current HTTP request.
func (c *Context) Request() *http.Request {
	return c.request
}

// Response returns plain response of current HTTP request.
func (c *Context) Response() http.ResponseWriter {
	return c.response
}

// Param returns path parameter's value.
func (c *Context) Param(name string) string {
	return c.params[name]
}

// Params returns all of the path parameters.
func (c *Context) Params() Params {
	return c.params
}

// QueryParam returns value of a query parameter
func (c *Context) QueryParam(name string) string {
	queryParam := c.request.URL.Query().Get(name)
	return queryParam
}

// QueryParamMultiple returns multiple values of a query parameter.
//
// Useful when query parameters are like ?name=x&name=y.
func (c *Context) QueryParamMultiple(name string) []string {
	return c.request.URL.Query()[name]
}

// QueryParams returns all of the query parameters.
func (c *Context) QueryParams() url.Values {
	return c.request.URL.Query()
}

// JSON sends JSON response with the given status code.
//
// Returns an error happenedd during sending response.
func (c *Context) JSON(code int, obj any) error {
	c.writeContentType("application/json")
	c.response.WriteHeader(code)
	return c.kid.jsonSerializer.Write(c, obj, "")
}

// JSONIndent sends JSON response with the given status code.
// Creates response with the given indent.
//
// Returns an error happenedd during sending response.
func (c *Context) JSONIndent(code int, obj any, indent string) error {
	c.writeContentType("application/json")
	c.response.WriteHeader(code)
	return c.kid.jsonSerializer.Write(c, obj, indent)
}

// ReadJSON reads request's body as JSON and stores it in the given object.
// The object must be a pointer.
func (c *Context) ReadJSON(out any) error {
	return c.kid.jsonSerializer.Read(c, out)
}

// NoContent returns an empty response with the given status code.
func (c *Context) NoContent(code int) {
	c.response.WriteHeader(code)
}

// writeContentType sets content type of response.
func (c *Context) writeContentType(contentType string) {
	header := c.response.Header()
	header.Set("Content-Type", contentType)
}

// Set sets a key-value pair to current request's context.
func (c *Context) Set(key string, val any) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.storage[key] = val
}

// Get gets a value from current request's context.
func (c *Context) Get(key string) (any, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	val, ok := c.storage[key]
	return val, ok
}
