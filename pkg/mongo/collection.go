package adapter

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

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
