package main

import (
	"context"
	"log"

	"github.com/example/mjv-challenge/config"
	"github.com/example/mjv-challenge/core/usecase"
	mongoadapter "github.com/example/mjv-challenge/mongo/adapter"
	rabbitadapter "github.com/example/mjv-challenge/rabbitmq/adapter"
)

func main() {
	ctx := context.Background()
	settings := config.Load()
	repository, client, err := mongoadapter.NewJobRepository(ctx, settings.MongoDBURI, settings.MongoDatabase)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	connection, err := rabbitadapter.Connect(settings.RabbitMQURL)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer channel.Close()

	deliveries, err := rabbitadapter.Consume(channel)
	if err != nil {
		log.Fatal(err)
	}

	storeJob := usecase.NewStoreJob(repository)
	for delivery := range deliveries {
		process(ctx, storeJob, delivery)
	}
}
