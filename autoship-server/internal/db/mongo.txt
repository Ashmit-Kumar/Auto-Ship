package db

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var mongoURI string

// SetMongoURI sets the MongoDB connection string.
func SetMongoURI(uri string) {
	mongoURI = uri
}

// Connect establishes a connection to MongoDB.
func Connect() {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

	var err error
	client, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Ping the database to verify the connection.
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	fmt.Println("Successfully connected to MongoDB!")
}

// Disconnect closes the MongoDB connection.
func Disconnect() {
	if client != nil {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
		fmt.Println("Connection to MongoDB closed.")
	}
}

// GetCollection returns a collection from the database.
func GetCollection(name string) *mongo.Collection {
	return client.Database("autoship").Collection(name)
}
