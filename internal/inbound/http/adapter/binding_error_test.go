package adapter

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

type bindingStub struct {
	Name   string `json:"name" validate:"required"`
	Amount int    `json:"amount" validate:"gte=1"`
}

func TestBindingErrorUsesDTOJSONTags(t *testing.T) {
	err := validator.New().Struct(bindingStub{})
	result := bindingError[bindingStub](err)

	if result.Err != "invalid request data" || len(result.Errors) != 2 {
		t.Fatalf("expected validation errors, got %#v", result)
	}
	if result.Errors[0].Field != "name" || result.Errors[1].Field != "amount" {
		t.Fatalf("expected JSON field names, got %#v", result.Errors)
	}
}
