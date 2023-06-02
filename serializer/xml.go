package serializer

import (
	"encoding/xml"
	"net/http"
)

// defaultXMLSerializer is the default XML serializer used in Kid.
type defaultXMLSerializer struct {
}

// Verifying interface compliance.
var _ Serializer = defaultXMLSerializer{}

// NewXMLSerializer returns a new XML serializer.
func NewXMLSerializer() Serializer {
	return defaultXMLSerializer{}
}

// Write writes the given object as XML to response.
func (s defaultXMLSerializer) Write(w http.ResponseWriter, in any, indent string) {
	encoder := xml.NewEncoder(w)
	encoder.Indent("", indent)

	if err := encoder.Encode(in); err != nil {
		panic(err)
	}
}

// Read reads request's body as XML and puts it in the given obj.
func (s defaultXMLSerializer) Read(req *http.Request, out any) error {
	if err := xml.NewDecoder(req.Body).Decode(out); err != nil {
		if err.Error() == "non-pointer passed to Unmarshal" {
			panic(err)
		}
		return err
	}
	return nil
}
