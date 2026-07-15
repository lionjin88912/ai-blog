package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	configDirName  = ".ai-sandbox"
	configFileName = "config.json"
	configFileMode = 0600
	configDirMode  = 0700
)

// Config holds the sandbox configuration.
type Config struct {
	GeminiAPIKey  string `json:"gemini_api_key"`
	GitHubToken   string `json:"github_token"`
	WorkspacePath string `json:"workspace_path"`
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		WorkspacePath: filepath.Join(home, "workspace"),
	}
}

// DataDir returns the per-user directory where the sandbox keeps everything it
// generates: downloaded tools (sandbox/), skills (.agents/), workspace/, and
// the seeded docs. Placing it here means the exe is a pure launcher — it can
// live in Downloads and never scatters files next to itself.
//
//	Windows: %LOCALAPPDATA%\ai-sandbox
//	macOS:   ~/Library/Application Support/ai-sandbox
//	Linux:   $XDG_DATA_HOME/ai-sandbox  (or ~/.local/share/ai-sandbox)
func DataDir() (string, error) {
	const app = "ai-sandbox"
	switch runtime.GOOS {
	case "windows":
		if la := os.Getenv("LOCALAPPDATA"); la != "" {
			return filepath.Join(la, app), nil
		}
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, "Library", "Application Support", app), nil
		}
	default: // linux and others
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return filepath.Join(xdg, app), nil
		}
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".local", "share", app), nil
		}
	}
	// Fallback: user config dir (roaming on Windows) — still user-writable.
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine data directory: %w", err)
	}
	return filepath.Join(base, app), nil
}

// ConfigDir returns the path to ~/.ai-sandbox.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, configDirName), nil
}

// ConfigPath returns the full path to the config file.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// Exists returns true if the config file exists.
func Exists() bool {
	path, err := ConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// Load reads the config from disk. Returns defaults if the file does not exist.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}

// QuickInit creates a default config if one does not already exist.
// It sets WorkspacePath to the given path and creates the directory.
// API keys are left empty for the user to configure later.
func QuickInit(workspacePath string) error {
	if Exists() {
		return nil
	}

	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return fmt.Errorf("create workspace: %w", err)
	}

	cfg := &Config{
		WorkspacePath: workspacePath,
	}
	return Save(cfg)
}

// Save writes the config to disk, creating the directory if needed.
func Save(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, configDirMode); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	path := filepath.Join(dir, configFileName)
	if err := os.WriteFile(path, data, configFileMode); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
