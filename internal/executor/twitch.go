package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/anastreamer/ana/internal/config"
	"github.com/anastreamer/ana/internal/llm"
	"github.com/anastreamer/ana/pkg/logger"
	"github.com/rs/zerolog"
)

const (
	twitchAPIBaseURL = "https://api.twitch.tv/helix"
	twitchValidateURL = "https://id.twitch.tv/oauth2/validate"
)

// TwitchExecutor handles Twitch integration via Helix API
type TwitchExecutor struct {
	cfg        config.TwitchConfig
	httpClient *http.Client
	log        zerolog.Logger
	connected  bool
	username   string
}

// NewTwitchExecutor creates a new Twitch executor
func NewTwitchExecutor(cfg config.TwitchConfig) (*TwitchExecutor, error) {
	exec := &TwitchExecutor{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: logger.Component("twitch-executor"),
	}

	if cfg.Enabled && cfg.AccessToken != "" {
		if err := exec.validateToken(); err != nil {
			exec.log.Warn().Err(err).Msg("Failed to validate Twitch token, executor will be unavailable")
		} else {
			exec.connected = true
		}
	}

	return exec, nil
}

// validateToken validates the access token and gets user info
func (e *TwitchExecutor) validateToken() error {
	e.log.Info().Msg("Validating Twitch access token")

	req, err := http.NewRequest("GET", twitchValidateURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("Authorization", "OAuth "+e.cfg.AccessToken)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token validation failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		ClientID string   `json:"client_id"`
		Login    string   `json:"login"`
		Scopes   []string `json:"scopes"`
		UserID   string   `json:"user_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode validation response: %w", err)
	}

	e.username = result.Login
	e.log.Info().
		Str("username", result.Login).
		Str("user_id", result.UserID).
		Strs("scopes", result.Scopes).
		Msg("Successfully validated Twitch token")

	return nil
}

// Name returns the executor name
func (e *TwitchExecutor) Name() string {
	return "twitch"
}

// SupportedActions returns all supported Twitch actions
func (e *TwitchExecutor) SupportedActions() []string {
	return []string{
		"twitch.clip",
		"twitch.title",
		"twitch.category",
	}
}

// CanHandle checks if this executor can handle the given action
func (e *TwitchExecutor) CanHandle(action string) bool {
	return strings.HasPrefix(action, "twitch.")
}

// Execute executes a Twitch action
func (e *TwitchExecutor) Execute(ctx context.Context, action llm.Action) (Result, error) {
	if !e.IsAvailable() {
		return NewErrorResult(fmt.Errorf("Twitch not connected")), fmt.Errorf("Twitch not connected")
	}

	e.log.Debug().Str("action", action.Action).Msg("Executing Twitch action")

	switch action.Action {
	case "twitch.clip":
		return e.createClip(ctx, action)
	case "twitch.title":
		return e.updateTitle(ctx, action)
	case "twitch.category":
		return e.updateCategory(ctx, action)
	default:
		err := fmt.Errorf("unknown action: %s", action.Action)
		return NewErrorResult(err), err
	}
}

// createClip creates a Twitch clip
func (e *TwitchExecutor) createClip(ctx context.Context, action llm.Action) (Result, error) {
	e.log.Info().Msg("Creating Twitch clip")

	url := fmt.Sprintf("%s/clips?broadcaster_id=%s", twitchAPIBaseURL, e.cfg.BroadcasterID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return NewErrorResult(err), err
	}

	e.setHeaders(req)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return NewErrorResult(err), err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("failed to create clip (status %d): %s", resp.StatusCode, string(body))
		return NewErrorResult(err), err
	}

	var result struct {
		Data []struct {
			ID      string `json:"id"`
			EditURL string `json:"edit_url"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return NewErrorResult(err), err
	}

	if len(result.Data) == 0 {
		err := fmt.Errorf("no clip created")
		return NewErrorResult(err), err
	}

	e.log.Info().Str("clip_id", result.Data[0].ID).Msg("Clip created successfully")

	return NewResultWithData("Clip creado", map[string]interface{}{
		"clip_id":  result.Data[0].ID,
		"edit_url": result.Data[0].EditURL,
	}), nil
}

// updateTitle updates the stream title
func (e *TwitchExecutor) updateTitle(ctx context.Context, action llm.Action) (Result, error) {
	title, ok := action.Params["title"].(string)
	if !ok || title == "" {
		err := fmt.Errorf("title parameter is required")
		return NewErrorResult(err), err
	}

	e.log.Info().Str("title", title).Msg("Updating stream title")

	url := fmt.Sprintf("%s/channels?broadcaster_id=%s", twitchAPIBaseURL, e.cfg.BroadcasterID)

	body := map[string]interface{}{
		"title": title,
	}

	return e.patchChannel(ctx, url, body, "Título actualizado")
}

// updateCategory updates the stream category/game
func (e *TwitchExecutor) updateCategory(ctx context.Context, action llm.Action) (Result, error) {
	category, ok := action.Params["category"].(string)
	if !ok || category == "" {
		err := fmt.Errorf("category parameter is required")
		return NewErrorResult(err), err
	}

	e.log.Info().Str("category", category).Msg("Updating stream category")

	// First, search for the game/category
	gameID, err := e.searchGame(ctx, category)
	if err != nil {
		return NewErrorResult(err), err
	}

	url := fmt.Sprintf("%s/channels?broadcaster_id=%s", twitchAPIBaseURL, e.cfg.BroadcasterID)

	body := map[string]interface{}{
		"game_id": gameID,
	}

	return e.patchChannel(ctx, url, body, "Categoría actualizada")
}

// searchGame searches for a game/category by name
func (e *TwitchExecutor) searchGame(ctx context.Context, name string) (string, error) {
	url := fmt.Sprintf("%s/games?name=%s", twitchAPIBaseURL, name)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	e.setHeaders(req)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to search game (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("game/category not found: %s", name)
	}

	return result.Data[0].ID, nil
}

// patchChannel makes a PATCH request to update channel info
func (e *TwitchExecutor) patchChannel(ctx context.Context, url string, body map[string]interface{}, successMsg string) (Result, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return NewErrorResult(err), err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return NewErrorResult(err), err
	}

	e.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return NewErrorResult(err), err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("failed to update channel (status %d): %s", resp.StatusCode, string(body))
		return NewErrorResult(err), err
	}

	e.log.Info().Msg("Channel updated successfully")
	return NewResult(successMsg), nil
}

// setHeaders sets the required Twitch API headers
func (e *TwitchExecutor) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+e.cfg.AccessToken)
	req.Header.Set("Client-Id", e.cfg.ClientID)
}

// IsAvailable checks if the executor is ready to handle actions
func (e *TwitchExecutor) IsAvailable() bool {
	return e.connected && e.cfg.Enabled && e.cfg.AccessToken != ""
}

// Close releases any resources
func (e *TwitchExecutor) Close() error {
	e.log.Info().Msg("Closing Twitch executor")
	e.connected = false
	return nil
}

// GetStatus returns detailed status information for the status command
func (e *TwitchExecutor) GetStatus() string {
	if !e.IsAvailable() {
		return "Twitch: desconectado"
	}
	return fmt.Sprintf("Twitch: conectado como @%s", e.username)
}
