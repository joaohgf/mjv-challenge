package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joaohgf/mjv-challenge/config"
	"github.com/joaohgf/mjv-challenge/internal/core/usecase"
	outboxinbound "github.com/joaohgf/mjv-challenge/internal/inbound/outbox/adapter"
	outboxadapter "github.com/joaohgf/mjv-challenge/internal/outbound/outbox/adapter"
	publisheradapter "github.com/joaohgf/mjv-challenge/internal/outbound/publisher/adapter"
	publisherdto "github.com/joaohgf/mjv-challenge/internal/outbound/publisher/dto"
	publishermapper "github.com/joaohgf/mjv-challenge/internal/outbound/publisher/mapper"
	repositorymapper "github.com/joaohgf/mjv-challenge/internal/outbound/repository/mapper"
	repositorymodel "github.com/joaohgf/mjv-challenge/internal/outbound/repository/model"
	mongoadapter "github.com/joaohgf/mjv-challenge/pkg/mongo"
	rabbitadapter "github.com/joaohgf/mjv-challenge/pkg/rabbitmq"
)

// startOutboxRelay launches durable outbox dispatching independently from queue consumption.
func startOutboxRelay(ctx context.Context, mongo *mongoadapter.Repository[*repositorymodel.Order], settings config.Config) error {
	outbox := outboxadapter.NewStore(mongo.Collection(settings.Database.OutboxCollectionName), &repositorymapper.Order{}, settings.Queue.OutboxLease)
	publisher := rabbitadapter.NewPublisher[*publisherdto.Message[*publisherdto.Order]](settings.Queue)
	if err := publisher.Connect(ctx); err != nil {
		return fmt.Errorf("connecting outbox publisher: %w", err)
	}
	deadLetter := rabbitadapter.NewDeadLetterPublisher[*publisherdto.Message[*publisherdto.Order]](settings.Queue)
	if err := deadLetter.Connect(ctx); err != nil {
		_ = publisher.Close()
		return fmt.Errorf("connecting outbox dead-letter publisher: %w", err)
	}
	dispatcher := usecase.NewDispatchOutbox(
		outbox, publisheradapter.NewPublisher(publisher, &publishermapper.Order{}),
		publisheradapter.NewDeadLetterPublisher(deadLetter, &publishermapper.Order{}), settings.Queue.OutboxMaxAttempts,
	)
	relay := outboxinbound.NewRelay(dispatcher, settings.Queue.OutboxRetryInterval)
	slog.Info("outbox relay started", "interval", settings.Queue.OutboxRetryInterval)
	go runOutboxRelay(ctx, relay, publisher, deadLetter)
	return nil
}

// runOutboxRelay logs an unexpected relay stop and closes its publishers on shutdown.
func runOutboxRelay(ctx context.Context, relay *outboxinbound.Relay, publisher, deadLetter interface{ Close() error }) {
	defer closeRabbit(deadLetter)
	defer closeRabbit(publisher)
	if err := relay.Run(ctx); err != nil && ctx.Err() == nil {
		slog.Error("outbox relay stopped unexpectedly", "error", err)
	}
}
