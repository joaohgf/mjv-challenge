package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joaohgf/mjv-challenge/config"
	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/core/usecase"
	outboxinbound "github.com/joaohgf/mjv-challenge/internal/inbound/outbox/adapter"
	outboxadapter "github.com/joaohgf/mjv-challenge/internal/outbound/outbox/adapter"
	publisheradapter "github.com/joaohgf/mjv-challenge/internal/outbound/publisher/adapter"
	publisherdto "github.com/joaohgf/mjv-challenge/internal/outbound/publisher/dto"
	publishermapper "github.com/joaohgf/mjv-challenge/internal/outbound/publisher/mapper"
	repositoryadapter "github.com/joaohgf/mjv-challenge/internal/outbound/repository/adapter"
	repositorymapper "github.com/joaohgf/mjv-challenge/internal/outbound/repository/mapper"
	repositorymodel "github.com/joaohgf/mjv-challenge/internal/outbound/repository/model"
	mongoadapter "github.com/joaohgf/mjv-challenge/pkg/mongo"
	rabbitadapter "github.com/joaohgf/mjv-challenge/pkg/rabbitmq"
)

// startRelay launches durable outbox dispatching independently from queue consumption.
func startRelay(ctx context.Context, mongo *mongoadapter.Repository[*repositorymodel.Order], repository *repositoryadapter.Repository[*domain.Order, *repositorymodel.Order], settings config.Config) error {
	outbox := outboxadapter.NewStore(mongo.Collection(settings.Database.OutboxCollectionName), &repositorymapper.OrderMapper{}, settings.Queue.OutboxLease)
	publisher := rabbitadapter.NewPublisher[*publisherdto.Message[*publisherdto.Order]](settings.Queue)
	if err := publisher.Connect(ctx); err != nil {
		return fmt.Errorf("connecting outbox publisher: %w", err)
	}
	dispatcher := usecase.NewDispatchOutbox(outbox, publisheradapter.NewPublisher(publisher, &publishermapper.Order{}))
	relay := outboxinbound.NewRelay(dispatcher, settings.Queue.OutboxRetryInterval)
	slog.Info("outbox relay started", "interval", settings.Queue.OutboxRetryInterval)
	go runRelay(ctx, relay, publisher)
	return nil
}

// runRelay logs an unexpected relay stop and closes its publisher on shutdown.
func runRelay(ctx context.Context, relay *outboxinbound.Relay, publisher interface{ Close() error }) {
	defer closeRabbit(publisher)
	if err := relay.Run(ctx); err != nil && ctx.Err() == nil {
		slog.Error("outbox relay stopped unexpectedly", "error", err)
	}
}
