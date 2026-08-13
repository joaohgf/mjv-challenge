package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joaohgf/mjv-challenge/config"
	docs "github.com/joaohgf/mjv-challenge/docs"
	"github.com/joaohgf/mjv-challenge/internal/bootstrap"
	repositorymodel "github.com/joaohgf/mjv-challenge/internal/outbound/repository/model"
	mongoadapter "github.com/joaohgf/mjv-challenge/pkg/mongo"
	"github.com/joaohgf/mjv-challenge/pkg/telemetry"
)

// @title MJV Challenge API
// @version 1.0
// @description API para criação assíncrona de pedidos.
// @host localhost:8080
// @BasePath /
// main reports a startup or server failure after the API stops.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		slog.Error("api stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}

// run connects infrastructure and starts the HTTP server.
func run(ctx context.Context) error {
	settings := config.Load()
	docs.SwaggerInfo.Host = settings.SwaggerHost
	shutdownTelemetry, err := telemetry.Start(ctx, settings.Telemetry)
	if err != nil {
		return fmt.Errorf("starting telemetry: %w", err)
	}
	defer telemetry.Close(context.Background(), shutdownTelemetry)
	mongo := mongoadapter.NewRepository[*repositorymodel.Order](settings.Database)
	if err := mongo.Connect(ctx); err != nil {
		return fmt.Errorf("connecting mongodb: %w", err)
	}
	defer closeMongo(context.Background(), mongo)
	slog.Info("mongodb connected", "database", settings.Database.MongoDatabase)
	engine := bootstrap.NewEngine(mongo, settings)
	server := newServer(settings.HTTPAddr, engine)
	slog.Info("api started", "address", settings.HTTPAddr)
	return serve(ctx, server, settings.HTTPShutdownTimeout)
}

// closeMongo logs shutdown failures without masking the original API error.
func closeMongo(ctx context.Context, repository interface{ Close(context.Context) error }) {
	if err := repository.Close(ctx); err != nil {
		slog.Error("closing mongodb connection", "error", err)
	}
}
