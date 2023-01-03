package serializer

import (
	"encoding/xml"
	"net/http"

	"github.com/mojixcoder/kid/errors"
)

type defaultXMLSerializer struct {
}

// Verifying interface compliance.
var _ Serializer = defaultXMLSerializer{}

// NewXMLSerializer returns a new XML serializer.
func NewXMLSerializer() Serializer {
	return defaultXMLSerializer{}
}

// Write writes the given object as XML to response.
func (s defaultXMLSerializer) Write(w http.ResponseWriter, in any, indent string) error {
	encoder := xml.NewEncoder(w)
	encoder.Indent("", indent)

	if err := encoder.Encode(in); err != nil {
		return errors.NewHTTPError(http.StatusInternalServerError).WithMessage(err.Error()).WithError(err)
	}

	return nil
}

// Read reads request's body as XML and puts it in the given obj.
func (s defaultXMLSerializer) Read(req *http.Request, out any) error {
	if err := xml.NewDecoder(req.Body).Decode(out); err != nil {
		if err.Error() == "non-pointer passed to Unmarshal" {
			return errors.NewHTTPError(http.StatusInternalServerError).WithMessage(err.Error()).WithError(err)
		}
		return errors.NewHTTPError(http.StatusBadRequest).WithMessage(err.Error()).WithError(err)
	}
	return nil
}
