package kid

import (
	"encoding/xml"
	"net/http"
)

type (
	// XMLSerializer is the interface for marshaling and unmarshling XML data.
	XMLSerializer interface {
		// Write writes object with the given object to response.
		Write(c *Context, in any, indent string) error

		// Read reads request's body to the given object.
		Read(c *Context, out any) error
	}

	defaultXMLSerializer struct{}
)

// Verifying interface compliance.
var _ XMLSerializer = defaultXMLSerializer{}

// Write writes the given object as XML to response.
func (s defaultXMLSerializer) Write(c *Context, in any, indent string) error {
	encoder := xml.NewEncoder(c.Response())
	encoder.Indent("", indent)

	if err := encoder.Encode(in); err != nil {
		return NewHTTPError(http.StatusInternalServerError).WithMessage(err.Error()).WithError(err)
	}

	return nil
}

// Read reads request's body as XML and puts it in the given obj.
func (s defaultXMLSerializer) Read(c *Context, out any) error {
	if err := xml.NewDecoder(c.Request().Body).Decode(out); err != nil {
		if err.Error() == "non-pointer passed to Unmarshal" {
			return NewHTTPError(http.StatusInternalServerError).WithMessage(err.Error()).WithError(err)
		}
		return NewHTTPError(http.StatusBadRequest).WithMessage(err.Error()).WithError(err)
	}
	return nil
}
