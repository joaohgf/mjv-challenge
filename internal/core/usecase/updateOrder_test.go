package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/core/enum"
)

type repositoryStub struct {
	updated []*domain.Order
	err     error
}

func (stub *repositoryStub) Update(_ context.Context, source *domain.Order) (*domain.Order, error) {
	copy := *source
	stub.updated = append(stub.updated, &copy)
	return source, stub.err
}

func TestUpdateOrderStopsWhenFirstPersistenceFails(t *testing.T) {
	repository := &repositoryStub{err: errors.New("database unavailable")}
	usecase := NewUpdateOrder(repository)
	usecase.wait = func(time.Duration) { t.Fatal("expected no wait after persistence error") }

	order, err := usecase.Update(context.Background(), &domain.Order{ID: "order-1"})

	if err == nil || order != nil || len(repository.updated) != 1 {
		t.Fatalf("expected one failed update, got err=%v order=%#v updates=%d", err, order, len(repository.updated))
	}
}

func TestUpdateOrderSetsUpdatedAtForEachPersistedState(t *testing.T) {
	repository := new(repositoryStub)
	order := &domain.Order{ID: "order-1"}

	usecase := NewUpdateOrder(repository)
	usecase.wait = func(time.Duration) {}
	_, err := usecase.Update(context.Background(), order)

	if err != nil || len(repository.updated) != 2 {
		t.Fatalf("expected two updates, got err=%v updates=%d", err, len(repository.updated))
	}
	if repository.updated[0].Status != enum.Processando || repository.updated[0].UpdatedAt == nil {
		t.Fatalf("expected processing state with updated_at, got %#v", repository.updated[0])
	}
	if repository.updated[1].Status != enum.Processado || repository.updated[1].UpdatedAt == nil {
		t.Fatalf("expected processed state with updated_at, got %#v", repository.updated[1])
	}
}
