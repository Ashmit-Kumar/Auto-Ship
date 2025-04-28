package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// GitHub OAuth endpoints and credentials
var (
	githubClientID     = os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
	githubRedirectURI  = os.Getenv("GITHUB_REDIRECT_URI") // example: http://localhost:3000/auth/github/callback
)

// GetGitHubAuthURL returns the URL to redirect users to GitHub's OAuth page
func GetGitHubAuthURL() string {
	return fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user",
		githubClientID,
		githubRedirectURI,
	)
}

// ExchangeCodeForAccessToken exchanges the authorization code for an access token
// func ExchangeCodeForAccessToken(code string) (string, error) {
// 	url := "https://github.com/login/oauth/access_token"

// 	payload := map[string]string{
// 		"client_id":     githubClientID,
// 		"client_secret": githubClientSecret,
// 		"code":          code,
// 		"redirect_uri":  githubRedirectURI,
// 	}

// 	body, _ := json.Marshal(payload)

// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
// 	if err != nil {
// 		return "", err
// 	}
// 	req.Header.Set("Accept", "application/json")
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	respBody, _ := ioutil.ReadAll(resp.Body)

// 	var result map[string]interface{}
// 	if err := json.Unmarshal(respBody, &result); err != nil {
// 		return "", err
// 	}

// 	accessToken, ok := result["access_token"].(string)
// 	if !ok {
// 		return "", fmt.Errorf("failed to get access token")
// 	}

// 	return accessToken, nil
// }
func ExchangeCodeForAccessToken(code string) (string, error) {
	url := "https://github.com/login/oauth/access_token"

	data := fmt.Sprintf(
		"client_id=%s&client_secret=%s&code=%s&redirect_uri=%s",
		githubClientID, githubClientSecret, code, githubRedirectURI,
	)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // <-- Important!

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("failed to get access token: %v", result)
	}

	return accessToken, nil
}

// GetGitHubUserInfo fetches user info from GitHub using the access token
func GetGitHubUserInfo(accessToken string) (map[string]interface{}, error) {
	url := "https://api.github.com/user"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var userInfo map[string]interface{}
	if err := json.Unmarshal(respBody, &userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}
