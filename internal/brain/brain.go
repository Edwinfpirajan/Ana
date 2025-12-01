// Package brain provides the core orchestration for AnaStreamer
package brain

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/anastreamer/ana/internal/executor"
	"github.com/anastreamer/ana/internal/llm"
	"github.com/anastreamer/ana/internal/tts"
	"github.com/anastreamer/ana/pkg/logger"
	"github.com/rs/zerolog"
)

// Brain is the central orchestrator that processes commands
type Brain struct {
	llmProvider llm.Provider
	ttsProvider tts.Provider
	registry    *executor.Registry
	log         zerolog.Logger
}

// New creates a new Brain instance
func New(llmProvider llm.Provider, ttsProvider tts.Provider) *Brain {
	return &Brain{
		llmProvider: llmProvider,
		ttsProvider: ttsProvider,
		registry:    executor.NewRegistry(),
		log:         logger.Component("brain"),
	}
}

// RegisterExecutor registers an action executor
func (b *Brain) RegisterExecutor(exec executor.Executor) {
	b.registry.Register(exec)
	b.log.Debug().
		Str("executor", exec.Name()).
		Strs("actions", exec.SupportedActions()).
		Msg("Registered executor")
}

// ProcessCommand processes a voice command and returns the response
func (b *Brain) ProcessCommand(ctx context.Context, text string) (string, error) {
	b.log.Info().Str("input", text).Msg("Processing command")

	if b.llmProvider == nil || !b.llmProvider.IsAvailable(ctx) {
		b.log.Warn().Msg("LLM provider is not available, skipping command")
		return "No hay ningún proveedor de IA disponible. Revisa tu configuración o prueba más tarde.", nil
	}

	// Get action from LLM
	action, err := b.llmProvider.Complete(ctx, text)
	if err != nil {
		b.log.Error().Err(err).Msg("LLM completion failed")
		return "", fmt.Errorf("failed to interpret command: %w", err)
	}

	b.log.Debug().
		Str("action", action.Action).
		Interface("params", action.Params).
		Str("reply", action.Reply).
		Msg("LLM response")

	// Handle special actions
	if action.Action == "none" || action.Action == "" {
		// Just respond without executing
		return action.Reply, nil
	}

	if action.Action == "system.status" {
		return b.handleStatus(ctx, action)
	}

	if action.Action == "system.help" {
		return b.handleHelp(ctx, action)
	}

	if action.Action == "calc" {
		return b.handleCalc(ctx, action)
	}

	// Execute the action
	result, err := b.registry.Execute(ctx, action)
	if err != nil {
		b.log.Error().Err(err).Str("action", action.Action).Msg("Action execution failed")
		// Return the LLM's reply anyway, plus error info
		return fmt.Sprintf("%s. Sin embargo, hubo un error: %s", action.Reply, err.Error()), nil
	}

	if !result.Success {
		b.log.Warn().
			Str("action", action.Action).
			Str("error", result.Error).
			Msg("Action failed")
		return fmt.Sprintf("%s. Error: %s", action.Reply, result.Error), nil
	}

	b.log.Info().
		Str("action", action.Action).
		Str("result", result.Message).
		Msg("Action executed successfully")

	// Return the LLM's reply (which should be natural language)
	return action.Reply, nil
}

// ProcessAndSpeak processes a command and speaks the response
func (b *Brain) ProcessAndSpeak(ctx context.Context, text string) error {
	response, err := b.ProcessCommand(ctx, text)
	if err != nil {
		response = "Lo siento, ocurrió un error procesando tu solicitud."
	}

	if response == "" {
		return nil
	}

	// Speak the response (but don't fail if TTS has issues)
	if b.ttsProvider != nil {
		if b.ttsProvider.IsAvailable(ctx) {
			if err := b.ttsProvider.Speak(ctx, response); err != nil {
				b.log.Warn().Err(err).Msg("TTS failed, but continuing without audio")
				// Don't return error, just warn and continue
			}
		} else {
			b.log.Debug().Msg("TTS provider is not available, skipping speech")
		}
	}

	return nil
}

// handleStatus returns the system status
func (b *Brain) handleStatus(ctx context.Context, action llm.Action) (string, error) {
	var status []string

	// Check LLM
	if b.llmProvider.IsAvailable(ctx) {
		status = append(status, fmt.Sprintf("LLM %s: activo", b.llmProvider.Name()))
	} else {
		status = append(status, fmt.Sprintf("LLM %s: no disponible", b.llmProvider.Name()))
	}

	// Check TTS
	if b.ttsProvider != nil && b.ttsProvider.IsAvailable(ctx) {
		status = append(status, fmt.Sprintf("TTS %s: activo", b.ttsProvider.Name()))
	} else {
		status = append(status, "TTS: no disponible")
	}

	// Check executors
	for _, actionName := range []string{"twitch", "obs", "music"} {
		exec, ok := b.registry.Get(actionName)
		if ok && exec.IsAvailable() {
			status = append(status, fmt.Sprintf("%s: conectado", actionName))
		} else if ok {
			status = append(status, fmt.Sprintf("%s: desconectado", actionName))
		}
	}

	return "Estado del sistema: " + strings.Join(status, ", "), nil
}

// handleHelp returns help information
func (b *Brain) handleHelp(ctx context.Context, action llm.Action) (string, error) {
	actions := b.registry.GetAllActions()

	categories := make(map[string][]string)
	for _, a := range actions {
		parts := strings.Split(a, ".")
		if len(parts) >= 2 {
			categories[parts[0]] = append(categories[parts[0]], a)
		}
	}

	var help []string
	help = append(help, "Puedo ayudarte con:")

	if actions, ok := categories["twitch"]; ok {
		help = append(help, fmt.Sprintf("- Twitch: %d acciones (clips, título, bans)", len(actions)))
	}
	if actions, ok := categories["obs"]; ok {
		help = append(help, fmt.Sprintf("- OBS: %d acciones (escenas, fuentes, volumen)", len(actions)))
	}
	if actions, ok := categories["music"]; ok {
		help = append(help, fmt.Sprintf("- Música: %d acciones (play, pause, volumen)", len(actions)))
	}

	help = append(help, "¿Qué necesitas?")

	return strings.Join(help, " "), nil
}

// handleCalc performs mathematical calculations
func (b *Brain) handleCalc(ctx context.Context, action llm.Action) (string, error) {
	expression, ok := action.Params["expression"].(string)
	if !ok || expression == "" {
		return "No entiendo la expresión matemática. Por favor intenta de nuevo.", nil
	}

	result, err := evaluateExpression(expression)
	if err != nil {
		b.log.Error().Err(err).Str("expression", expression).Msg("Calculation failed")
		return fmt.Sprintf("No pude calcular esa expresión: %v", err), nil
	}

	b.log.Debug().Str("expression", expression).Float64("result", result).Msg("Calculation successful")
	return action.Reply, nil
}

// evaluateExpression safely evaluates a mathematical expression
func evaluateExpression(expr string) (float64, error) {
	// Remove spaces
	expr = strings.ReplaceAll(expr, " ", "")

	// Validate: only allow numbers, operators, and parentheses
	validChars := regexp.MustCompile(`^[0-9+\-*/.()^]+$`)
	if !validChars.MatchString(expr) {
		return 0, fmt.Errorf("expresión inválida")
	}

	// Simple evaluation - supports +, -, *, /, ^
	// Replace ^ with ** for Go's math notation (not needed, we'll use basic eval)
	// For safety, we'll implement a basic evaluator

	return simpleEval(expr)
}

// simpleEval performs basic arithmetic evaluation
func simpleEval(expr string) (float64, error) {
	// Handle parentheses recursively
	for {
		// Find innermost parentheses
		openIdx := strings.LastIndex(expr, "(")
		if openIdx == -1 {
			break
		}

		closeIdx := strings.Index(expr[openIdx:], ")")
		if closeIdx == -1 {
			return 0, fmt.Errorf("paréntesis no balanceados")
		}
		closeIdx += openIdx

		inner := expr[openIdx+1 : closeIdx]
		result, err := simpleEval(inner)
		if err != nil {
			return 0, err
		}

		expr = expr[:openIdx] + fmt.Sprintf("%g", result) + expr[closeIdx+1:]
	}

	// Split by + and - (lowest precedence)
	parts := regexp.MustCompile(`([+\-])`).Split(expr, -1)
	if len(parts) == 0 {
		return 0, fmt.Errorf("expresión vacía")
	}

	// Handle the first number
	result, err := parseAndMultiplyDivide(parts[0])
	if err != nil {
		return 0, err
	}

	// Process remaining parts
	for i := 1; i < len(parts); i += 2 {
		if i >= len(parts) {
			break
		}

		operator := parts[i-1][len(parts[i-1])-1:] // Get last char
		if operator != "+" && operator != "-" {
			operator = "+"
			i--
		}

		nextVal, err := parseAndMultiplyDivide(parts[i])
		if err != nil {
			return 0, err
		}

		if operator == "+" {
			result += nextVal
		} else {
			result -= nextVal
		}
	}

	return result, nil
}

// parseAndMultiplyDivide handles multiplication and division
func parseAndMultiplyDivide(expr string) (float64, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return 0, fmt.Errorf("expresión vacía")
	}

	// Split by * and /
	parts := regexp.MustCompile(`([*/])`).Split(expr, -1)

	result, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("número inválido: %s", parts[0])
	}

	for i := 1; i < len(parts); i += 2 {
		if i >= len(parts) {
			break
		}

		operator := parts[i-1][len(parts[i-1])-1:]
		nextStr := parts[i]

		nextVal, err := strconv.ParseFloat(nextStr, 64)
		if err != nil {
			return 0, fmt.Errorf("número inválido: %s", nextStr)
		}

		if operator == "*" {
			result *= nextVal
		} else if operator == "/" {
			if nextVal == 0 {
				return 0, fmt.Errorf("división entre cero")
			}
			result /= nextVal
		}
	}

	return result, nil
}

// GetAvailableActions returns all available actions
func (b *Brain) GetAvailableActions() []string {
	return b.registry.GetAllActions()
}

// SetLLM sets the LLM provider
func (b *Brain) SetLLM(provider llm.Provider) {
	b.llmProvider = provider
}

// SetTTS sets the TTS provider
func (b *Brain) SetTTS(provider tts.Provider) {
	b.ttsProvider = provider
}

// Close releases resources
func (b *Brain) Close() error {
	var errs []error

	if err := b.registry.Close(); err != nil {
		errs = append(errs, err)
	}

	if b.llmProvider != nil {
		if err := b.llmProvider.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if b.ttsProvider != nil {
		if err := b.ttsProvider.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing brain: %v", errs)
	}

	return nil
}
