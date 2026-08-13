package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/joaohgf/mjv-challenge/internal/core/domain"
)

type getRepositoryStub struct {
	order *domain.Order
	err   error
	id    string
}

func (stub *getRepositoryStub) Get(_ context.Context, id string) (*domain.Order, error) {
	stub.id = id
	return stub.order, stub.err
}

func TestGetOrderReturnsRepositoryResult(t *testing.T) {
	repository := &getRepositoryStub{order: &domain.Order{ID: "order-1"}}

	order, err := NewGetOrder(repository).Get(context.Background(), "order-1")

	if err != nil || order != repository.order || repository.id != "order-1" {
		t.Fatalf("expected repository result, got err=%v order=%#v id=%q", err, order, repository.id)
	}
}

func TestGetOrderReturnsRepositoryError(t *testing.T) {
	repository := &getRepositoryStub{err: errors.New("database unavailable")}

	order, err := NewGetOrder(repository).Get(context.Background(), "order-1")

	if err == nil || order != nil {
		t.Fatalf("expected repository error, got err=%v order=%#v", err, order)
	}
}
