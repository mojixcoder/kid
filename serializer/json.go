package serializer

import (
	"encoding/json"
	"net/http"
)

// defaultJSONSerializer is the default Kid's JSON serializer.
type defaultJSONSerializer struct {
}

// Verifying interface compliance.
var _ Serializer = defaultJSONSerializer{}

// NewJSONSerializer returns a new JSON serializer.
func NewJSONSerializer() Serializer {
	return defaultJSONSerializer{}
}

// Marshal writes the given object as JSON to response.
func (s defaultJSONSerializer) Write(w http.ResponseWriter, in any, indent string) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", indent)

	if err := encoder.Encode(in); err != nil {
		return newHTTPErrorFromError(http.StatusInternalServerError, err)
	}

	return nil
}

// Unmarshal reads request's body as JSON and puts it in the given obj.
func (s defaultJSONSerializer) Read(req *http.Request, out any) error {
	if err := json.NewDecoder(req.Body).Decode(out); err != nil {
		if _, ok := err.(*json.InvalidUnmarshalError); ok {
			return newHTTPErrorFromError(http.StatusInternalServerError, err)
		}
		return newHTTPErrorFromError(http.StatusBadRequest, err)
	}
	return nil
}
