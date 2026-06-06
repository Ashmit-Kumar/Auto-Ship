package utils

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/cloud"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/config"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/db"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// maxRecycleAttempts bounds how many "available" docs we'll try before
// giving up on the recycle pool and falling through to a fresh allocation.
// Protects against an unbounded loop if the firewall is broken and every
// reauth fails (we'd just keep marking docs available and re-picking them).
const maxRecycleAttempts = 5

// GetOrReserveValidFreePort allocates a host port for a new container.
//
// Two-stage strategy:
//
//  1. Recycle pool — try the lowest-numbered doc with status "available"
//     (set by db.ReleasePort when a previous container was deleted). The
//     cloud firewall rule from the previous owner is intentionally left
//     alive; reauthorize is a no-op (AWS dedupes by description, Azure by
//     rule name), so recycling skips the NSG/SG round-trip in the common
//     case.
//  2. Fresh allocation — watermark scan above the highest port currently
//     in the collection, insert a new doc atomically. Unique index on
//     `port` (db.EnsurePortsIndex) makes concurrent claims for the same
//     port fail with a duplicate-key error; the loser falls through.
//
// Mongo client: reuses the long-lived pooled client from db.GetCollection
// rather than opening a fresh connection per call.
func GetOrReserveValidFreePort(containerName string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	coll := db.GetCollection(config.Get().MongoCollection)

	// Stage 1: try to recycle a freed port.
	for attempt := 0; attempt < maxRecycleAttempts; attempt++ {
		var recycled models.PortMapping
		err := coll.FindOneAndUpdate(
			ctx,
			bson.M{"status": "available"},
			bson.M{"$set": bson.M{
				"status":        "used",
				"containerName": containerName,
				"timestamp":     time.Now(),
			}},
			options.FindOneAndUpdate().SetSort(bson.D{{Key: "port", Value: 1}}),
		).Decode(&recycled)

		if err == mongo.ErrNoDocuments {
			break // nothing recyclable; fall through to fresh allocation
		}
		if err != nil {
			return 0, fmt.Errorf("failed to claim recyclable port: %w", err)
		}

		if !IsPortAvailable(recycled.Port) {
			// Doc says available but the OS port is bound by something else.
			// Prune the phantom and try the next available doc.
			log.Printf("Recycled port %d not OS-free; pruning stale entry", recycled.Port)
			_, _ = coll.DeleteOne(ctx, bson.M{"_id": recycled.ID})
			continue
		}

		if err := cloud.Get().AuthorizePort(recycled.Port); err != nil {
			log.Printf("Failed to re-authorize recycled port %d: %v", recycled.Port, err)
			// Put it back in the available pool — the cloud problem is likely
			// transient and a later allocation should try again.
			_, _ = coll.UpdateOne(ctx,
				bson.M{"_id": recycled.ID},
				bson.M{"$set": bson.M{
					"status":    "available",
					"timestamp": time.Now(),
				}},
			)
			continue
		}
		return recycled.Port, nil
	}

	// Stage 2: nothing recyclable worked. Scan above the watermark and insert.
	var portDoc struct {
		Port int `bson:"port"`
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "port", Value: -1}})
	err := coll.FindOne(ctx, bson.M{}, opts).Decode(&portDoc)
	if err == mongo.ErrNoDocuments {
		log.Println("No ports found in DB, starting from 2000")
		portDoc.Port = 1999
	} else if err != nil {
		return 0, fmt.Errorf("failed to find latest port: %w", err)
	}
	startPort := portDoc.Port + 1

	for port := startPort; port <= 65535; port++ {
		if !IsPortAvailable(port) {
			continue
		}
		_, err := coll.InsertOne(ctx, models.PortMapping{
			Port:          port,
			ContainerPort: 0,
			Status:        "used",
			ContainerName: containerName,
			Timestamp:     time.Now(),
		})
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				continue
			}
			return 0, fmt.Errorf("failed to reserve port %d: %w", port, err)
		}
		if err := cloud.Get().AuthorizePort(port); err != nil {
			log.Printf("Failed to open port %d in firewall: %v", port, err)
			if _, delErr := coll.DeleteOne(ctx, bson.M{
				"port":          port,
				"containerName": containerName,
			}); delErr != nil {
				log.Printf("Failed to release Mongo reservation for port %d: %v", port, delErr)
			}
			continue
		}
		return port, nil
	}

	return 0, fmt.Errorf("no free ports found")
}

func IsPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	_ = listener.Close()
	return true
}

// tryDefaultPorts checks common default ports inside the container.
func tryDefaultPorts(containerID string) (int, error) {
	defaultPorts := []int{3000, 5000, 8080, 80, 8000}
	for _, port := range defaultPorts {
		cmd := exec.Command("docker", "exec", containerID, "bash", "-c", fmt.Sprintf("netstat -tuln | grep ':%d '", port))
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no default port matched")
}

// detectPortWithNetstat uses netstat inside the container to find open ports.
func detectPortWithNetstat(containerID string) (int, error) {
	fmt.Println("Running netstat inside the container to detect open ports... ", containerID)

	time.Sleep(5 * time.Second) // wait for the container to be fully up

	cmd := exec.Command("docker", "exec", containerID, "netstat", "-tuln")
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("failed to exec netstat: %w", err)
	}
	output, err := cmd.CombinedOutput()
	fmt.Println("Netstat Output:\n", string(output))
	if err != nil {
		return 0, fmt.Errorf("failed to exec netstat: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fmt.Println("Inspecting line:", line)
		fields := strings.Fields(line)
		if len(fields) >= 4 && (strings.HasPrefix(fields[0], "tcp") || strings.HasPrefix(fields[0], "udp")) {
			addr := fields[3]
			if parts := strings.Split(addr, ":"); len(parts) > 1 {
				portStr := parts[len(parts)-1]
				if port, err := strconv.Atoi(portStr); err == nil {
					return port, nil
				}
			}
		}
	}
	return 0, fmt.Errorf("no open port found")
}

func DetectExposedPort(containerID string) (int, error) {
	fmt.Println("Detecting Exposed Port using default ports... ", containerID)
	if port, err := tryDefaultPorts(containerID); err == nil {
		return port, nil
	}
	fmt.Println("No default port matched, trying dynamic detection...")
	return detectPortWithNetstat(containerID)
}

func FindFreeHostPort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	addr := listener.Addr().String()
	parts := strings.Split(addr, ":")
	portStr := parts[len(parts)-1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, err
	}
	return port, nil
}
