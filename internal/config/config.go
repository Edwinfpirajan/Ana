// Package config handles all configuration for AnaStreamer
package config

import "time"

// Config is the main configuration structure
type Config struct {
	General GeneralConfig `yaml:"general" mapstructure:"general"`
	Audio   AudioConfig   `yaml:"audio" mapstructure:"audio"`
	Hotkey  HotkeyConfig  `yaml:"hotkey" mapstructure:"hotkey"`
	STT     STTConfig     `yaml:"stt" mapstructure:"stt"`
	LLM     LLMConfig     `yaml:"llm" mapstructure:"llm"`
	TTS     TTSConfig     `yaml:"tts" mapstructure:"tts"`
	Twitch  TwitchConfig  `yaml:"twitch" mapstructure:"twitch"`
	OBS     OBSConfig     `yaml:"obs" mapstructure:"obs"`
	Music   MusicConfig   `yaml:"music" mapstructure:"music"`
	Sounds  SoundsConfig  `yaml:"sounds" mapstructure:"sounds"`
}

// GeneralConfig contains general application settings
type GeneralConfig struct {
	Language      string `yaml:"language" mapstructure:"language"`
	LogLevel      string `yaml:"log_level" mapstructure:"log_level"`
	DataDir       string `yaml:"data_dir" mapstructure:"data_dir"`
	StreamerName  string `yaml:"streamer_name" mapstructure:"streamer_name"`
}

// AudioConfig contains audio capture settings
type AudioConfig struct {
	Device     string       `yaml:"device" mapstructure:"device"`
	SampleRate int          `yaml:"sample_rate" mapstructure:"sample_rate"`
	Channels   int          `yaml:"channels" mapstructure:"channels"`
	ChunkSize  int          `yaml:"chunk_size" mapstructure:"chunk_size"`
	VAD        VADConfig    `yaml:"vad" mapstructure:"vad"`
	WakeWord   WakeWordConfig `yaml:"wake_word" mapstructure:"wake_word"`
}

// VADConfig contains Voice Activity Detection settings
type VADConfig struct {
	Enabled            bool    `yaml:"enabled" mapstructure:"enabled"`
	Sensitivity        float64 `yaml:"sensitivity" mapstructure:"sensitivity"`
	SilenceThresholdMs int     `yaml:"silence_threshold_ms" mapstructure:"silence_threshold_ms"`
	MinSpeechMs        int     `yaml:"min_speech_ms" mapstructure:"min_speech_ms"`
}

// WakeWordConfig contains wake word detection settings
type WakeWordConfig struct {
	Enabled   bool    `yaml:"enabled" mapstructure:"enabled"`
	Word      string  `yaml:"word" mapstructure:"word"`
	Threshold float64 `yaml:"threshold" mapstructure:"threshold"`
}

// HotkeyConfig contains hotkey settings
type HotkeyConfig struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Key     string `yaml:"key" mapstructure:"key"`
	Mode    string `yaml:"mode" mapstructure:"mode"` // "hold" or "toggle"
}

// STTConfig contains Speech-to-Text settings
type STTConfig struct {
	Provider string        `yaml:"provider" mapstructure:"provider"` // "whisper" or "openai"
	Whisper  WhisperConfig `yaml:"whisper" mapstructure:"whisper"`
	OpenAI   OpenAISTTConfig `yaml:"openai" mapstructure:"openai"`
}

// WhisperConfig contains local Whisper settings
type WhisperConfig struct {
	BinaryPath string `yaml:"binary_path" mapstructure:"binary_path"`
	ModelPath  string `yaml:"model_path" mapstructure:"model_path"`
	Language   string `yaml:"language" mapstructure:"language"`
}

// OpenAISTTConfig contains OpenAI Whisper API settings
type OpenAISTTConfig struct {
	APIKey string `yaml:"api_key" mapstructure:"api_key"`
	Model  string `yaml:"model" mapstructure:"model"`
}

// LLMConfig contains Language Model settings
type LLMConfig struct {
	Provider string       `yaml:"provider" mapstructure:"provider"` // "ollama" or "openai"
	Ollama   OllamaConfig `yaml:"ollama" mapstructure:"ollama"`
	OpenAI   OpenAILLMConfig `yaml:"openai" mapstructure:"openai"`
}

// OllamaConfig contains local Ollama settings
type OllamaConfig struct {
	URL            string `yaml:"url" mapstructure:"url"`
	Model          string `yaml:"model" mapstructure:"model"`
	TimeoutSeconds int    `yaml:"timeout_seconds" mapstructure:"timeout_seconds"`
}

// Timeout returns the timeout as a time.Duration
func (c OllamaConfig) Timeout() time.Duration {
	return time.Duration(c.TimeoutSeconds) * time.Second
}

// OpenAILLMConfig contains OpenAI GPT settings
type OpenAILLMConfig struct {
	APIKey      string  `yaml:"api_key" mapstructure:"api_key"`
	Model       string  `yaml:"model" mapstructure:"model"`
	Temperature float64 `yaml:"temperature" mapstructure:"temperature"`
}

// TTSConfig contains Text-to-Speech settings
type TTSConfig struct {
	Provider string        `yaml:"provider" mapstructure:"provider"` // "piper" or "openai"
	Piper    PiperConfig   `yaml:"piper" mapstructure:"piper"`
	OpenAI   OpenAITTSConfig `yaml:"openai" mapstructure:"openai"`
}

// PiperConfig contains local Piper TTS settings
type PiperConfig struct {
	BinaryPath string  `yaml:"binary_path" mapstructure:"binary_path"`
	ModelPath  string  `yaml:"model_path" mapstructure:"model_path"`
	Speed      float64 `yaml:"speed" mapstructure:"speed"`
}

// OpenAITTSConfig contains OpenAI TTS settings
type OpenAITTSConfig struct {
	APIKey string `yaml:"api_key" mapstructure:"api_key"`
	Model  string `yaml:"model" mapstructure:"model"`
	Voice  string `yaml:"voice" mapstructure:"voice"`
}

// TwitchConfig contains Twitch integration settings
type TwitchConfig struct {
	Enabled       bool   `yaml:"enabled" mapstructure:"enabled"`
	ClientID      string `yaml:"client_id" mapstructure:"client_id"`
	ClientSecret  string `yaml:"client_secret" mapstructure:"client_secret"`
	RedirectURI   string `yaml:"redirect_uri" mapstructure:"redirect_uri"`
	BroadcasterID string `yaml:"broadcaster_id" mapstructure:"broadcaster_id"`
	AccessToken   string `yaml:"access_token" mapstructure:"access_token"`
	RefreshToken  string `yaml:"refresh_token" mapstructure:"refresh_token"`
}

// OBSConfig contains OBS integration settings
type OBSConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	URL      string `yaml:"url" mapstructure:"url"`
	Password string `yaml:"password" mapstructure:"password"`
}

// MusicConfig contains music player settings
type MusicConfig struct {
	Enabled          bool     `yaml:"enabled" mapstructure:"enabled"`
	Folders          []string `yaml:"folders" mapstructure:"folders"`
	SupportedFormats []string `yaml:"supported_formats" mapstructure:"supported_formats"`
	DefaultVolume    float64  `yaml:"default_volume" mapstructure:"default_volume"`
	Shuffle          bool     `yaml:"shuffle" mapstructure:"shuffle"`
}

// SoundsConfig contains system sound settings
type SoundsConfig struct {
	Enabled        bool   `yaml:"enabled" mapstructure:"enabled"`
	Wake           string `yaml:"wake" mapstructure:"wake"`
	Error          string `yaml:"error" mapstructure:"error"`
	StartRecording string `yaml:"start_recording" mapstructure:"start_recording"`
	StopRecording  string `yaml:"stop_recording" mapstructure:"stop_recording"`
}
