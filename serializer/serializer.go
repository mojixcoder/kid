// Package serializer implements JSON and XML serializers.
// It also has a serializer interface for building custom serializers.
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
