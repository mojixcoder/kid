package kid

import (
	"encoding/json"
	"net/http"
)

type (
	// JSONSerializer is the interface for marshaling and unmarshling JSON data.
	JSONSerializer interface {
		// Write writes object with the given object to response.
		Write(c *Context, in any, indent string) error

		// Read reads request's body to the given object.
		Read(c *Context, out any) error
	}

	// defaultJSONSerializer is the default Kid's JSON serializer.
	defaultJSONSerializer struct{}
)

// Verifying interface compliance.
var _ JSONSerializer = defaultJSONSerializer{}

// Marshal writes the given object as JSON to response.
func (s defaultJSONSerializer) Write(c *Context, in any, indent string) error {
	encoder := json.NewEncoder(c.Response())
	encoder.SetIndent("", indent)

	if err := encoder.Encode(in); err != nil {
		return NewHTTPError(http.StatusInternalServerError).WithMessage(err.Error()).WithError(err)
	}

	return nil
}

// Unmarshal reads request's body as JSON and puts it in the given obj.
func (s defaultJSONSerializer) Read(c *Context, out any) error {
	err := json.NewDecoder(c.Request().Body).Decode(out)

	if _, ok := err.(*json.SyntaxError); ok {
		return NewHTTPError(http.StatusBadRequest).WithMessage(err.Error()).WithError(err)
	} else if _, ok := err.(*json.UnmarshalTypeError); ok {
		return NewHTTPError(http.StatusBadRequest).WithMessage(err.Error()).WithError(err)
	} else if err != nil {
		return NewHTTPError(http.StatusInternalServerError).WithMessage(err.Error()).WithError(err)
	}

	return err
}
