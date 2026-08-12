package main

import (
	"context"
	"log"

	"github.com/example/mjv-challenge/core/usecase"
	rabbitadapter "github.com/example/mjv-challenge/rabbitmq/adapter"
	"github.com/rabbitmq/amqp091-go"
)

func process(ctx context.Context, storeJob usecase.StoreJob, delivery amqp091.Delivery) {
	job, err := rabbitadapter.Decode(delivery)
	if err == nil {
		err = storeJob.Execute(ctx, job)
	}
	if err != nil {
		log.Printf("job failed: %v", err)
		_ = delivery.Nack(false, true)
		return
	}

	if err := delivery.Ack(false); err != nil {
		log.Printf("ack job: %v", err)
	}
}
