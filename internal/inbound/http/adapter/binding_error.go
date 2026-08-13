package adapter

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	errs "github.com/joaohgf/mjv-challenge/internal/core/error"
)

// bindingError presents a binding failure for any DTO without leaking Go field names.
func bindingError[T any](err error) *errs.RequestError {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return errs.NewRequestError("invalid request body")
	}
	fieldErrors := make([]*errs.FieldError, 0, len(validationErrors))
	for _, validationError := range validationErrors {
		fieldErrors = append(fieldErrors, errs.NewFieldError(jsonField[T](validationError.Field()), validationMessage(validationError.Tag())))
	}
	return errs.NewRequestError("invalid request data", fieldErrors...)
}

func jsonField[T any](fieldName string) string {
	field, found := reflect.TypeFor[T]().FieldByName(fieldName)
	if !found {
		return strings.ToLower(fieldName)
	}
	return strings.Split(field.Tag.Get("json"), ",")[0]
}

func validationMessage(tag string) string {
	switch tag {
	case "required":
		return "field is required"
	case "gte":
		return "must be greater than zero"
	default:
		return tag
	}
}
