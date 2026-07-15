package toolchain

import (
	"fmt"
	"os"
	"path/filepath"
)

// CopilotBinPath returns the path to the copilot CLI executable.
func CopilotBinPath(sandboxDir string, p Platform) string {
	if p.OS == "windows" {
		return filepath.Join(sandboxDir, "copilot", "node_modules", ".bin", "copilot.cmd")
	}
	return filepath.Join(sandboxDir, "copilot", "node_modules", ".bin", "copilot")
}

// InstallCopilot installs @github/copilot using the sandbox's npm.
func InstallCopilot(sandboxDir string, p Platform) error {
	if _, err := os.Stat(CopilotBinPath(sandboxDir, p)); err == nil {
		fmt.Println("  GitHub Copilot CLI already installed, skipping.")
		return nil
	}

	npmPath := NpmBinPath(sandboxDir, p)
	if _, err := os.Stat(npmPath); os.IsNotExist(err) {
		return fmt.Errorf("npm not found at %s — run Node setup first", npmPath)
	}

	copilotDir := filepath.Join(sandboxDir, "copilot")
	if err := os.MkdirAll(copilotDir, 0755); err != nil {
		return err
	}

	fmt.Println("  Installing GitHub Copilot CLI...")
	cmd := execCommand(npmPath, "install", "--prefix", copilotDir, "@github/copilot")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = copilotDir

	cmd.Env = append(os.Environ(),
		"PATH="+filepath.Dir(NodeBinPath(sandboxDir, p))+string(os.PathListSeparator)+os.Getenv("PATH"),
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm install @github/copilot: %w", err)
	}

	fmt.Println("  ✅ GitHub Copilot CLI installed.")
	return nil
}
