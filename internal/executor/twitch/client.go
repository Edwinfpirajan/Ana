// Package twitch provides Twitch integration for AnaStreamer
package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/anastreamer/ana/internal/config"
	"github.com/anastreamer/ana/internal/executor"
	"github.com/anastreamer/ana/internal/llm"
	"github.com/anastreamer/ana/pkg/logger"
	"github.com/rs/zerolog"
)

const (
	twitchAPIURL  = "https://api.twitch.tv/helix"
	twitchAuthURL = "https://id.twitch.tv/oauth2"
)

// Executor implements the Twitch action executor
type Executor struct {
	clientID      string
	clientSecret  string
	redirectURI   string
	broadcasterID string
	accessToken   string
	refreshToken  string
	client        *http.Client
	log           zerolog.Logger
	enabled       bool
}

// NewExecutor creates a new Twitch executor
func NewExecutor(cfg config.TwitchConfig) *Executor {
	return &Executor{
		clientID:      cfg.ClientID,
		clientSecret:  cfg.ClientSecret,
		redirectURI:   cfg.RedirectURI,
		broadcasterID: cfg.BroadcasterID,
		accessToken:   cfg.AccessToken,
		refreshToken:  cfg.RefreshToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		log:     logger.Component("twitch"),
		enabled: cfg.Enabled,
	}
}

// Name returns the executor name
func (e *Executor) Name() string {
	return "twitch"
}

// SupportedActions returns the list of supported actions
func (e *Executor) SupportedActions() []string {
	return []string{
		"twitch.clip",
		"twitch.title",
		"twitch.category",
		"twitch.ban",
		"twitch.timeout",
		"twitch.unban",
	}
}

// CanHandle returns true if this executor can handle the action
func (e *Executor) CanHandle(action string) bool {
	return strings.HasPrefix(action, "twitch.")
}

// Execute executes a Twitch action
func (e *Executor) Execute(ctx context.Context, action llm.Action) (executor.Result, error) {
	if !e.enabled {
		return executor.NewErrorResult(fmt.Errorf("Twitch is not enabled")), nil
	}

	switch action.Action {
	case "twitch.clip":
		return e.createClip(ctx, action)
	case "twitch.title":
		return e.setTitle(ctx, action)
	case "twitch.category":
		return e.setCategory(ctx, action)
	case "twitch.ban":
		return e.banUser(ctx, action)
	case "twitch.timeout":
		return e.timeoutUser(ctx, action)
	case "twitch.unban":
		return e.unbanUser(ctx, action)
	default:
		return executor.NewErrorResult(fmt.Errorf("unknown Twitch action: %s", action.Action)), nil
	}
}

// IsAvailable checks if Twitch is available
func (e *Executor) IsAvailable() bool {
	return e.enabled && e.accessToken != ""
}

// Close releases resources
func (e *Executor) Close() error {
	return nil
}

// createClip creates a clip
func (e *Executor) createClip(ctx context.Context, action llm.Action) (executor.Result, error) {
	e.log.Info().Msg("Creating clip")

	// POST /clips?broadcaster_id=xxx
	params := url.Values{}
	params.Set("broadcaster_id", e.broadcasterID)

	resp, err := e.apiRequest(ctx, "POST", "/clips?"+params.Encode(), nil)
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	var result struct {
		Data []struct {
			ID      string `json:"id"`
			EditURL string `json:"edit_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return executor.NewErrorResult(err), err
	}

	if len(result.Data) == 0 {
		return executor.NewErrorResult(fmt.Errorf("no clip created")), nil
	}

	return executor.NewResultWithData("Clip created", map[string]interface{}{
		"clip_id":  result.Data[0].ID,
		"edit_url": result.Data[0].EditURL,
	}), nil
}

// setTitle sets the stream title
func (e *Executor) setTitle(ctx context.Context, action llm.Action) (executor.Result, error) {
	title := action.GetStringParam("title")
	if title == "" {
		return executor.NewErrorResult(fmt.Errorf("title is required")), nil
	}

	e.log.Info().Str("title", title).Msg("Setting stream title")

	// PATCH /channels?broadcaster_id=xxx
	body := map[string]string{"title": title}
	jsonBody, _ := json.Marshal(body)

	params := url.Values{}
	params.Set("broadcaster_id", e.broadcasterID)

	_, err := e.apiRequest(ctx, "PATCH", "/channels?"+params.Encode(), jsonBody)
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	return executor.NewResult("Title updated to: " + title), nil
}

// setCategory sets the stream category
func (e *Executor) setCategory(ctx context.Context, action llm.Action) (executor.Result, error) {
	category := action.GetStringParam("category")
	if category == "" {
		return executor.NewErrorResult(fmt.Errorf("category is required")), nil
	}

	e.log.Info().Str("category", category).Msg("Setting stream category")

	// First, search for the game/category
	gameID, err := e.searchCategory(ctx, category)
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	// PATCH /channels?broadcaster_id=xxx
	body := map[string]string{"game_id": gameID}
	jsonBody, _ := json.Marshal(body)

	params := url.Values{}
	params.Set("broadcaster_id", e.broadcasterID)

	_, err = e.apiRequest(ctx, "PATCH", "/channels?"+params.Encode(), jsonBody)
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	return executor.NewResult("Category updated to: " + category), nil
}

// searchCategory searches for a category and returns its ID
func (e *Executor) searchCategory(ctx context.Context, name string) (string, error) {
	params := url.Values{}
	params.Set("query", name)

	resp, err := e.apiRequest(ctx, "GET", "/search/categories?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}

	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", err
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("category not found: %s", name)
	}

	return result.Data[0].ID, nil
}

// banUser bans a user
func (e *Executor) banUser(ctx context.Context, action llm.Action) (executor.Result, error) {
	user := action.GetStringParam("user")
	if user == "" {
		return executor.NewErrorResult(fmt.Errorf("user is required")), nil
	}
	reason := action.GetStringParam("reason")

	e.log.Info().Str("user", user).Str("reason", reason).Msg("Banning user")

	// Get user ID
	userID, err := e.getUserID(ctx, user)
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	// POST /moderation/bans
	params := url.Values{}
	params.Set("broadcaster_id", e.broadcasterID)
	params.Set("moderator_id", e.broadcasterID)

	body := map[string]interface{}{
		"data": map[string]interface{}{
			"user_id": userID,
			"reason":  reason,
		},
	}
	jsonBody, _ := json.Marshal(body)

	_, err = e.apiRequest(ctx, "POST", "/moderation/bans?"+params.Encode(), jsonBody)
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	return executor.NewResult("User banned: " + user), nil
}

// timeoutUser gives a user a timeout
func (e *Executor) timeoutUser(ctx context.Context, action llm.Action) (executor.Result, error) {
	user := action.GetStringParam("user")
	if user == "" {
		return executor.NewErrorResult(fmt.Errorf("user is required")), nil
	}
	duration := action.GetIntParam("duration")
	if duration <= 0 {
		duration = 600 // Default 10 minutes
	}

	e.log.Info().Str("user", user).Int("duration", duration).Msg("Timing out user")

	// Get user ID
	userID, err := e.getUserID(ctx, user)
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	// POST /moderation/bans
	params := url.Values{}
	params.Set("broadcaster_id", e.broadcasterID)
	params.Set("moderator_id", e.broadcasterID)

	body := map[string]interface{}{
		"data": map[string]interface{}{
			"user_id":  userID,
			"duration": duration,
		},
	}
	jsonBody, _ := json.Marshal(body)

	_, err = e.apiRequest(ctx, "POST", "/moderation/bans?"+params.Encode(), jsonBody)
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	return executor.NewResult(fmt.Sprintf("User %s timed out for %d seconds", user, duration)), nil
}

// unbanUser unbans a user
func (e *Executor) unbanUser(ctx context.Context, action llm.Action) (executor.Result, error) {
	user := action.GetStringParam("user")
	if user == "" {
		return executor.NewErrorResult(fmt.Errorf("user is required")), nil
	}

	e.log.Info().Str("user", user).Msg("Unbanning user")

	// Get user ID
	userID, err := e.getUserID(ctx, user)
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	// DELETE /moderation/bans
	params := url.Values{}
	params.Set("broadcaster_id", e.broadcasterID)
	params.Set("moderator_id", e.broadcasterID)
	params.Set("user_id", userID)

	_, err = e.apiRequest(ctx, "DELETE", "/moderation/bans?"+params.Encode(), nil)
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	return executor.NewResult("User unbanned: " + user), nil
}

// getUserID gets a user's ID from their username
func (e *Executor) getUserID(ctx context.Context, username string) (string, error) {
	params := url.Values{}
	params.Set("login", username)

	resp, err := e.apiRequest(ctx, "GET", "/users?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}

	var result struct {
		Data []struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", err
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("user not found: %s", username)
	}

	return result.Data[0].ID, nil
}

// apiRequest makes an authenticated request to the Twitch API
func (e *Executor) apiRequest(ctx context.Context, method, endpoint string, body []byte) ([]byte, error) {
	reqURL := twitchAPIURL + endpoint

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, reqURL, strings.NewReader(string(body)))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, reqURL, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+e.accessToken)
	req.Header.Set("Client-Id", e.clientID)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		// Try to refresh token
		if err := e.refreshAccessToken(ctx); err != nil {
			return nil, fmt.Errorf("unauthorized and failed to refresh token: %w", err)
		}
		// Retry request
		return e.apiRequest(ctx, method, endpoint, body)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// refreshAccessToken refreshes the access token
func (e *Executor) refreshAccessToken(ctx context.Context) error {
	if e.refreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", e.refreshToken)
	data.Set("client_id", e.clientID)
	data.Set("client_secret", e.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", twitchAuthURL+"/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token refresh failed: %s", string(body))
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	e.accessToken = result.AccessToken
	e.refreshToken = result.RefreshToken

	e.log.Info().Msg("Access token refreshed")
	return nil
}

// SetTokens sets the OAuth tokens
func (e *Executor) SetTokens(accessToken, refreshToken string) {
	e.accessToken = accessToken
	e.refreshToken = refreshToken
}

// GetTokens returns the current tokens
func (e *Executor) GetTokens() (accessToken, refreshToken string) {
	return e.accessToken, e.refreshToken
}
