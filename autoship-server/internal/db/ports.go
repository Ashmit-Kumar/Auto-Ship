package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// ReleasePort marks the port owned by containerName as available, so that
// GetOrReserveValidFreePort's recycle branch can hand it to a future deploy
// instead of skipping past it forever as a watermarked phantom.
//
// Static projects don't have a port reservation; for them this is a no-op
// (MatchedCount == 0 is treated as success rather than an error, so callers
// can invoke this unconditionally on every delete).
//
// The bson key is "containerName" (camelCase) to match the PortMapping
// model's tag — DELIBERATELY different from Project's snake_case
// "container_name" tag. The two collections use inconsistent naming
// conventions; unifying them would require a data migration, out of scope
// here.
func ReleasePort(collectionName, containerName string) error {
	coll := GetCollection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := coll.UpdateOne(ctx,
		bson.M{"containerName": containerName},
		bson.M{"$set": bson.M{
			"status":    "available",
			"timestamp": time.Now(),
		}},
	)
	if err != nil {
		return fmt.Errorf("release port for %s: %w", containerName, err)
	}
	return nil
}
