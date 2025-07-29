package utils

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
	"time"
)

// Mutex to prevent race conditions when appending to file
var fileMutex sync.Mutex

// AppendJSONToFile appends a JSON object to a JSON array in a file (creates file if not present)
func AppendJSONToFile(path string, newEntry map[string]interface{}) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	var data []map[string]interface{}

	// Read existing entries if file exists
	if f, err := os.Open(path); err == nil {
		defer f.Close()
		decoder := json.NewDecoder(f)
		if err := decoder.Decode(&data); err != nil && err != io.EOF {
			return err
		}
	}

	// Append new entry
	data = append(data, newEntry)

	// Write back to file
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}


// WaitForResponse waits for a specific ID to appear in the JSON response file
func WaitForResponse(path string, id string, timeout time.Duration) (map[string]interface{}, error) {
	start := time.Now()

	for {
		// Check timeout
		if time.Since(start) > timeout {
			return nil, errors.New("timeout waiting for deployment response")
		}

		// Open and read file
		f, err := os.Open(path)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		defer f.Close()

		var responses []map[string]interface{}
		if err := json.NewDecoder(f).Decode(&responses); err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// Look for matching ID
		for _, resp := range responses {
			if respID, ok := resp["id"].(string); ok && respID == id {
				return resp, nil
			}
		}

		time.Sleep(1 * time.Second)
	}
}
