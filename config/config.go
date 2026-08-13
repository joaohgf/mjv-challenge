package config

import (
	"os"
	"time"
)

type (
	// Config groups the runtime settings shared by the API and worker.
	Config struct {
		HTTPAddr    string
		SwaggerHost string
		Database    *Database
		Queue       *Queue
	}
	// Database identifies the MongoDB server, database and collection.
	Database struct {
		MongoDBURI           string
		MongoDatabase        string
		CollectionName       string
		OutboxCollectionName string
		SaveTimeout          time.Duration
	}
	// Queue identifies the RabbitMQ queues and broker connection.
	Queue struct {
		Name                string
		DeadLetterName      string
		RabbitMQURL         string
		PublishTimeout      time.Duration
		OutboxLease         time.Duration
		OutboxRetryInterval time.Duration
	}
)

// Load reads all application settings, using local defaults when unset.
func Load() Config {
	return Config{
		HTTPAddr:    value("HTTP_ADDR", ":8080"),
		SwaggerHost: value("SWAGGER_HOST", "localhost:8080"),
		Queue:       LoadQueue(),
		Database:    LoadDatabase(),
	}
}

// LoadDatabase returns the MongoDB connection and collection settings.
func LoadDatabase() *Database {
	target := new(Database)
	target.MongoDBURI = value("MONGODB_URI", "mongodb://root:root@localhost:27017/?authSource=admin&replicaSet=rs0")
	target.MongoDatabase = value("MONGODB_DATABASE", "orders")
	target.CollectionName = value("MONGO_COLLECTION_NAME", "orders")
	target.OutboxCollectionName = value("MONGO_OUTBOX_COLLECTION_NAME", "outbox")
	target.SaveTimeout = duration("MONGODB_SAVE_TIMEOUT", 5*time.Second)
	return target
}

// value returns an environment value or its fallback when it is unset or empty.
func value(key, fallback string) string {
	if result := os.Getenv(key); result != "" {
		return result
	}
	return fallback
}

// duration reads a positive duration from the environment or uses its fallback.
func duration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
