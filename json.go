package kid

import (
	"encoding/json"
	"errors"
	"net/http"
)

type (
	// JSONSerializer is the interface for marshaling and unmarshling JSON data.
	JSONSerializer interface {
		Marshal(*Context, any, string) error
		Unmarshal(*Context, any) error
	}

	// defaultJSONSerializer is the default Kid's JSON serializer.
	defaultJSONSerializer struct{}
)

// Verifying interface compliance.
var _ JSONSerializer = defaultJSONSerializer{}

// Marshal writes the given object as JSON to response.
func (s defaultJSONSerializer) Marshal(c *Context, obj any, indent string) error {
	encoder := json.NewEncoder(c.Response())
	encoder.SetIndent("", indent)

	if err := encoder.Encode(obj); err != nil {
		return NewHTTPError(http.StatusInternalServerError).WithMessage(err.Error()).WithError(err)
	}

	return nil
}

// Unmarshal reads request's body as JSON and puts it in the given obj.
func (s defaultJSONSerializer) Unmarshal(c *Context, obj any) error {
	err := json.NewDecoder(c.Request().Body).Decode(obj)

	jsonErr := &json.SyntaxError{}
	typeErr := &json.UnmarshalTypeError{}

	if errors.As(err, &jsonErr) || errors.As(err, &typeErr) {
		return NewHTTPError(http.StatusBadRequest).WithMessage(err.Error()).WithError(err)
	} else if err != nil {
		return NewHTTPError(http.StatusInternalServerError).WithMessage(err.Error()).WithError(err)
	}

	return err
}
