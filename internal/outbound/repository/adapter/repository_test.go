package adapter

import (
	"context"
	"errors"
	"testing"
)

type (
	domainStub  struct{ ID string }
	modelStub   struct{ ID string }
	mapperStub  struct{}
	storageStub struct {
		created *modelStub
		updated *modelStub
		err     error
	}
)

func (model *modelStub) GetID() string { return model.ID }

func (*modelStub) GetIDField() string { return "id" }

func (mapperStub) To(domain *domainStub) *modelStub { return &modelStub{ID: domain.ID} }

func (mapperStub) From(model *modelStub) *domainStub { return &domainStub{ID: model.ID} }

func (storage *storageStub) Create(_ context.Context, model *modelStub) (*modelStub, error) {
	storage.created = model
	return model, storage.err
}

func (storage *storageStub) Update(_ context.Context, model *modelStub) (*modelStub, error) {
	storage.updated = model
	return model, storage.err
}

func (storage *storageStub) Get(_ context.Context, id string) (*modelStub, error) {
	return &modelStub{ID: id}, storage.err
}

func TestRepositoryWrapsStoreError(t *testing.T) {
	repository := NewRepository[*domainStub, *modelStub](&storageStub{err: errors.New("unavailable")}, mapperStub{})

	_, err := repository.Create(context.Background(), &domainStub{ID: "order-1"})

	if err == nil {
		t.Fatal("expected store error")
	}
}

func TestRepositoryMapsAtRepositoryBoundary(t *testing.T) {
	storage := new(storageStub)
	repository := NewRepository[*domainStub, *modelStub](storage, mapperStub{})

	result, err := repository.Create(context.Background(), &domainStub{ID: "order-1"})

	if err != nil || storage.created.ID != "order-1" || result.ID != "order-1" {
		t.Fatalf("expected mapped repository creation, got err=%v created=%#v result=%#v", err, storage.created, result)
	}
}

func TestRepositoryMapsRetrievedModel(t *testing.T) {
	repository := NewRepository[*domainStub, *modelStub](new(storageStub), mapperStub{})

	result, err := repository.Get(context.Background(), "order-1")

	if err != nil || result.ID != "order-1" {
		t.Fatalf("expected mapped repository get, got err=%v result=%#v", err, result)
	}
}

func TestRepositoryMapsUpdateAtRepositoryBoundary(t *testing.T) {
	storage := new(storageStub)
	repository := NewRepository[*domainStub, *modelStub](storage, mapperStub{})

	result, err := repository.Update(context.Background(), &domainStub{ID: "order-1"})

	if err != nil || storage.updated.ID != "order-1" || result.ID != "order-1" {
		t.Fatalf("expected mapped repository update, got err=%v updated=%#v result=%#v", err, storage.updated, result)
	}
}

func TestRepositoryWrapsUpdateStoreError(t *testing.T) {
	repository := NewRepository[*domainStub, *modelStub](&storageStub{err: errors.New("unavailable")}, mapperStub{})

	_, err := repository.Update(context.Background(), &domainStub{ID: "order-1"})

	if err == nil {
		t.Fatal("expected update store error")
	}
}
