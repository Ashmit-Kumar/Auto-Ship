package db

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var UserCollection *mongo.Collection

// Connect to MongoDB
func Connect() {
	// Get MongoDB URI from environment variables
	mongoURI := os.Getenv("MONGO_URI")
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
