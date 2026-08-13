package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound identifies a requested resource that does not exist in storage.
	ErrNotFound = errors.New("resource not found")
)

type (
	// FieldError identifies one invalid input field and its client-safe message.
	FieldError struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	}
	// RequestError groups client-safe errors returned by an inbound adapter.
	RequestError struct {
		Err    string        `json:"err"`
		Errors []*FieldError `json:"errors,omitempty"`
	}
)

// NewFieldError creates an error associated with the supplied client field.
func NewFieldError(field, message string) *FieldError {
	return &FieldError{Field: field, Message: message}
}

// Error implements error while preserving the field and message for adapters.
func (err *FieldError) Error() string {
	return fmt.Sprintf("%s: %s", err.Field, err.Message)
}

// NewRequestError creates a structured request error with optional field details.
func NewRequestError(message string, fieldErrors ...*FieldError) *RequestError {
	return &RequestError{Err: message, Errors: fieldErrors}
}

// Error implements error while retaining structured error details for adapters.
func (err *RequestError) Error() string {
	return err.Err
}
