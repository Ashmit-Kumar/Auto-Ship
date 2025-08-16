package db

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func DeleteProjectByContainerName(containerName string) error {
	collection := GetCollection("projects")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.DeleteOne(ctx, bson.M{"containername": containerName})
	return err
}
