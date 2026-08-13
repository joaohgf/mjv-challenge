package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/enum"
)

type createOrderStub struct {
	created *domain.Order
	err     error
}

func (stub *createOrderStub) Create(_ context.Context, order *domain.Order) (*domain.Order, error) {
	stub.created = order
	return order, stub.err
}

func TestCreateOrderInitializesBeforeTransactionalCreation(t *testing.T) {
	creator := new(createOrderStub)

	created, err := NewCreateOrder(creator).Create(context.Background(), domain.NewOrder())

	if err != nil || created.ID == "" || created.CreatedAt.IsZero() || created.Status != enum.Criado {
		t.Fatalf("expected initialized order, got err=%v order=%#v", err, created)
	}
	if creator.created != created {
		t.Fatal("expected initialized order sent to transactional creator")
	}
}

func TestCreateOrderReturnsNoOrderWhenTransactionalCreationFails(t *testing.T) {
	creator := &createOrderStub{err: errors.New("database unavailable")}

	created, err := NewCreateOrder(creator).Create(context.Background(), domain.NewOrder())

	if err == nil || created != nil || creator.created == nil {
		t.Fatalf("expected transactional persistence error, got err=%v order=%#v", err, created)
	}
}
