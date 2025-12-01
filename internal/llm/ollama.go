package llm

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
	"github.com/anastreamer/ana/pkg/logger"
	"github.com/anastreamer/ana/pkg/utils"
	"github.com/rs/zerolog"
)

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	url          string
	model        string
	timeout      time.Duration
	client       *http.Client
	log          zerolog.Logger
	streamerName string
}

// OllamaRequest represents a request to the Ollama API
type OllamaRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	System  string         `json:"system,omitempty"`
	Stream  bool           `json:"stream"`
	Format  string         `json:"format,omitempty"`
	Options *OllamaOptions `json:"options,omitempty"`
}

// OllamaOptions represents Ollama generation options
type OllamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

// OllamaResponse represents a response from the Ollama API
type OllamaResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	Context   []int  `json:"context,omitempty"`
	CreatedAt string `json:"created_at"`
}

// OllamaTagsResponse represents the response from /api/tags
type OllamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(cfg config.OllamaConfig, streamerName string) (*OllamaProvider, error) {
	timeout := cfg.Timeout()
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &OllamaProvider{
		url:          strings.TrimSuffix(cfg.URL, "/"),
		model:        cfg.Model,
		timeout:      timeout,
		client: &http.Client{
			Timeout: timeout,
		},
		log:          logger.Component("ollama"),
		streamerName: streamerName,
	}, nil
}

// Name returns the provider name
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// Complete sends a prompt to Ollama and returns an Action
func (p *OllamaProvider) Complete(ctx context.Context, prompt string) (Action, error) {
	p.log.Debug().Str("prompt", prompt).Msg("Sending prompt to Ollama")

	// Create request
	reqBody := OllamaRequest{
		Model:  p.model,
		Prompt: prompt,
		System: GetSystemPromptWithStreamer(p.streamerName),
		Stream: false,
		Format: "json", // Force JSON output
		Options: &OllamaOptions{
			Temperature: 0.3,
			NumPredict:  500,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return Action{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.url+"/api/generate", bytes.NewReader(jsonBody))
	if err != nil {
		return Action{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return Action{}, fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Action{}, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return Action{}, fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse Ollama response
	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return Action{}, fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	p.log.Debug().Str("response", ollamaResp.Response).Msg("Received response from Ollama")

	// Parse the action from the response
	action, err := p.parseAction(ollamaResp.Response)
	if err != nil {
		p.log.Warn().Err(err).Str("raw_response", ollamaResp.Response).Msg("Failed to parse action, returning fallback")
		return Action{
			Action: "none",
			Params: map[string]interface{}{},
			Reply:  "Lo siento, no pude entender tu solicitud. ¿Puedes repetirlo?",
		}, nil
	}

	return action, nil
}

// CompleteRaw sends a prompt and returns the raw response
func (p *OllamaProvider) CompleteRaw(ctx context.Context, prompt string) (string, error) {
	reqBody := OllamaRequest{
		Model:  p.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.url+"/api/generate", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return ollamaResp.Response, nil
}

// IsAvailable checks if Ollama is available
func (p *OllamaProvider) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", p.url+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	// Check if our model is available
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	var tagsResp OllamaTagsResponse
	if err := json.Unmarshal(body, &tagsResp); err != nil {
		return false
	}

	for _, model := range tagsResp.Models {
		if model.Name == p.model || strings.HasPrefix(model.Name, p.model+":") {
			return true
		}
	}

	p.log.Warn().Str("model", p.model).Msg("Model not found in Ollama")
	return false
}

// Close releases resources
func (p *OllamaProvider) Close() error {
	return nil
}

// parseAction parses the LLM response into an Action
func (p *OllamaProvider) parseAction(response string) (Action, error) {
	// Extract JSON from response
	jsonStr := utils.ExtractJSON(response)

	var action Action
	if err := json.Unmarshal([]byte(jsonStr), &action); err != nil {
		return Action{}, fmt.Errorf("failed to parse action JSON: %w", err)
	}

	// Initialize params map if nil
	if action.Params == nil {
		action.Params = make(map[string]interface{})
	}

	// Validate action
	if action.Action == "" {
		return Action{}, fmt.Errorf("action field is empty")
	}

	return action, nil
}

// SetModel changes the model being used
func (p *OllamaProvider) SetModel(model string) {
	p.model = model
}

// GetModel returns the current model
func (p *OllamaProvider) GetModel() string {
	return p.model
}
