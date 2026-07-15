package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.WorkspacePath == "" {
		t.Error("expected default WorkspacePath to be set")
	}
	if cfg.GeminiAPIKey != "" {
		t.Error("expected default GeminiAPIKey to be empty")
	}
	if cfg.GitHubToken != "" {
		t.Error("expected default GitHubToken to be empty")
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Use a temp directory as home
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir) // Windows

	cfg := &Config{
		GeminiAPIKey:  "test-gemini-key",
		GitHubToken:   "test-github-token",
		WorkspacePath: "/tmp/workspace",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file permissions (skip on Windows)
	configPath := filepath.Join(tmpDir, configDirName, configFileName)
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("config file not found: %v", err)
	}
	if info.Mode().Perm()&0077 != 0 && os.Getenv("OS") != "Windows_NT" {
		t.Errorf("config file permissions too open: %v", info.Mode().Perm())
	}

	// Load back
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.GeminiAPIKey != cfg.GeminiAPIKey {
		t.Errorf("GeminiAPIKey mismatch: got %q, want %q", loaded.GeminiAPIKey, cfg.GeminiAPIKey)
	}
	if loaded.GitHubToken != cfg.GitHubToken {
		t.Errorf("GitHubToken mismatch: got %q, want %q", loaded.GitHubToken, cfg.GitHubToken)
	}
	if loaded.WorkspacePath != cfg.WorkspacePath {
		t.Errorf("WorkspacePath mismatch: got %q, want %q", loaded.WorkspacePath, cfg.WorkspacePath)
	}
}

func TestLoadReturnsDefaultsWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.GeminiAPIKey != "" {
		t.Error("expected empty GeminiAPIKey for missing config")
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	if Exists() {
		t.Error("expected Exists() to return false before save")
	}

	cfg := DefaultConfig()
	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if !Exists() {
		t.Error("expected Exists() to return true after save")
	}
}

func TestQuickInit(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	wsPath := filepath.Join(tmpDir, "myworkspace")

	// Should create config and workspace dir
	if err := QuickInit(wsPath); err != nil {
		t.Fatalf("QuickInit failed: %v", err)
	}

	// Config file should exist
	if !Exists() {
		t.Error("expected config to exist after QuickInit")
	}

	// Workspace dir should exist
	info, err := os.Stat(wsPath)
	if err != nil {
		t.Fatalf("workspace dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("workspace path is not a directory")
	}

	// Config should have correct values
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.WorkspacePath != wsPath {
		t.Errorf("WorkspacePath = %q, want %q", cfg.WorkspacePath, wsPath)
	}
	if cfg.GeminiAPIKey != "" {
		t.Error("expected empty GeminiAPIKey")
	}
	if cfg.GitHubToken != "" {
		t.Error("expected empty GitHubToken")
	}
}

func TestQuickInitSkipsIfExists(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	// Save a config with a key first
	original := &Config{
		GeminiAPIKey:  "my-key",
		WorkspacePath: "/original",
	}
	if err := Save(original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// QuickInit should not overwrite
	wsPath := filepath.Join(tmpDir, "newworkspace")
	if err := QuickInit(wsPath); err != nil {
		t.Fatalf("QuickInit failed: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.GeminiAPIKey != "my-key" {
		t.Errorf("QuickInit overwrote existing config: GeminiAPIKey = %q", cfg.GeminiAPIKey)
	}
}
