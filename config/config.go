package config

import "os"

type Config struct {
	HTTPAddr      string
	MongoDBURI    string
	MongoDatabase string
	RabbitMQURL   string
}

func Load() Config {
	return Config{
		HTTPAddr:      value("HTTP_ADDR", ":8080"),
		MongoDBURI:    value("MONGODB_URI", "mongodb://root:root@localhost:27017/?authSource=admin"),
		MongoDatabase: value("MONGODB_DATABASE", "jobs"),
		RabbitMQURL:   value("RABBITMQ_URL", "amqp://app:app@localhost:5672/"),
	}
}

func value(key, fallback string) string {
	if result := os.Getenv(key); result != "" {
		return result
	}

	return fallback
}
