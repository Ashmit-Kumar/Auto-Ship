package db

import (
	"context"
	"time"

	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SaveProject inserts a project document into the "projects" collection.
func SaveProject(project *models.Project) error {
	collection := GetCollection("projects")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, project)
	return err
}

// UpdateProjectByID applies the given $set fields to a project. updated_at
// is stamped automatically so callers don't have to remember.
func UpdateProjectByID(id primitive.ObjectID, fields bson.M) error {
	collection := GetCollection("projects")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fields["updated_at"] = time.Now()
	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": fields})
	return err
}

// GetProjectByID returns a single project by its Mongo _id.
func GetProjectByID(id primitive.ObjectID) (*models.Project, error) {
	collection := GetCollection("projects")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var p models.Project
	if err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}
