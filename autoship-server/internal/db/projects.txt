package db

import (
	"context"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/models"
)

func SaveProject(project *models.Project) error {
	collection := Client.Database("autoship").Collection("projects")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, project)
	return err
}
func GetCollection(name string) *mongo.Collection {
	return Client.Database("autoship").Collection(name)
}

