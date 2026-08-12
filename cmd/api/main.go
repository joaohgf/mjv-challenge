package main

import (
	"log"
	"net/http"

	apiadapter "github.com/example/mjv-challenge/api/adapter"
	"github.com/example/mjv-challenge/config"
	"github.com/example/mjv-challenge/core/usecase"
	rabbitadapter "github.com/example/mjv-challenge/rabbitmq/adapter"
)

func main() {
	settings := config.Load()
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

	publisher, err := rabbitadapter.NewPublisher(channel)
	if err != nil {
		log.Fatal(err)
	}

	server := apiadapter.NewHTTPServer(usecase.NewPublishJob(publisher))
	log.Printf("api listening on %s", settings.HTTPAddr)
	log.Fatal(http.ListenAndServe(settings.HTTPAddr, server.Handler()))
}
