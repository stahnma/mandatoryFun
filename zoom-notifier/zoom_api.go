package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Zoom API token URL
const zoomTokenURL = "https://zoom.us/oauth/token"

type ZoomMeetingResponse struct {
	JoinURL string `json:"join_url"`
}

type ZoomTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Function to get a Zoom API access token
func GetZoomAccessToken() (string, error) {
	clientID := viper.GetString("ZOOM_API_CLIENT_ID")
	clientSecret := viper.GetString("ZOOM_API_CLIENT_SECRET")
	accountID := viper.GetString("ZOOM_API_ACCOUNT_ID")
	// Encode client credentials in Base64
	authString := fmt.Sprintf("%s:%s", clientID, clientSecret)
	authEncoded := base64.StdEncoding.EncodeToString([]byte(authString))

	// Prepare request body
	data := url.Values{}
	data.Set("grant_type", "account_credentials")
	data.Set("account_id", accountID)

	// Create request
	req, err := http.NewRequest("POST", zoomTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+authEncoded)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make API call: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-200 response
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s", body)
	}

	// Parse JSON response
	var tokenResponse ZoomTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return tokenResponse.AccessToken, nil
}

func getMeetingJoinLink(meetingID string, token string) (string, error) {
	log.Debugln("(getMeetingJoinLink) Getting meeting join link")
	url := fmt.Sprintf("https://api.zoom.us/v2/meetings/%s", meetingID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var zoomResponse ZoomMeetingResponse
	if err := json.NewDecoder(resp.Body).Decode(&zoomResponse); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return zoomResponse.JoinURL, nil
}

func callZoomApi(meeting_id string) (joinuri string) {
	// Fetch access token
	token, err := GetZoomAccessToken()
	log.Debugln("meeting_id: ", meeting_id)
	if err != nil {
		fmt.Printf("Error getting access token: %v\n", err)
		return
	}

	log.Debugln("Zoom Access Token: %s\n", token)
	joinuri, err = getMeetingJoinLink(meeting_id, token)

	return joinuri
}
