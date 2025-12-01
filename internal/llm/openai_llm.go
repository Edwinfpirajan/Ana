package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anastreamer/ana/internal/config"
	"github.com/anastreamer/ana/pkg/logger"
	"github.com/anastreamer/ana/pkg/utils"
	"github.com/rs/zerolog"
)

const openAIBaseURL = "https://api.openai.com/v1"

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	apiKey       string
	model        string
	temperature  float64
	client       *http.Client
	log          zerolog.Logger
	streamerName string
}

// OpenAIChatRequest represents a chat completion request
type OpenAIChatRequest struct {
	Model          string          `json:"model"`
	Messages       []OpenAIMessage `json:"messages"`
	Temperature    float64         `json:"temperature,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

// ResponseFormat specifies the output format
type ResponseFormat struct {
	Type string `json:"type"` // "json_object" for JSON mode
}

// OpenAIMessage represents a message in the conversation
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIChatResponse represents a chat completion response
type OpenAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *OpenAIError `json:"error,omitempty"`
}

// OpenAIError represents an error from the OpenAI API
type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(cfg config.OpenAILLMConfig, streamerName string) (*OpenAIProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	temperature := cfg.Temperature
	if temperature == 0 {
		temperature = 0.3
	}

	model := cfg.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	return &OpenAIProvider{
		apiKey:       cfg.APIKey,
		model:        model,
		temperature:  temperature,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		log:          logger.Component("openai-llm"),
		streamerName: streamerName,
	}, nil
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Complete sends a prompt to OpenAI and returns an Action
func (p *OpenAIProvider) Complete(ctx context.Context, prompt string) (Action, error) {
	p.log.Debug().Str("prompt", prompt).Msg("Sending prompt to OpenAI")

	// Create request
	reqBody := OpenAIChatRequest{
		Model: p.model,
		Messages: []OpenAIMessage{
			{
				Role:    "system",
				Content: GetSystemPromptWithStreamer(p.streamerName),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: p.temperature,
		MaxTokens:   500,
		ResponseFormat: &ResponseFormat{
			Type: "json_object",
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return Action{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", openAIBaseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return Action{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return Action{}, fmt.Errorf("failed to send request to OpenAI: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Action{}, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var openAIResp OpenAIChatResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return Action{}, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	// Check for API error
	if openAIResp.Error != nil {
		return Action{}, fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return Action{}, fmt.Errorf("OpenAI returned status %d: %s", resp.StatusCode, string(body))
	}

	// Check for valid response
	if len(openAIResp.Choices) == 0 {
		return Action{}, fmt.Errorf("no choices in OpenAI response")
	}

	// Validate choice has valid message
	if openAIResp.Choices[0].Message.Content == "" {
		return Action{}, fmt.Errorf("empty content in OpenAI response")
	}

	content := openAIResp.Choices[0].Message.Content
	p.log.Debug().Str("response", content).Msg("Received response from OpenAI")

	// Parse the action from the response
	action, err := p.parseAction(content)
	if err != nil {
		p.log.Warn().Err(err).Str("raw_response", content).Msg("Failed to parse action, returning fallback")
		return Action{
			Action: "none",
			Params: map[string]interface{}{},
			Reply:  "Lo siento, no pude entender tu solicitud. ¿Puedes repetirlo?",
		}, nil
	}

	return action, nil
}

// CompleteRaw sends a prompt and returns the raw response
func (p *OpenAIProvider) CompleteRaw(ctx context.Context, prompt string) (string, error) {
	reqBody := OpenAIChatRequest{
		Model: p.model,
		Messages: []OpenAIMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: p.temperature,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openAIBaseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var openAIResp OpenAIChatResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if openAIResp.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	// Validate choice has valid message
	if openAIResp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("empty content in OpenAI response")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

// IsAvailable checks if OpenAI is available
func (p *OpenAIProvider) IsAvailable(ctx context.Context) bool {
	// Simple check - try to list models
	req, err := http.NewRequestWithContext(ctx, "GET", openAIBaseURL+"/models", nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// Close releases resources
func (p *OpenAIProvider) Close() error {
	return nil
}

// parseAction parses the LLM response into an Action
func (p *OpenAIProvider) parseAction(response string) (Action, error) {
	// Extract JSON from response (in case there's extra text)
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
func (p *OpenAIProvider) SetModel(model string) {
	p.model = model
}

// GetModel returns the current model
func (p *OpenAIProvider) GetModel() string {
	return p.model
}
