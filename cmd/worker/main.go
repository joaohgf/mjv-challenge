package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joaohgf/mjv-challenge/config"
	"github.com/joaohgf/mjv-challenge/internal/core/usecase"
	consumeradapter "github.com/joaohgf/mjv-challenge/internal/inbound/consumer/adapter"
	consumerdto "github.com/joaohgf/mjv-challenge/internal/inbound/consumer/dto"
	consumermapper "github.com/joaohgf/mjv-challenge/internal/inbound/consumer/mapper"
	repositoryadapter "github.com/joaohgf/mjv-challenge/internal/outbound/repository/adapter"
	repositorymapper "github.com/joaohgf/mjv-challenge/internal/outbound/repository/mapper"
	repositorymodel "github.com/joaohgf/mjv-challenge/internal/outbound/repository/model"
	mongoadapter "github.com/joaohgf/mjv-challenge/pkg/mongo"
	rabbitadapter "github.com/joaohgf/mjv-challenge/pkg/rabbitmq"
	"github.com/joaohgf/mjv-challenge/pkg/telemetry"
)

// main reports an unexpected worker stop after run returns.
func main() {
	if err := run(); err != nil {
		slog.Error("worker stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}

// run connects dependencies and consumes until a termination signal arrives.
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	settings := config.Load()
	shutdownTelemetry, err := telemetry.Start(ctx, settings.Telemetry)
	if err != nil {
		return fmt.Errorf("starting telemetry: %w", err)
	}
	defer telemetry.Close(ctx, shutdownTelemetry)
	mongo := mongoadapter.NewRepository[*repositorymodel.Order](settings.Database)
	if err := mongo.Connect(ctx); err != nil {
		return fmt.Errorf("connecting mongodb: %w", err)
	}
	defer closeMongo(context.Background(), mongo)
	slog.Info("mongodb connected", "database", settings.Database.MongoDatabase)
	if err := startOutboxRelay(ctx, mongo, settings); err != nil {
		return err
	}
	consumer := rabbitadapter.NewConsumer[*consumerdto.Message[*consumerdto.Order]](settings.Queue)
	if err := consumer.Connect(ctx); err != nil {
		return fmt.Errorf("connecting rabbitmq: %w", err)
	}
	defer closeRabbit(consumer)
	slog.Info("worker started", "queue", settings.Queue.Name)
	repository := repositoryadapter.NewRepository(mongo, &repositorymapper.Order{})
	updateOrder := usecase.NewUpdateOrder(repository)
	handler := consumeradapter.NewConsumer(consumer, &consumermapper.Order{}, updateOrder)
	if err := handler.Consume(ctx); err != nil && ctx.Err() == nil {
		return fmt.Errorf("consuming messages: %w", err)
	}
	slog.Info("worker stopped")
	return nil
}

// closeMongo logs shutdown failures without masking the worker result.
func closeMongo(ctx context.Context, repository interface{ Close(context.Context) error }) {
	if err := repository.Close(ctx); err != nil {
		slog.Error("closing mongodb connection", "error", err)
	}
}

// closeRabbit logs AMQP shutdown failures without masking the worker result.
func closeRabbit(rabbit interface{ Close() error }) {
	if err := rabbit.Close(); err != nil {
		slog.Error("closing rabbitmq connection", "error", err)
	}
}
