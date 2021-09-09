package local_library

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-playground/validator"
)

// Supported MIME Content-Types.
const (
	_mimeApplicationJSON = "application/json"
)

var _validate = validator.New()

// Bind deserializes a request body into the given destination.
//
// The type of binding is dependent on the "Content-Type" for the request.
// If the type is "application/json" it will use "json.NewDecoder".
// This function may invoke data validation after deserialization.
func Bind(r *http.Request, destination interface{}) error {
	// We default to application/json if content type is not specified but return
	// http.StatusUnsupportedMediaType if it's specified but not supported.
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		ct = _mimeApplicationJSON
	}

	switch {
	case strings.HasPrefix(ct, _mimeApplicationJSON):
		return bindJSON(r.Context(), r.Body, destination)
	default:
		return NewErrorf(http.StatusUnsupportedMediaType, "unsupported media type: %s", ct)
	}
}

func bindJSON(ctx context.Context, r io.Reader, destination interface{}) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	// In order to detect empty request body, we check for len(b) to be zero.
	// ReadAll is defined to read from src until EOF and it does not
	// treat it as en error as it happens when using json.Decoder.
	if len(b) == 0 {
		return NewErrorf(400, "Request body is empty")
	}

	if err := json.Unmarshal(b, destination); err != nil {
		switch e := err.(type) {
		case *json.UnmarshalTypeError:
			return NewErrorf(400,
				"Unmarshal type error: expected=%v, got=%v, field=%v, offset=%v",
				e.Type, e.Value, e.Field, e.Offset)
		case *json.SyntaxError:
			return NewErrorf(400, "Syntax error: offset=%v, error=%v", e.Offset, e)
		default:
			return NewErrorf(400, err.Error())
		}
	}

	if err := _validate.StructCtx(ctx, destination); err != nil {
		var invalidValidationError *validator.InvalidValidationError
		if errors.As(err, &invalidValidationError) {
			// We choose to ignore errors related to types
			// that can't be validated like time.Time and slices.
			return nil
		}

		message := err.Error()

		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			fields := make([]string, 0, len(validationErrs))
			for _, v := range validationErrs {
				fields = append(fields, v.Field())
			}
			message = fmt.Sprintf("invalid fields: %s", strings.Join(fields, ","))
		}

		return NewErrorf(http.StatusUnprocessableEntity, "validation_error: %s", message)
	}

	return nil
}
