package kid

import (
	"encoding/json"
	"errors"
	"net/http"
)

type (
	JSONSerializer interface {
		Marshal(Context, any, string) error
		Unmarshal(Context, any) error
	}
	defaultJSONSerializer struct{}
)

// Verifying interface compliance.
var _ JSONSerializer = defaultJSONSerializer{}

func (s defaultJSONSerializer) Marshal(c Context, obj any, indent string) error {
	encoder := json.NewEncoder(c.Response())
	encoder.SetIndent("", indent)

	if err := encoder.Encode(obj); err != nil {
		return NewHTTPError(http.StatusInternalServerError).WithMessage(err.Error()).WithError(err)
	}

	return nil
}

func (s defaultJSONSerializer) Unmarshal(c Context, obj any) error {
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
