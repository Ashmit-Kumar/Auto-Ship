package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Project struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username   string             `bson:"username" json:"username"`
	RepoURL    string             `bson:"repo_url" json:"repo_url"`
	RepoName   string             `bson:"repo_name" json:"repo_name"`
	ProjectType string            `bson:"project_type" json:"project_type"`
	HostedURL  string             `bson:"hosted_url" json:"hosted_url"`
}
