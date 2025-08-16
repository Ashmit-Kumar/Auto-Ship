package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PortMapping struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"` // MongoDB document ID
	Port          int                `bson:"port"`          // Host port on EC2
	ContainerPort int                `bson:"containerPort"` // Port exposed inside the Docker container
	Status        string             `bson:"status"`        // "available" or "used"
	ContainerName string             `bson:"containerName"` // Docker container name
	Timestamp     time.Time          `bson:"timestamp"`     // Timestamp of allocation or update
}
