package usecase

import (
	"context"
	"time"

	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/core/enum"
	"github.com/joaohgf/mjv-challenge/internal/core/port"
)

// UpdateOrder persists the two processing states of a consumed order message.
type UpdateOrder struct {
	repository port.PersistentUpdater[*domain.Order]
	wait       func(time.Duration)
}

// NewUpdateOrder creates the worker use case and keeps its processing delay.
func NewUpdateOrder(repository port.PersistentUpdater[*domain.Order]) *UpdateOrder {
	target := new(UpdateOrder)
	target.repository = repository
	target.wait = time.Sleep
	return target
}

// Update persists each state transition so a reader can observe in-flight work.
func (uo *UpdateOrder) Update(ctx context.Context, source *domain.Order) (*domain.Order, error) {
	source.Status = enum.Processando
	source.UpdatedAt = now()
	saved, err := uo.repository.Update(ctx, source)
	if err != nil {
		return nil, err
	}
	uo.wait(time.Second * 2)
	saved.Status = enum.Processado
	saved.UpdatedAt = now()
	saved, err = uo.repository.Update(ctx, saved)
	return saved, err
}

// now returns a distinct UTC timestamp pointer for each state transition.
func now() *time.Time {
	value := time.Now().UTC()
	return &value
}
