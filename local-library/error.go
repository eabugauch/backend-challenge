package local_library

import (
	"fmt"
)

// Error is our custom error's interface implementation.
type Error struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status"`
}

// Error returns a string message of the error. It is a concatenation of Code and Message fields.
// This means the Error implements the error interface.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}

// NewErrorf creates a new error with the given status code and the message
// formatted according to args and format.
func NewErrorf(statusCode int, format string, args ...interface{}) error {
	return &Error{
		Message:    fmt.Sprintf(format, args...),
		StatusCode: statusCode,
	}
}
