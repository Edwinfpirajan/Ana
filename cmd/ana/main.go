package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/anastreamer/ana/internal/audio"
	"github.com/anastreamer/ana/internal/brain"
	"github.com/anastreamer/ana/internal/config"
	"github.com/anastreamer/ana/internal/hotkey"
	"github.com/anastreamer/ana/internal/llm"
	"github.com/anastreamer/ana/internal/pipeline"
	"github.com/anastreamer/ana/internal/stt"
	"github.com/anastreamer/ana/internal/tts"
	"github.com/anastreamer/ana/pkg/logger"
)

func main() {
	// Initialize logger
	logger.Init("info", nil)

	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		logger.Error("Failed to load configuration", err)
		os.Exit(1)
	}

	logger.Info("Ana Streamer starting...")

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize STT Provider
	var sttProvider stt.Provider
	sttProvider, err = initializeSTT(ctx, cfg)
	if err != nil {
		logger.Error("Failed to initialize STT provider", err)
		os.Exit(1)
	}

	// Initialize LLM Provider
	var llmProvider llm.Provider
	llmProvider, err = initializeLLM(ctx, cfg)
	if err != nil {
		logger.Error("Failed to initialize LLM provider", err)
		os.Exit(1)
	}

	// Initialize TTS Provider
	var ttsProvider tts.Provider
	ttsProvider, err = initializeTTS(ctx, cfg)
	if err != nil {
		logger.Warn(fmt.Sprintf("TTS provider not available, continuing without audio: %v", err))
	}

	// Create Brain
	brn := brain.New(llmProvider, ttsProvider)

	// Register executors
	if cfg.Twitch.Enabled {
		logger.Info("Registering Twitch executor")
		// TODO: Initialize Twitch executor
	}
	if cfg.OBS.Enabled {
		logger.Info("Registering OBS executor")
		// TODO: Initialize OBS executor
	}
	if cfg.Music.Enabled {
		logger.Info("Registering Music executor")
		// TODO: Initialize Music executor
	}

	// Create Pipeline
	ppl := pipeline.NewPipeline(cfg, sttProvider, brn)

	// Set callbacks for UI feedback
	ppl.SetCallbacks(
		func(state pipeline.State) {
			logger.Info(fmt.Sprintf("Pipeline state: %s", state.String()))
		},
		func(text string) {
			fmt.Printf("📝 You: %s\n", text)
		},
		func(response string) {
			fmt.Printf("🤖 Ana: %s\n", response)
		},
		func(err error) {
			logger.Error(fmt.Sprintf("Pipeline error: %v", err), nil)
		},
	)

	// Start pipeline
	if err := ppl.Start(ctx); err != nil {
		logger.Error("Failed to start pipeline", err)
		os.Exit(1)
	}

	// Initialize audio capture
	audioCapture, err := audio.Start(ctx, cfg.Audio, ppl, sttProvider, nil)
	if err != nil {
		logger.Error("Failed to initialize audio capture", err)
		os.Exit(1)
	}

	// Initialize hotkey listener
	hk, err := hotkey.NewListener(ctx, ppl.TriggerHotkeyDown, ppl.TriggerHotkeyUp)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to initialize hotkey listener: %v", err))
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("🎤 Ana Streamer Active")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("🔊 How to use:")
	fmt.Println("   1. Say 'Ana' to activate persistent session")
	fmt.Println("   2. Keep talking - no need to repeat 'Ana'")
	fmt.Println("   3. Say 'Adiós Ana' or similar to deactivate")
	fmt.Println()
	fmt.Println("💬 Deactivation words: adiós, detente, silencio, para ana,")
	fmt.Println("   cállate, quieta, deja de grabar, stop, adiós ana")
	fmt.Println()
	fmt.Println("⌨️  Hotkey: F4 (press and hold to record)")
	fmt.Println()
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	// Wait for shutdown signal
	<-sigChan

	fmt.Println("\n🛑 Shutting down Ana Streamer...")

	// Cleanup
	cancel()
	if audioCapture != nil {
		audioCapture.Stop()
	}
	if hk != nil {
		hk.Close()
	}
	ppl.Stop()

	if sttProvider != nil {
		sttProvider.Close()
	}
	if llmProvider != nil {
		llmProvider.Close()
	}
	if ttsProvider != nil {
		ttsProvider.Close()
	}

	fmt.Println("✅ Ana Streamer stopped")
}

func initializeSTT(ctx context.Context, cfg *config.Config) (stt.Provider, error) {
	switch cfg.STT.Provider {
	case "whisper":
		logger.Info("Initializing Whisper STT provider")
		return stt.NewWhisperProvider(cfg.STT.Whisper)
	case "openai":
		logger.Info("Initializing OpenAI STT provider")
		return stt.NewOpenAIProvider(cfg.STT.OpenAI)
	default:
		return nil, fmt.Errorf("unknown STT provider: %s", cfg.STT.Provider)
	}
}

func initializeLLM(ctx context.Context, cfg *config.Config) (llm.Provider, error) {
	streamerName := cfg.General.StreamerName
	switch cfg.LLM.Provider {
	case "ollama":
		logger.Info("Initializing Ollama LLM provider")
		return llm.NewOllamaProvider(cfg.LLM.Ollama, streamerName)
	case "openai":
		logger.Info("Initializing OpenAI LLM provider")
		return llm.NewOpenAIProvider(cfg.LLM.OpenAI, streamerName)
	case "auto":
		logger.Info("Trying Ollama first, fallback to OpenAI")
		provider, err := llm.NewOllamaProvider(cfg.LLM.Ollama, streamerName)
		if err != nil || !provider.IsAvailable(ctx) {
			logger.Warn("Ollama not available, using OpenAI")
			return llm.NewOpenAIProvider(cfg.LLM.OpenAI, streamerName)
		}
		return provider, nil
	default:
		return nil, fmt.Errorf("unknown LLM provider: %s", cfg.LLM.Provider)
	}
}

func initializeTTS(ctx context.Context, cfg *config.Config) (tts.Provider, error) {
	switch cfg.TTS.Provider {
	case "piper":
		logger.Info("Initializing Piper TTS provider")
		return tts.NewPiperProvider(cfg.TTS.Piper)
	case "openai":
		logger.Info("Initializing OpenAI TTS provider")
		return tts.NewOpenAIProvider(cfg.TTS.OpenAI)
	case "auto":
		logger.Info("Trying Piper first, fallback to OpenAI")
		provider, err := tts.NewPiperProvider(cfg.TTS.Piper)
		if err != nil || !provider.IsAvailable(ctx) {
			logger.Warn("Piper not available, using OpenAI")
			return tts.NewOpenAIProvider(cfg.TTS.OpenAI)
		}
		return provider, nil
	default:
		return nil, fmt.Errorf("unknown TTS provider: %s", cfg.TTS.Provider)
	}
}
