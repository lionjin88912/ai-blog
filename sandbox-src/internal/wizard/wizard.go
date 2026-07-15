package wizard

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ai-sandbox/cli/internal/config"
)

// Run executes the interactive setup wizard and returns a Config.
func Run() (*config.Config, error) {
	cfg := config.DefaultConfig()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🔧 AI Sandbox Setup")
	fmt.Println("-------------------")
	fmt.Println()

	// Gemini API Key
	fmt.Print("Gemini API Key: ")
	key, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	key = strings.TrimSpace(key)
	if key != "" {
		cfg.GeminiAPIKey = key
	}

	// GitHub Token
	fmt.Print("GitHub Token: ")
	token, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	token = strings.TrimSpace(token)
	if token != "" {
		cfg.GitHubToken = token
	}

	// Workspace Path
	fmt.Printf("Workspace path [%s]: ", cfg.WorkspacePath)
	wsPath, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	wsPath = strings.TrimSpace(wsPath)
	if wsPath != "" {
		cfg.WorkspacePath = wsPath
	}

	// Create workspace directory if it doesn't exist
	if err := os.MkdirAll(cfg.WorkspacePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace directory: %w", err)
	}

	return cfg, nil
}
