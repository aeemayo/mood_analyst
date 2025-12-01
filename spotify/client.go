package spotify

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	spotifyAuthURL   = "https://accounts.spotify.com/api/token"
	spotifySearchURL = "https://api.spotify.com/v1/search"
	spotifyAPIURL    = "https://api.spotify.com/v1"
)

// Track represents a Spotify track
type Track struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Artists []struct {
		Name string `json:"name"`
	} `json:"artists"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	PreviewURL string `json:"preview_url"`
	URI        string `json:"uri"`
}

// User represents a Spotify user
type User struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// Playlist represents a Spotify playlist
type Playlist struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
}

// SearchResult represents Spotify search results
type SearchResult struct {
	Tracks struct {
		Items []Track `json:"items"`
	} `json:"tracks"`
}

// Client represents a Spotify API client
type Client struct {
	clientID     string
	clientSecret string
	accessToken  string
}

// NewClient creates a new Spotify client
func NewClient(clientID, clientSecret string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// Authenticate gets an access token from Spotify
func (c *Client) Authenticate() error {
	auth := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))

	data := url.Values{}

	// Check if we have a refresh token in env
	refreshToken := os.Getenv("SPOTIFY_REFRESH_TOKEN")
	if refreshToken != "" {
		log.Printf("Using refresh token for user authentication")
		data.Set("grant_type", "refresh_token")
		data.Set("refresh_token", refreshToken)
	} else {
		log.Printf("No refresh token found, using client credentials (limited API access)")
		data.Set("grant_type", "client_credentials")
	}

	req, err := http.NewRequest("POST", spotifyAuthURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return fmt.Errorf("access token not found in response")
	}

	// Log the scope we received
	if scope, ok := result["scope"].(string); ok {
		log.Printf("Authenticated with scopes: %s", scope)
	}

	c.accessToken = accessToken
	return nil
}

// SearchTracks searches for tracks on Spotify
func (c *Client) SearchTracks(query string, limit int) ([]Track, error) {
	if c.accessToken == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("type", "track")
	params.Set("limit", fmt.Sprintf("%d", limit))

	searchURL := spotifySearchURL + "?" + params.Encode()

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+c.accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search tracks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed with status %d: %s", resp.StatusCode, body)
	}

	var result SearchResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return result.Tracks.Items, nil
}

// GetRecommendations gets track recommendations based on seed tracks and mood parameters
func (c *Client) GetRecommendations(seedTracks []string, seedGenres []string, moodParams map[string]interface{}, limit int) ([]Track, error) {
	if c.accessToken == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	params := url.Values{}

	if len(seedTracks) > 0 {
		// Spotify expects comma-separated IDs, but params.Encode() will URL-encode commas to %2C
		// We need to manually build this part of the URL
		params.Set("seed_tracks", strings.Join(seedTracks, ","))
	}

	if len(seedGenres) > 0 {
		params.Set("seed_genres", strings.Join(seedGenres, ","))
	}

	// Add mood parameters
	for key, value := range moodParams {
		params.Set(key, fmt.Sprintf("%v", value))
	}

	params.Set("limit", fmt.Sprintf("%d", limit))

	recURL := spotifyAPIURL + "/recommendations?" + params.Encode()

	// Debug logging
	log.Printf("Recommendations URL: %s", recURL)
	log.Printf("Seed tracks: %v, Seed genres: %v", seedTracks, seedGenres)

	req, err := http.NewRequest("GET", recURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create recommendations request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+c.accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if bodyStr == "" {
			bodyStr = "(empty response)"
		}
		// Log the full URL and error for debugging
		log.Printf("Recommendations API error - Status: %d, Body: %s", resp.StatusCode, bodyStr)
		return nil, fmt.Errorf("recommendations failed with status %d: %s (URL: %s)", resp.StatusCode, bodyStr, recURL)
	}

	var result struct {
		Tracks []Track `json:"tracks"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode recommendations response: %w", err)
	}

	return result.Tracks, nil
}

// GetCurrentUser gets the current authenticated user
func (c *Client) GetCurrentUser() (*User, error) {
	if c.accessToken == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	req, err := http.NewRequest("GET", spotifyAPIURL+"/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+c.accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get user failed with status %d: %s", resp.StatusCode, body)
	}

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	return &user, nil
}

// CreatePlaylist creates a new playlist for a user
func (c *Client) CreatePlaylist(userID, name, description string) (*Playlist, error) {
	if c.accessToken == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	data := map[string]string{
		"name":        name,
		"description": description,
		"public":      "false",
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal playlist data: %w", err)
	}

	url := fmt.Sprintf("%s/users/%s/playlists", spotifyAPIURL, userID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create playlist request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+c.accessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create playlist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create playlist failed with status %d: %s", resp.StatusCode, body)
	}

	var playlist Playlist
	err = json.NewDecoder(resp.Body).Decode(&playlist)
	if err != nil {
		return nil, fmt.Errorf("failed to decode playlist response: %w", err)
	}

	return &playlist, nil
}

// AddTracksToPlaylist adds tracks to a playlist
func (c *Client) AddTracksToPlaylist(playlistID string, trackURIs []string) error {
	if c.accessToken == "" {
		return fmt.Errorf("not authenticated")
	}

	data := map[string][]string{
		"uris": trackURIs,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal tracks data: %w", err)
	}

	url := fmt.Sprintf("%s/playlists/%s/tracks", spotifyAPIURL, playlistID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create add tracks request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+c.accessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add tracks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("add tracks failed with status %d: %s", resp.StatusCode, body)
	}

	return nil
}

// LoadFromEnv loads Spotify credentials from environment variables
func LoadFromEnv() (*Client, error) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET environment variables are required")
	}

	return NewClient(clientID, clientSecret), nil
}
