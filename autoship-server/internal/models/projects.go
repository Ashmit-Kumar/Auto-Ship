package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Project struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username      string             `bson:"username" json:"username"`
	RepoURL       string             `bson:"repo_url" json:"repo_url"`
	RepoName      string             `bson:"repo_name" json:"repo_name"`
	ProjectType   string             `bson:"project_type" json:"project_type"`
	HostedURL     string             `bson:"hosted_url" json:"hosted_url"`
	StartCommand  string             `bson:"start_command" json:"start_command"`
	ContainerPort int                `bson:"container_port" json:"container_port"`
	HostPort      int                `bson:"host_port" json:"host_port"`
	ContainerName string             `bson:"container_name" json:"container_name"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}
