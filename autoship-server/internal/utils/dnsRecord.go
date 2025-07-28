package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// AddDNSRecord adds a DNS record for the given subdomain and IP address using Hostinger's API.
func addDNSRecord(subdomain, ip string) error {
    body := map[string]interface{}{
        "type": "A",
        "name": subdomain, // without a.com
        "content": ip,
        "ttl": 300,
    }

    jsonBody, _ := json.Marshal(body)
    req, _ := http.NewRequest("POST", "https://api.hostinger.com/dns-records", bytes.NewBuffer(jsonBody))
    req.Header.Set("Authorization", "Bearer "+os.Getenv("HOSTINGER_API_KEY"))
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 201 && resp.StatusCode != 200 {
        return fmt.Errorf("failed to add DNS: %v", resp.Status)
    }

    return nil
}

