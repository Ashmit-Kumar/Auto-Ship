package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsurePortsIndex creates a unique index on the `port` field of the ports
// collection. The unique constraint is what makes GetOrReserveValidFreePort
// race-safe: concurrent InsertOne calls for the same port get a duplicate-key
// error rather than both succeeding (which was the silent bug in the prior
// upsert-with-$setOnInsert pattern).
//
// Idempotent: Mongo returns the existing index name if it's already present
// with the same key+options. Will FAIL if the collection currently holds
// duplicate port docs (a remnant of the old buggy upsert behavior); operators
// must dedupe before startup will succeed.
func EnsurePortsIndex(collectionName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := GetCollection(collectionName)
	_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "port", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("port_unique"),
	})
	return err
}
