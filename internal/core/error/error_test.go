package errors

import "testing"

func TestNewFieldErrorPreservesFieldAndMessage(t *testing.T) {
	err := NewFieldError("product_name", "field is required")

	if err.Field != "product_name" || err.Message != "field is required" || err.Error() != "product_name: field is required" {
		t.Fatalf("expected field error details, got %#v", err)
	}
}

func TestNewRequestErrorPreservesFieldErrors(t *testing.T) {
	fieldError := NewFieldError("quantity", "must be greater than zero")
	err := NewRequestError("invalid request data", fieldError)

	if err.Err != "invalid request data" || len(err.Errors) != 1 || err.Errors[0] != fieldError || err.Error() != err.Err {
		t.Fatalf("expected request error details, got %#v", err)
	}
}
