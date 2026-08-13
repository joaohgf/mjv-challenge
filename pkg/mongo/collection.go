package adapter

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ensureUniqueIDIndex prevents two documents from sharing a business identifier.
func (repository *Repository[M]) ensureUniqueIDIndex(ctx context.Context) error {
	_, err := repository.collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: repository.idField(), Value: 1}},
		Options: options.Index().SetUnique(true).SetPartialFilterExpression(
			bson.M{repository.idField(): bson.M{"$type": "string"}},
		),
	})
	if err != nil {
		return fmt.Errorf("create unique index: %w", err)
	}
	return nil
}

// collection resolves the configured database collection for each operation.
func (repository *Repository[M]) collection() *mongo.Collection {
	return repository.Collection(repository.config.CollectionName)
}

// Collection exposes a MongoDB collection for an infrastructure-specific adapter.
func (repository *Repository[M]) Collection(name string) *mongo.Collection {
	return repository.client.Database(repository.config.MongoDatabase).Collection(name)
}

func (repository *Repository[M]) filter(id string) bson.M {
	return bson.M{repository.idField(): id}
}

func (*Repository[M]) idField() string {
	var model M
	return model.GetIDField()
}
