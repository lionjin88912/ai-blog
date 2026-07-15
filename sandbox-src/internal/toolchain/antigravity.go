package toolchain

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Manifest represents the antigravity auto-updater manifest
type Manifest struct {
	Version string `json:"version"`
	URL     string `json:"url"`
	SHA512  string `json:"sha512"`
}

// AntigravityBinPath returns the path to the antigravity CLI executable.
func AntigravityBinPath(sandboxDir string, p Platform) string {
	return filepath.Join(sandboxDir, "antigravity", "agy"+p.ExeExt())
}

// InstallAntigravityCLI downloads and installs antigravity CLI.
func InstallAntigravityCLI(sandboxDir string, p Platform) error {
	if _, err := os.Stat(AntigravityBinPath(sandboxDir, p)); err == nil {
		fmt.Println("  Antigravity CLI already installed, skipping.")
		return nil
	}

	platformStr := p.OS + "_" + p.Arch

	manifestURL := fmt.Sprintf("https://antigravity-cli-auto-updater-974169037036.us-central1.run.app/manifests/%s.json", platformStr)
	fmt.Printf("  Downloading Antigravity CLI manifest for %s...\n", platformStr)
	
	resp, err := http.Get(manifestURL)
	if err != nil {
		return fmt.Errorf("download manifest %s: %w", manifestURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download manifest %s: HTTP %d", manifestURL, resp.StatusCode)
	}

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return fmt.Errorf("decode manifest: %w", err)
	}

	agyDir := filepath.Join(sandboxDir, "antigravity")
	if err := os.MkdirAll(agyDir, 0755); err != nil {
		return err
	}

	targetFile := AntigravityBinPath(sandboxDir, p)
	
	fmt.Printf("  Downloading Antigravity CLI %s...\n", manifest.Version)
	if err := DownloadFile(manifest.URL, targetFile); err != nil {
		return fmt.Errorf("download antigravity: %w", err)
	}

	fmt.Println("  Verifying checksum...")
	if err := verifyChecksum(targetFile, manifest.SHA512); err != nil {
		os.Remove(targetFile)
		return fmt.Errorf("checksum failed: %w", err)
	}

	// Make executable
	if err := os.Chmod(targetFile, 0755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	fmt.Println("  Running Antigravity setup...")
	cmd := execCommand(targetFile, "install", "--dir", agyDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = agyDir

	// Ignore errors from `install` as the binary is successfully copied (similar to powershell script logic)
	cmd.Run()

	fmt.Println("  ✅ Antigravity CLI installed.")
	return nil
}

func verifyChecksum(filePath, expected string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if actual != strings.ToLower(expected) {
		return fmt.Errorf("expected %s, got %s", expected, actual)
	}
	return nil
}
