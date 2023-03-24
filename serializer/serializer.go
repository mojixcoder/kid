// Package serializer provides an interface to read from request body or write to response body.
// Can be used for reading/writing JSON, XML, MessagePack, etc.
//
// Currently the supported ones are JSON and XML.
package serializer

import (
	"net/http"
)

// Serializer is the interface for reading from request body or writing to response body.
//
// It can be implemented to read/write custom JSON/XML serializers.
type Serializer interface {
	// Write writes object with the given indent to response body.
	Write(w http.ResponseWriter, in any, indent string) error

	// Read reads request body and store it in the given object.
	Read(req *http.Request, out any) error
}
