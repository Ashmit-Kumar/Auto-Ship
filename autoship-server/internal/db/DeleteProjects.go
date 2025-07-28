package db

import (
    "context"
    "time"
    "go.mongodb.org/mongo-driver/bson"
)

// ...existing code...

func DeleteProjectByContainerName(containerName string) error {
    collection := Client.Database("autoship").Collection("projects")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _, err := collection.DeleteOne(ctx, bson.M{"containername": containerName})
    return err
}