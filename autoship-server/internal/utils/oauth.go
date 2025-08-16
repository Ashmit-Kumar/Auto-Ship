// internal/utils/oauth.go
package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// GitHub OAuth endpoints and credentials
// var (
//
//	githubClientID     = os.Getenv("GITHUB_CLIENT_ID")
//	githubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
//	githubRedirectURI  = os.Getenv("GITHUB_REDIRECT_URI") // example: http://localhost:5000/github/callback
//
// )
var (
	githubClientID     = mustGetEnv("GITHUB_CLIENT_ID")
	githubClientSecret = mustGetEnv("GITHUB_CLIENT_SECRET")
	githubRedirectURI  = mustGetEnv("GITHUB_REDIRECT_URI")
)

func mustGetEnv(key string) string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	value := os.Getenv(key)
	if value == "" {
		log.Printf("[ERROR] Environment variable is not set or empty")
	} else {
		log.Printf("[OK] Loaded")
	}
	return value
}

// GetGitHubAuthURL returns the URL to redirect users to GitHub's OAuth page
func GetGitHubAuthURL() string {
	return fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user",
		githubClientID,
		githubRedirectURI,
	)
}

// ExchangeCodeForAccessToken exchanges the authorization code for an access token
func ExchangeCodeForAccessToken(code string) (string, error) {
	url := "https://github.com/login/oauth/access_token"
	fmt.Println("Exchanging code for access token") // Debugging
	fmt.Println("Code:", code)                      // Debugging
	// log.Println("GITHUB_CLIENT_ID:", githubClientID)
	// log.Println("GITHUB_CLIENT_SECRET:", githubClientSecret)
	// log.Println("GITHUB_REDIRECT_URI:", githubRedirectURI)
	// Prepare data for the POST request
	data := fmt.Sprintf(
		"client_id=%s&client_secret=%s&code=%s&redirect_uri=%s",
		githubClientID, githubClientSecret, code, githubRedirectURI,
	)

	// Create the POST request
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
	if err != nil {
		fmt.Println("Error creating request:", err) // Debugging
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // <-- Important!

	// Debugging: Print the request details
	fmt.Println("Sending request to GitHub OAuth token endpoint")
	fmt.Printf("Request URL: %s\n", url)
	fmt.Printf("Request Body: %s\n", data)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err) // Debugging
		return "", err
	}
	defer resp.Body.Close()

	// Debugging: Check the status code
	fmt.Printf("GitHub response status: %s\n", resp.Status)

	// Read and debug the response body
	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("GitHub response body: %s\n", string(respBody)) // Debugging

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Println("Error unmarshalling response:", err) // Debugging
		return "", err
	}

	// Check if access token is in the response
	if _, ok := result["access_token"]; !ok {
		// Print full result for debugging
		fmt.Println("GitHub token exchange error response:", result) // Debugging
		return "", fmt.Errorf("failed to get access token: %v", result)
	}

	// Extract the access token from the response
	accessToken, ok := result["access_token"].(string)
	if !ok {
		fmt.Println("Failed to parse access token from response:", result) // Debugging
		return "", fmt.Errorf("failed to get access token: %v", result)
	}

	return accessToken, nil
}

// GetGitHubUserInfo fetches user info from GitHub using the access token
func GetGitHubUserInfo(accessToken string) (map[string]interface{}, error) {
	url := "https://api.github.com/user"

	// Create the GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err) // Debugging
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	// Debugging: Print the request details
	fmt.Println("Fetching user info from GitHub")
	fmt.Printf("Request URL: %s\n", url)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err) // Debugging
		return nil, err
	}
	defer resp.Body.Close()

	// Debugging: Check the status code
	fmt.Printf("GitHub response status: %s\n", resp.Status)

	// Read and debug the response body
	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("GitHub response body: %s\n", string(respBody)) // Debugging

	var userInfo map[string]interface{}
	if err := json.Unmarshal(respBody, &userInfo); err != nil {
		fmt.Println("Error unmarshalling response:", err) // Debugging
		return nil, err
	}

	return userInfo, nil
}
