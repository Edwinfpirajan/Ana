package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// ConfigPaths defines the paths where config files are searched
var ConfigPaths = []string{
	".",
	"./config",
	"$HOME/.ana",
	"/etc/ana",
}

// ConfigName is the name of the config file (without extension)
const ConfigName = "ana.config"

// Load loads the configuration from file
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set config file directly if path is provided
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Set config name and type
		v.SetConfigName(ConfigName)
		v.SetConfigType("yaml")

		// Add search paths
		for _, path := range ConfigPaths {
			expandedPath := os.ExpandEnv(path)
			v.AddConfigPath(expandedPath)
		}
	}

	// Enable environment variable override
	v.SetEnvPrefix("ANA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults
			cfg := DefaultConfig()
			return cfg, nil
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal into config struct
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Apply defaults for missing values
	ApplyDefaults(cfg)

	// Expand environment variables in string fields
	expandEnvVars(cfg)

	// Validate config
	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	return cfg, nil
}

// expandEnvVars expands environment variables in config strings
func expandEnvVars(cfg *Config) {
	// STT
	cfg.STT.OpenAI.APIKey = os.ExpandEnv(cfg.STT.OpenAI.APIKey)
	cfg.STT.Whisper.BinaryPath = os.ExpandEnv(cfg.STT.Whisper.BinaryPath)
	cfg.STT.Whisper.ModelPath = os.ExpandEnv(cfg.STT.Whisper.ModelPath)

	// LLM
	cfg.LLM.OpenAI.APIKey = os.ExpandEnv(cfg.LLM.OpenAI.APIKey)

	// TTS
	cfg.TTS.OpenAI.APIKey = os.ExpandEnv(cfg.TTS.OpenAI.APIKey)
	cfg.TTS.Piper.BinaryPath = os.ExpandEnv(cfg.TTS.Piper.BinaryPath)
	cfg.TTS.Piper.ModelPath = os.ExpandEnv(cfg.TTS.Piper.ModelPath)

	// Paths
	cfg.General.DataDir = os.ExpandEnv(cfg.General.DataDir)

	// Music folders
	for i, folder := range cfg.Music.Folders {
		cfg.Music.Folders[i] = os.ExpandEnv(folder)
	}

	// Sounds
	cfg.Sounds.Wake = os.ExpandEnv(cfg.Sounds.Wake)
	cfg.Sounds.Error = os.ExpandEnv(cfg.Sounds.Error)
	cfg.Sounds.StartRecording = os.ExpandEnv(cfg.Sounds.StartRecording)
	cfg.Sounds.StopRecording = os.ExpandEnv(cfg.Sounds.StopRecording)
}

// Validate validates the configuration
func Validate(cfg *Config) error {
	var errors []string

	// Validate STT config
	switch cfg.STT.Provider {
	case "whisper":
		// Whisper binary and model will be validated at runtime
	case "openai":
		if cfg.STT.OpenAI.APIKey == "" {
			errors = append(errors, "OpenAI API key required for STT when using OpenAI provider")
		}
	default:
		errors = append(errors, fmt.Sprintf("invalid STT provider: %s (must be 'whisper' or 'openai')", cfg.STT.Provider))
	}

	// Validate LLM config
	switch cfg.LLM.Provider {
	case "ollama":
		if cfg.LLM.Ollama.URL == "" {
			errors = append(errors, "Ollama URL required when using Ollama provider")
		}
		if cfg.LLM.Ollama.Model == "" {
			errors = append(errors, "Ollama model required when using Ollama provider")
		}
	case "openai":
		if cfg.LLM.OpenAI.APIKey == "" {
			errors = append(errors, "OpenAI API key required for LLM when using OpenAI provider")
		}
	case "auto":
		if cfg.LLM.OpenAI.APIKey == "" && cfg.LLM.Ollama.URL == "" {
			errors = append(errors, "Auto LLM provider requires either Ollama config or OpenAI API key")
		}
	default:
		errors = append(errors, fmt.Sprintf("invalid LLM provider: %s (must be 'ollama', 'openai' or 'auto')", cfg.LLM.Provider))
	}

	// Validate TTS config
	switch cfg.TTS.Provider {
	case "piper":
		// Piper binary and model will be validated at runtime
	case "openai":
		if cfg.TTS.OpenAI.APIKey == "" {
			errors = append(errors, "OpenAI API key required for TTS when using OpenAI provider")
		}
	case "auto":
		if cfg.TTS.OpenAI.APIKey == "" && cfg.TTS.Piper.BinaryPath == "" {
			errors = append(errors, "Auto TTS provider requires either Piper binary or OpenAI API key")
		}
	default:
		errors = append(errors, fmt.Sprintf("invalid TTS provider: %s (must be 'piper', 'openai' or 'auto')", cfg.TTS.Provider))
	}

	// Validate Twitch config if enabled
	if cfg.Twitch.Enabled {
		if cfg.Twitch.ClientID == "" {
			errors = append(errors, "Twitch client_id required when Twitch is enabled")
		}
	}

	// Validate audio config
	if cfg.Audio.SampleRate <= 0 {
		errors = append(errors, "audio sample_rate must be positive")
	}
	if cfg.Audio.Channels <= 0 {
		errors = append(errors, "audio channels must be positive")
	}

	// Validate VAD config
	if cfg.Audio.VAD.Sensitivity < 0 || cfg.Audio.VAD.Sensitivity > 1 {
		errors = append(errors, "VAD sensitivity must be between 0 and 1")
	}

	// Validate hotkey mode
	if cfg.Hotkey.Enabled {
		if cfg.Hotkey.Mode != "hold" && cfg.Hotkey.Mode != "toggle" {
			errors = append(errors, "hotkey mode must be 'hold' or 'toggle'")
		}
	}

	// Validate music config
	if cfg.Music.DefaultVolume < 0 || cfg.Music.DefaultVolume > 1 {
		errors = append(errors, "music default_volume must be between 0 and 1")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors:\n- %s", strings.Join(errors, "\n- "))
	}

	return nil
}

// Save saves the configuration to a file
func Save(cfg *Config, path string) error {
	v := viper.New()

	// Set all values
	v.Set("general", cfg.General)
	v.Set("audio", cfg.Audio)
	v.Set("hotkey", cfg.Hotkey)
	v.Set("stt", cfg.STT)
	v.Set("llm", cfg.LLM)
	v.Set("tts", cfg.TTS)
	v.Set("twitch", cfg.Twitch)
	v.Set("obs", cfg.OBS)
	v.Set("music", cfg.Music)
	v.Set("sounds", cfg.Sounds)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file
	if err := v.WriteConfigAs(path); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the path to the config file that was loaded
func GetConfigPath() string {
	return viper.ConfigFileUsed()
}
