package kid

import (
	"net/http"
	"net/url"
	"sync"
)

const contentTypeHeader string = "Content-Type"

// Context is the context of current HTTP request.
// It holds data related to current HTTP request.
type Context struct {
	request  *http.Request
	response ResponseWriter
	params   Params
	storage  Map
	kid      *Kid
	lock     sync.Mutex
}

// newContext returns a new empty context.
func newContext(k *Kid) *Context {
	c := Context{kid: k}
	return &c
}

// reset resets the context.
func (c *Context) reset(request *http.Request, response http.ResponseWriter) {
	c.request = request
	c.response = newResponse(response)
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
func (c *Context) Response() ResponseWriter {
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

// Path returns request's path used for matching request to a handler.
func (c *Context) Path() string {
	u := c.request.URL
	if u.RawPath != "" {
		return u.RawPath
	}
	return u.Path
}

// Method returns request method.
func (c *Context) Method() string {
	return c.request.Method
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
	params := c.request.URL.Query()[name]
	if params == nil {
		return []string{}
	}
	return params
}

// QueryParams returns all of the query parameters.
func (c *Context) QueryParams() url.Values {
	return c.request.URL.Query()
}

// mustWrite writes raw bytes to the response.
func (c *Context) mustWrite(blob []byte) {
	if _, err := c.Response().Write(blob); err != nil {
		panic(err)
	}
}

// JSON sends JSON response with the given status code.
func (c *Context) JSON(code int, obj any) {
	c.writeContentType("application/json")
	c.response.WriteHeader(code)
	c.kid.jsonSerializer.Write(c.Response(), obj, "")
}

// JSONIndent sends JSON response with the given status code.
// Sends response with the given indent.
func (c *Context) JSONIndent(code int, obj any, indent string) {
	c.writeContentType("application/json")
	c.response.WriteHeader(code)
	c.kid.jsonSerializer.Write(c.Response(), obj, indent)
}

// JSONByte sends JSON response with the given status code.
// Writes JSON blob untouched to response.
func (c *Context) JSONByte(code int, blob []byte) {
	c.writeContentType("application/json")
	c.response.WriteHeader(code)
	c.mustWrite(blob)
}

// ReadJSON reads request's body as JSON and stores it in the given object.
// The object must be a pointer.
func (c *Context) ReadJSON(out any) error {
	return c.kid.jsonSerializer.Read(c.Request(), out)
}

// XML sends XML response with the given status code.
//
// Returns an error if an error happened during sending response otherwise returns nil.
func (c *Context) XML(code int, obj any) {
	c.writeContentType("application/xml")
	c.response.WriteHeader(code)
	c.kid.xmlSerializer.Write(c.Response(), obj, "")
}

// XMLIndent sends XML response with the given status code.
// Sends response with the given indent.
func (c *Context) XMLIndent(code int, obj any, indent string) {
	c.writeContentType("application/xml")
	c.response.WriteHeader(code)
	c.kid.xmlSerializer.Write(c.Response(), obj, indent)
}

// XMLByte sends XML response with the given status code.
// Writes JSON blob untouched to response.
func (c *Context) XMLByte(code int, blob []byte) {
	c.writeContentType("application/xml")
	c.response.WriteHeader(code)
	c.mustWrite(blob)
}

// ReadXML reads request's body as XML and stores it in the given object.
// The object must be a pointer.
func (c *Context) ReadXML(out any) error {
	return c.kid.xmlSerializer.Read(c.Request(), out)
}

// HTML sends HTML response with the given status code.
//
// tpl must be a relative path to templates root directory.
// Defaults to "templates/".
func (c *Context) HTML(code int, tpl string, data any) {
	c.writeContentType("text/html")
	c.response.WriteHeader(code)
	c.kid.htmlRenderer.RenderHTML(c.Response(), tpl, data)
}

// HTMLString sends bare string as HTML response with the given status code.
func (c *Context) HTMLString(code int, tpl string) {
	c.writeContentType("text/html")
	c.response.WriteHeader(code)
	c.mustWrite([]byte(tpl))
}

// String sends bare string as a plain text response with the given status code.
func (c *Context) String(code int, data string) {
	c.writeContentType("text/plain")
	c.response.WriteHeader(code)
	c.mustWrite([]byte(data))
}

// Byte sends bare bytes as response with the given status code.
func (c *Context) Byte(code int, data []byte) {
	c.writeContentType("application/octet-stream")
	c.response.WriteHeader(code)
	c.mustWrite([]byte(data))
}

// NoContent returns an empty response with the given status code.
func (c *Context) NoContent(code int) {
	c.response.WriteHeader(code)
	c.response.WriteHeaderNow()
}

// GetResponseHeader gets a response header.
func (c *Context) GetResponseHeader(key string) string {
	return c.response.Header().Get(key)
}

// SetResponseHeader sets a header to the response.
func (c *Context) SetResponseHeader(key, value string) {
	c.response.Header().Set(key, value)
}

// SetRequestHeader sets a header to the request.
func (c *Context) SetRequestHeader(key, value string) {
	c.request.Header.Set(key, value)
}

// GetRequestHeader gets a request header.
func (c *Context) GetRequestHeader(key string) string {
	return c.request.Header.Get(key)
}

// writeContentType sets content type header for response.
// It won't overwrite content type if it's already set.
func (c *Context) writeContentType(contentType string) {
	if c.GetResponseHeader(contentTypeHeader) == "" {
		c.SetResponseHeader(contentTypeHeader, contentType)
	}
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

// Debug returns whether we are in debug mode or not.
func (c *Context) Debug() bool {
	return c.kid.Debug()
}
