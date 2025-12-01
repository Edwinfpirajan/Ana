// Package obs provides OBS Studio integration for AnaStreamer
package obs

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/anastreamer/ana/internal/config"
	"github.com/anastreamer/ana/internal/executor"
	"github.com/anastreamer/ana/internal/llm"
	"github.com/anastreamer/ana/pkg/logger"
	"github.com/rs/zerolog"
)

// Executor implements the OBS action executor
type Executor struct {
	url       string
	password  string
	conn      *websocket.Conn
	log       zerolog.Logger
	enabled   bool
	connected bool

	mu          sync.Mutex
	requestID   atomic.Int64
	responses   map[int64]chan json.RawMessage
	responsesMu sync.RWMutex
}

// OBS WebSocket message types
type (
	// Message is the base message structure
	Message struct {
		Op int             `json:"op"`
		D  json.RawMessage `json:"d"`
	}

	// HelloData is sent by OBS on connection
	HelloData struct {
		ObsWebSocketVersion string `json:"obsWebSocketVersion"`
		RPCVersion          int    `json:"rpcVersion"`
		Authentication      *struct {
			Challenge string `json:"challenge"`
			Salt      string `json:"salt"`
		} `json:"authentication,omitempty"`
	}

	// IdentifyData is sent to authenticate
	IdentifyData struct {
		RPCVersion     int    `json:"rpcVersion"`
		Authentication string `json:"authentication,omitempty"`
	}

	// RequestData is the structure for requests
	RequestData struct {
		RequestType string                 `json:"requestType"`
		RequestID   string                 `json:"requestId"`
		RequestData map[string]interface{} `json:"requestData,omitempty"`
	}

	// ResponseData is the structure for responses
	ResponseData struct {
		RequestType   string                 `json:"requestType"`
		RequestID     string                 `json:"requestId"`
		RequestStatus RequestStatus          `json:"requestStatus"`
		ResponseData  map[string]interface{} `json:"responseData,omitempty"`
	}

	// RequestStatus contains the status of a request
	RequestStatus struct {
		Result  bool   `json:"result"`
		Code    int    `json:"code"`
		Comment string `json:"comment,omitempty"`
	}
)

// OBS WebSocket opcodes
const (
	OpHello        = 0
	OpIdentify     = 1
	OpIdentified   = 2
	OpReidentify   = 3
	OpEvent        = 5
	OpRequest      = 6
	OpRequestResp  = 7
	OpRequestBatch = 8
)

// NewExecutor creates a new OBS executor
func NewExecutor(cfg config.OBSConfig) *Executor {
	return &Executor{
		url:       cfg.URL,
		password:  cfg.Password,
		log:       logger.Component("obs"),
		enabled:   cfg.Enabled,
		responses: make(map[int64]chan json.RawMessage),
	}
}

// Name returns the executor name
func (e *Executor) Name() string {
	return "obs"
}

// SupportedActions returns the list of supported actions
func (e *Executor) SupportedActions() []string {
	return []string{
		"obs.scene",
		"obs.source.show",
		"obs.source.hide",
		"obs.volume",
		"obs.mute",
		"obs.unmute",
		"obs.text",
	}
}

// CanHandle returns true if this executor can handle the action
func (e *Executor) CanHandle(action string) bool {
	return strings.HasPrefix(action, "obs.")
}

// Connect establishes connection to OBS WebSocket
func (e *Executor) Connect(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.connected {
		return nil
	}

	e.log.Info().Str("url", e.url).Msg("Connecting to OBS")

	// Connect to WebSocket
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, e.url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to OBS: %w", err)
	}
	e.conn = conn

	// Read Hello message
	var msg Message
	if err := conn.ReadJSON(&msg); err != nil {
		conn.Close()
		return fmt.Errorf("failed to read hello: %w", err)
	}

	if msg.Op != OpHello {
		conn.Close()
		return fmt.Errorf("expected Hello, got op %d", msg.Op)
	}

	var hello HelloData
	if err := json.Unmarshal(msg.D, &hello); err != nil {
		conn.Close()
		return fmt.Errorf("failed to parse hello: %w", err)
	}

	e.log.Debug().
		Str("version", hello.ObsWebSocketVersion).
		Int("rpc", hello.RPCVersion).
		Msg("Received Hello from OBS")

	// Send Identify
	identify := IdentifyData{
		RPCVersion: 1,
	}

	// Generate authentication if required
	if hello.Authentication != nil && e.password != "" {
		auth := generateAuth(e.password, hello.Authentication.Salt, hello.Authentication.Challenge)
		identify.Authentication = auth
	}

	identifyMsg := Message{
		Op: OpIdentify,
	}
	identifyMsg.D, _ = json.Marshal(identify)

	if err := conn.WriteJSON(identifyMsg); err != nil {
		conn.Close()
		return fmt.Errorf("failed to send identify: %w", err)
	}

	// Read Identified response
	if err := conn.ReadJSON(&msg); err != nil {
		conn.Close()
		return fmt.Errorf("failed to read identified: %w", err)
	}

	if msg.Op != OpIdentified {
		conn.Close()
		return fmt.Errorf("authentication failed, got op %d", msg.Op)
	}

	e.connected = true
	e.log.Info().Msg("Connected to OBS")

	// Start message reader
	go e.readMessages()

	return nil
}

// readMessages reads incoming messages from OBS
func (e *Executor) readMessages() {
	for {
		var msg Message
		if err := e.conn.ReadJSON(&msg); err != nil {
			e.log.Error().Err(err).Msg("Error reading from OBS")
			e.mu.Lock()
			e.connected = false
			e.mu.Unlock()
			return
		}

		switch msg.Op {
		case OpRequestResp:
			var resp ResponseData
			if err := json.Unmarshal(msg.D, &resp); err != nil {
				e.log.Error().Err(err).Msg("Failed to parse response")
				continue
			}

			// Parse request ID
			var reqID int64
			fmt.Sscanf(resp.RequestID, "%d", &reqID)

			e.responsesMu.RLock()
			ch, ok := e.responses[reqID]
			e.responsesMu.RUnlock()

			if ok {
				ch <- msg.D
			}

		case OpEvent:
			// Handle events if needed
		}
	}
}

// generateAuth generates the authentication string
func generateAuth(password, salt, challenge string) string {
	// Base64(SHA256(password + salt))
	secretHash := sha256.Sum256([]byte(password + salt))
	secret := base64.StdEncoding.EncodeToString(secretHash[:])

	// Base64(SHA256(secret + challenge))
	authHash := sha256.Sum256([]byte(secret + challenge))
	return base64.StdEncoding.EncodeToString(authHash[:])
}

// Execute executes an OBS action
func (e *Executor) Execute(ctx context.Context, action llm.Action) (executor.Result, error) {
	if !e.enabled {
		return executor.NewErrorResult(fmt.Errorf("OBS is not enabled")), nil
	}

	if !e.connected {
		if err := e.Connect(ctx); err != nil {
			return executor.NewErrorResult(err), err
		}
	}

	switch action.Action {
	case "obs.scene":
		return e.setScene(ctx, action)
	case "obs.source.show":
		return e.setSourceVisibility(ctx, action, true)
	case "obs.source.hide":
		return e.setSourceVisibility(ctx, action, false)
	case "obs.volume":
		return e.setVolume(ctx, action)
	case "obs.mute":
		return e.setMute(ctx, action, true)
	case "obs.unmute":
		return e.setMute(ctx, action, false)
	case "obs.text":
		return e.setText(ctx, action)
	default:
		return executor.NewErrorResult(fmt.Errorf("unknown OBS action: %s", action.Action)), nil
	}
}

// sendRequest sends a request to OBS and waits for response
func (e *Executor) sendRequest(ctx context.Context, requestType string, data map[string]interface{}) (*ResponseData, error) {
	e.mu.Lock()
	if !e.connected {
		e.mu.Unlock()
		return nil, fmt.Errorf("not connected to OBS")
	}
	e.mu.Unlock()

	reqID := e.requestID.Add(1)
	respCh := make(chan json.RawMessage, 1)

	e.responsesMu.Lock()
	e.responses[reqID] = respCh
	e.responsesMu.Unlock()

	defer func() {
		e.responsesMu.Lock()
		delete(e.responses, reqID)
		e.responsesMu.Unlock()
	}()

	request := RequestData{
		RequestType: requestType,
		RequestID:   fmt.Sprintf("%d", reqID),
		RequestData: data,
	}

	msg := Message{Op: OpRequest}
	msg.D, _ = json.Marshal(request)

	e.mu.Lock()
	err := e.conn.WriteJSON(msg)
	e.mu.Unlock()

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response
	select {
	case respData := <-respCh:
		var resp ResponseData
		if err := json.Unmarshal(respData, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		return &resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("request timeout")
	}
}

// setScene changes the current scene
func (e *Executor) setScene(ctx context.Context, action llm.Action) (executor.Result, error) {
	scene := action.GetStringParam("scene")
	if scene == "" {
		return executor.NewErrorResult(fmt.Errorf("scene name is required")), nil
	}

	e.log.Info().Str("scene", scene).Msg("Changing scene")

	resp, err := e.sendRequest(ctx, "SetCurrentProgramScene", map[string]interface{}{
		"sceneName": scene,
	})
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	if !resp.RequestStatus.Result {
		return executor.NewErrorResult(fmt.Errorf("failed to change scene: %s", resp.RequestStatus.Comment)), nil
	}

	return executor.NewResult("Changed to scene: " + scene), nil
}

// setSourceVisibility shows or hides a source
func (e *Executor) setSourceVisibility(ctx context.Context, action llm.Action, visible bool) (executor.Result, error) {
	source := action.GetStringParam("source")
	if source == "" {
		return executor.NewErrorResult(fmt.Errorf("source name is required")), nil
	}

	// Get current scene
	sceneResp, err := e.sendRequest(ctx, "GetCurrentProgramScene", nil)
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	sceneName, ok := sceneResp.ResponseData["currentProgramSceneName"].(string)
	if !ok {
		return executor.NewErrorResult(fmt.Errorf("could not get current scene")), nil
	}

	// Get scene item ID
	itemResp, err := e.sendRequest(ctx, "GetSceneItemId", map[string]interface{}{
		"sceneName":  sceneName,
		"sourceName": source,
	})
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	if !itemResp.RequestStatus.Result {
		return executor.NewErrorResult(fmt.Errorf("source not found: %s", source)), nil
	}

	sceneItemId, ok := itemResp.ResponseData["sceneItemId"].(float64)
	if !ok {
		return executor.NewErrorResult(fmt.Errorf("could not get scene item ID")), nil
	}

	// Set visibility
	e.log.Info().Str("source", source).Bool("visible", visible).Msg("Setting source visibility")

	resp, err := e.sendRequest(ctx, "SetSceneItemEnabled", map[string]interface{}{
		"sceneName":        sceneName,
		"sceneItemId":      int(sceneItemId),
		"sceneItemEnabled": visible,
	})
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	if !resp.RequestStatus.Result {
		return executor.NewErrorResult(fmt.Errorf("failed to set visibility: %s", resp.RequestStatus.Comment)), nil
	}

	action_str := "shown"
	if !visible {
		action_str = "hidden"
	}
	return executor.NewResult(fmt.Sprintf("Source %s %s", source, action_str)), nil
}

// setVolume changes the volume of a source
func (e *Executor) setVolume(ctx context.Context, action llm.Action) (executor.Result, error) {
	source := action.GetStringParam("source")
	if source == "" {
		return executor.NewErrorResult(fmt.Errorf("source name is required")), nil
	}

	volume := action.GetFloatParam("volume")
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}

	// Convert to dB (0-1 linear to dB scale)
	// OBS uses dB, roughly: dB = 20 * log10(linear)
	var volumeDb float64
	if volume <= 0 {
		volumeDb = -100 // Effectively muted
	} else {
		volumeDb = 20 * (volume - 1) * 2 // Simplified mapping
	}

	e.log.Info().Str("source", source).Float64("volume", volume).Msg("Setting volume")

	resp, err := e.sendRequest(ctx, "SetInputVolume", map[string]interface{}{
		"inputName":     source,
		"inputVolumeDb": volumeDb,
	})
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	if !resp.RequestStatus.Result {
		return executor.NewErrorResult(fmt.Errorf("failed to set volume: %s", resp.RequestStatus.Comment)), nil
	}

	return executor.NewResult(fmt.Sprintf("Volume of %s set to %.0f%%", source, volume*100)), nil
}

// setMute mutes or unmutes a source
func (e *Executor) setMute(ctx context.Context, action llm.Action, muted bool) (executor.Result, error) {
	source := action.GetStringParam("source")
	if source == "" {
		return executor.NewErrorResult(fmt.Errorf("source name is required")), nil
	}

	e.log.Info().Str("source", source).Bool("muted", muted).Msg("Setting mute state")

	resp, err := e.sendRequest(ctx, "SetInputMute", map[string]interface{}{
		"inputName":  source,
		"inputMuted": muted,
	})
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	if !resp.RequestStatus.Result {
		return executor.NewErrorResult(fmt.Errorf("failed to set mute: %s", resp.RequestStatus.Comment)), nil
	}

	action_str := "muted"
	if !muted {
		action_str = "unmuted"
	}
	return executor.NewResult(fmt.Sprintf("Source %s %s", source, action_str)), nil
}

// setText changes the text of a text source
func (e *Executor) setText(ctx context.Context, action llm.Action) (executor.Result, error) {
	source := action.GetStringParam("source")
	if source == "" {
		return executor.NewErrorResult(fmt.Errorf("source name is required")), nil
	}

	text := action.GetStringParam("text")

	e.log.Info().Str("source", source).Str("text", text).Msg("Setting text")

	resp, err := e.sendRequest(ctx, "SetInputSettings", map[string]interface{}{
		"inputName": source,
		"inputSettings": map[string]interface{}{
			"text": text,
		},
	})
	if err != nil {
		return executor.NewErrorResult(err), err
	}

	if !resp.RequestStatus.Result {
		return executor.NewErrorResult(fmt.Errorf("failed to set text: %s", resp.RequestStatus.Comment)), nil
	}

	return executor.NewResult("Text updated"), nil
}

// IsAvailable checks if OBS is available
func (e *Executor) IsAvailable() bool {
	return e.enabled && e.connected
}

// Close releases resources
func (e *Executor) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.conn != nil {
		e.conn.Close()
		e.conn = nil
	}
	e.connected = false
	return nil
}
