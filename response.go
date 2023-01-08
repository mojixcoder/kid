package kid

import (
	"bufio"
	"net"
	"net/http"
)

type (
	// ResponseWriter is a wrapper for http.ResponseWriter
	// to make using http.ResponseWriter's methods easier.
	ResponseWriter interface {
		http.ResponseWriter
		http.Hijacker
		http.Flusher

		// WriteHeaderNow writes status code.
		WriteHeaderNow()

		// Size returns number of bytes written to response.
		Size() int

		// Written returns true if response has already been written otherwise returns false.
		Written() bool
	}

	// response implements ResponseWriter.
	response struct {
		http.ResponseWriter
		written bool
		status  int
		size    int
	}
)

// Verifying interface compliance.
var _ ResponseWriter = (*response)(nil)

// newResponse returns a new response writer.
func newResponse(w http.ResponseWriter) ResponseWriter {
	response := response{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
	return &response
}

// WriteHeader sets status code.
func (r *response) WriteHeader(code int) {
	if r.Written() {
		return
	}

	r.status = code
}

// WriteHeaderNow writes status code.
// Status code should already be specified using response.WriteHeader method.
func (r *response) WriteHeaderNow() {
	if r.Written() {
		return
	}

	r.written = true
	r.ResponseWriter.WriteHeader(r.status)
}

// Write writes byte data to response.
func (r *response) Write(b []byte) (int, error) {
	r.WriteHeaderNow()

	n, err := r.ResponseWriter.Write(b)
	r.size += n

	return n, err
}

// Size returns number of bytes written.
func (r *response) Size() int {
	return r.size
}

// Written returns true if response has already been written otherwise returns false.
func (r *response) Written() bool {
	return r.written
}

// Flush implements the http.Flusher interface.
func (r *response) Flush() {
	r.WriteHeaderNow()
	r.ResponseWriter.(http.Flusher).Flush()
}

// Hijack implements the http.Hijacker interface.
func (r *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.written = true
	return r.ResponseWriter.(http.Hijacker).Hijack()
}
