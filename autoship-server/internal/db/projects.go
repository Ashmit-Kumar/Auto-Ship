package db

import (
	"context"
	"time"

	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/models"
)

// SaveProject inserts a project document into the "projects" collection.
func SaveProject(project *models.Project) error {
	collection := GetCollection("projects") // use GetCollection from mongo.go
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, project)
	return err
}

// func GetCollection(name string) *mongo.Collection {
// 	return Client.Database("autoship").Collection(name)
// }
