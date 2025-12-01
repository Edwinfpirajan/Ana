package config

// DefaultConfig returns a Config with sensible default values
func DefaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			Language: "es",
			LogLevel: "info",
			DataDir:  "./data",
		},
		Audio: AudioConfig{
			Device:     "default",
			SampleRate: 16000,
			Channels:   1,
			ChunkSize:  1024,
			VAD: VADConfig{
				Enabled:            true,
				Sensitivity:        0.5,
				SilenceThresholdMs: 1500,
				MinSpeechMs:        300,
			},
			WakeWord: WakeWordConfig{
				Enabled:   true,
				Word:      "ana",
				Threshold: 0.7,
			},
		},
		Hotkey: HotkeyConfig{
			Enabled: true,
			Key:     "F4",
			Mode:    "hold",
		},
		STT: STTConfig{
			Provider: "whisper",
			Whisper: WhisperConfig{
				BinaryPath: "./bin/whisper/whisper-cli",
				ModelPath:  "./assets/models/whisper/ggml-base.bin",
				Language:   "es",
			},
			OpenAI: OpenAISTTConfig{
				Model: "whisper-1",
			},
		},
		LLM: LLMConfig{
			Provider: "ollama",
			Ollama: OllamaConfig{
				URL:            "http://localhost:11434",
				Model:          "mistral:latest",
				TimeoutSeconds: 30,
			},
			OpenAI: OpenAILLMConfig{
				Model:       "gpt-4o-mini",
				Temperature: 0.3,
			},
		},
		TTS: TTSConfig{
			Provider: "piper",
			Piper: PiperConfig{
				BinaryPath: "./bin/piper/piper",
				ModelPath:  "./assets/voices/piper/carlfm/es_ES-carlfm-x-low.onnx",
				Speed:      1.0,
			},
			OpenAI: OpenAITTSConfig{
				Model: "tts-1",
				Voice: "nova",
			},
		},
		Twitch: TwitchConfig{
			Enabled:     false,
			RedirectURI: "http://localhost:3000/callback",
		},
		OBS: OBSConfig{
			Enabled: false,
			URL:     "ws://localhost:4455",
		},
		Music: MusicConfig{
			Enabled: true,
			Folders: []string{"./music"},
			SupportedFormats: []string{
				".mp3",
				".wav",
				".ogg",
				".flac",
			},
			DefaultVolume: 0.5,
			Shuffle:       false,
		},
		Sounds: SoundsConfig{
			Enabled:        true,
			Wake:           "./assets/sounds/wake.wav",
			Error:          "./assets/sounds/error.wav",
			StartRecording: "./assets/sounds/beep_start.wav",
			StopRecording:  "./assets/sounds/beep_end.wav",
		},
	}
}

// ApplyDefaults applies default values to missing config fields
func ApplyDefaults(cfg *Config) {
	defaults := DefaultConfig()

	// General
	if cfg.General.Language == "" {
		cfg.General.Language = defaults.General.Language
	}
	if cfg.General.LogLevel == "" {
		cfg.General.LogLevel = defaults.General.LogLevel
	}
	if cfg.General.DataDir == "" {
		cfg.General.DataDir = defaults.General.DataDir
	}

	// Audio
	if cfg.Audio.SampleRate == 0 {
		cfg.Audio.SampleRate = defaults.Audio.SampleRate
	}
	if cfg.Audio.Channels == 0 {
		cfg.Audio.Channels = defaults.Audio.Channels
	}
	if cfg.Audio.ChunkSize == 0 {
		cfg.Audio.ChunkSize = defaults.Audio.ChunkSize
	}
	if cfg.Audio.Device == "" {
		cfg.Audio.Device = defaults.Audio.Device
	}

	// VAD
	if cfg.Audio.VAD.SilenceThresholdMs == 0 {
		cfg.Audio.VAD.SilenceThresholdMs = defaults.Audio.VAD.SilenceThresholdMs
	}
	if cfg.Audio.VAD.MinSpeechMs == 0 {
		cfg.Audio.VAD.MinSpeechMs = defaults.Audio.VAD.MinSpeechMs
	}

	// Wake Word
	if cfg.Audio.WakeWord.Word == "" {
		cfg.Audio.WakeWord.Word = defaults.Audio.WakeWord.Word
	}
	if cfg.Audio.WakeWord.Threshold == 0 {
		cfg.Audio.WakeWord.Threshold = defaults.Audio.WakeWord.Threshold
	}

	// Hotkey
	if cfg.Hotkey.Key == "" {
		cfg.Hotkey.Key = defaults.Hotkey.Key
	}
	if cfg.Hotkey.Mode == "" {
		cfg.Hotkey.Mode = defaults.Hotkey.Mode
	}

	// STT
	if cfg.STT.Provider == "" {
		cfg.STT.Provider = defaults.STT.Provider
	}
	if cfg.STT.Whisper.BinaryPath == "" {
		cfg.STT.Whisper.BinaryPath = defaults.STT.Whisper.BinaryPath
	}
	if cfg.STT.Whisper.Language == "" {
		cfg.STT.Whisper.Language = defaults.STT.Whisper.Language
	}
	if cfg.STT.OpenAI.Model == "" {
		cfg.STT.OpenAI.Model = defaults.STT.OpenAI.Model
	}

	// LLM
	if cfg.LLM.Provider == "" {
		cfg.LLM.Provider = defaults.LLM.Provider
	}
	if cfg.LLM.Ollama.URL == "" {
		cfg.LLM.Ollama.URL = defaults.LLM.Ollama.URL
	}
	if cfg.LLM.Ollama.Model == "" {
		cfg.LLM.Ollama.Model = defaults.LLM.Ollama.Model
	}
	if cfg.LLM.Ollama.TimeoutSeconds == 0 {
		cfg.LLM.Ollama.TimeoutSeconds = defaults.LLM.Ollama.TimeoutSeconds
	}
	if cfg.LLM.OpenAI.Model == "" {
		cfg.LLM.OpenAI.Model = defaults.LLM.OpenAI.Model
	}
	if cfg.LLM.OpenAI.Temperature == 0 {
		cfg.LLM.OpenAI.Temperature = defaults.LLM.OpenAI.Temperature
	}

	// TTS
	if cfg.TTS.Provider == "" {
		cfg.TTS.Provider = defaults.TTS.Provider
	}
	if cfg.TTS.Piper.BinaryPath == "" {
		cfg.TTS.Piper.BinaryPath = defaults.TTS.Piper.BinaryPath
	}
	if cfg.TTS.Piper.Speed == 0 {
		cfg.TTS.Piper.Speed = defaults.TTS.Piper.Speed
	}
	if cfg.TTS.OpenAI.Model == "" {
		cfg.TTS.OpenAI.Model = defaults.TTS.OpenAI.Model
	}
	if cfg.TTS.OpenAI.Voice == "" {
		cfg.TTS.OpenAI.Voice = defaults.TTS.OpenAI.Voice
	}

	// Twitch
	if cfg.Twitch.RedirectURI == "" {
		cfg.Twitch.RedirectURI = defaults.Twitch.RedirectURI
	}

	// OBS
	if cfg.OBS.URL == "" {
		cfg.OBS.URL = defaults.OBS.URL
	}

	// Music
	if len(cfg.Music.Folders) == 0 {
		cfg.Music.Folders = defaults.Music.Folders
	}
	if len(cfg.Music.SupportedFormats) == 0 {
		cfg.Music.SupportedFormats = defaults.Music.SupportedFormats
	}
	if cfg.Music.DefaultVolume == 0 {
		cfg.Music.DefaultVolume = defaults.Music.DefaultVolume
	}

	// Sounds
	if cfg.Sounds.Wake == "" {
		cfg.Sounds.Wake = defaults.Sounds.Wake
	}
	if cfg.Sounds.Error == "" {
		cfg.Sounds.Error = defaults.Sounds.Error
	}
	if cfg.Sounds.StartRecording == "" {
		cfg.Sounds.StartRecording = defaults.Sounds.StartRecording
	}
	if cfg.Sounds.StopRecording == "" {
		cfg.Sounds.StopRecording = defaults.Sounds.StopRecording
	}
}
