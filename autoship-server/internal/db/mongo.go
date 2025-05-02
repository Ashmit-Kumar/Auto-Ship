package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var UserCollection *mongo.Collection
var mongoURI string

// SetMongoURI allows you to set the MongoDB URI
func SetMongoURI(uri string) {
	mongoURI = uri
}

// Connect to MongoDB
func Connect() {
	if mongoURI == "" {
		log.Fatal("MONGO_URI is not set")
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Connect to MongoDB
	var err error
	Client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal("Error connecting to MongoDB: ", err)
	}

	// Ping the database
	err = Client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Error pinging MongoDB: ", err)
	}

	// Initialize collections
	UserCollection = Client.Database("autoship").Collection("users")

	log.Println("Connected to MongoDB successfully!")
}

// Disconnect from MongoDB
func Disconnect() {
	err := Client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal("Error disconnecting from MongoDB: ", err)
	}
}
func GetClient() *mongo.Client {
	return Client
}
