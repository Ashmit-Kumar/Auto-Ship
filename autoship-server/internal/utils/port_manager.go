package utils

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"strings"
	"strconv"
	"os/exec"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var (
    MongoURI        string
    DatabaseName    string
    CollectionName  string
    SecurityGroupID string
    Region          string
)

func init() {
    _ = godotenv.Load(".env") // Loads .env file from current directory (ignore error if not present)

    MongoURI        = os.Getenv("MONGO_URI")
    DatabaseName    = os.Getenv("MONGO_DB")
    CollectionName  = os.Getenv("MONGO_DB_COLLECTION")
    SecurityGroupID = os.Getenv("EC2_SECURITY_GROUP_ID")
    Region          = os.Getenv("AWS_REGION")
}

// // GetOrReserveValidFreePort finds an unused port and opens it in the EC2 security group
// func GetOrReserveValidFreePort(containerName string) (int, error) {
// 	ctx := context.Background()

// 	// 1. Connect to MongoDB
// 	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer client.Disconnect(ctx)

// 	coll := client.Database(DatabaseName).Collection(CollectionName)

// 	// 2. Search for an available port in DB
// 	var portDoc struct {
// 		Port int `bson:"port"`
// 	}
// 	err = coll.FindOneAndUpdate(ctx, bson.M{"status": "available"}, bson.M{
// 		"$set": bson.M{"status": "used", "containerName": containerName, "timestamp": time.Now()},
// 	}).Decode(&portDoc)
// 	if err == mongo.ErrNoDocuments {
// 		// No free ports in DB; optionally generate a new one
// 		return 0, fmt.Errorf("no free ports in DB")
// 	} else if err != nil {
// 		return 0, err
// 	}

// 	// 3. Check if it's free on this machine
// 	if !IsPortAvailable(portDoc.Port) {
// 		// Update status back to "available"
// 		_, _ = coll.UpdateOne(ctx, bson.M{"port": portDoc.Port}, bson.M{"$set": bson.M{"status": "available"}})
// 		return GetOrReserveValidFreePort(containerName)
// 	}

// 	// 4. Open port in EC2 Security Group
// 	if err := AuthorizeEC2Port(portDoc.Port); err != nil {
// 		log.Printf("Failed to open port %d in SG: %v", portDoc.Port, err)
// 		return 0, err
// 	}

// 	return portDoc.Port, nil
// }




func GetOrReserveValidFreePort(containerName string) (int, error) {
	ctx := context.Background()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		return 0, err
	}
	defer client.Disconnect(ctx)

	coll := client.Database(DatabaseName).Collection(CollectionName)

	// Step 1: Get the latest used or available port
	var portDoc struct {
		Port int `bson:"port"`
	}
	opts := options.FindOne().SetSort(bson.D{{"port", -1}})
	err = coll.FindOne(ctx, bson.M{}, opts).Decode(&portDoc)
	startPort := 2000 // default fallback
	if err == nil {
		startPort = portDoc.Port
	}

	// Step 2: Try finding a free port by incrementing
	for port := startPort; port <= 65535; port++ {
		if IsPortAvailable(port) {
			// Try reserving in DB (ensure atomicity in multi-user environments)
			_, err := coll.UpdateOne(ctx,
				bson.M{"port": port},
				bson.M{
					"$setOnInsert": bson.M{
						"port":          port,
						"status":        "used",
						"containerName": containerName,
						"timestamp":     time.Now(),
					},
				},
				options.Update().SetUpsert(true),
			)
			if err == nil {
				// Open in EC2 SG
				if err := AuthorizeEC2Port(port); err != nil {
					log.Printf("Failed to open port %d in SG: %v", port, err)
					continue
				}
				return port, nil
			}
		}
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

func AuthorizeEC2Port(port int) error {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(Region),
	}))

	svc := ec2.New(sess)

	_, err := svc.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(SecurityGroupID),
		IpPermissions: []*ec2.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int64(int64(port)),
				ToPort:     aws.Int64(int64(port)),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("Auto-opened for container hosting"),
					},
				},
			},
		},
	})

	if err != nil && !strings.Contains(err.Error(), "InvalidPermission.Duplicate") {
		return err
	}

	return nil
}

// tryDefaultPorts checks common default ports like 3000, 5000, 8080 inside the container.
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
	
	// Sleep for few seconds to ensure the container is fully up
	time.Sleep(5 * time.Second) // <-- wait for the container to be fully up
	
	// Execute netstat command inside the container
	cmd := exec.Command("docker", "exec", containerID, "netstat", "-tuln")
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("failed to exec netstat: %w", err)
	}
	// output, err := cmd.Output()
	// fmt.Println("Inspecting line:", line)
	output, err := cmd.CombinedOutput() // <-- not just Output()
	fmt.Println("Netstat Output:\n", string(output)) // <-- helpful to print it
	if err != nil {
		return 0, fmt.Errorf("failed to exec netstat: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fmt.Println("Inspecting line:", line) // <-- print each line for debugging
		// Example line: "tcp        0      0 0.0.0.0:8080"
		fields := strings.Fields(line)
		if len(fields) >= 4 && (strings.HasPrefix(fields[0], "tcp") || strings.HasPrefix(fields[0], "udp")) {
			addr := fields[3] // usually 0.0.0.0:8080
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
	// Try default common ports first
	fmt.Println("Detecting Exposed Port using default ports... ", containerID)
	if port, err := tryDefaultPorts(containerID); err == nil {
		return port, nil
	}
	fmt.Println("No default port matched, trying dynamic detection...")
	fmt.Println("Detecting port using netstat inside the container... ", containerID)
	// Fallback to dynamic detection using netstat
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