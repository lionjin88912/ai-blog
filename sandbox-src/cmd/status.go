package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ai-sandbox/cli/internal/toolchain"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sandbox tool status",
	Long:  `Display which tools are installed in the sandbox and their versions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		absDir, err := filepath.Abs(sandboxDir)
		if err != nil {
			return err
		}

		p := toolchain.DetectPlatform()
		fmt.Printf("Platform:  %s/%s\n", p.OS, p.Arch)
		fmt.Printf("Sandbox:   %s\n\n", absDir)

		tools := []struct {
			name    string
			binPath string
			verArgs []string
		}{
			{"Node", toolchain.NodeBinPath(absDir, p), []string{"--version"}},
			{"npm", toolchain.NpmBinPath(absDir, p), []string{"--version"}},
			{"Antigravity CLI", toolchain.AntigravityBinPath(absDir, p), []string{"--version"}},
			{"Copilot CLI", toolchain.CopilotBinPath(absDir, p), []string{"--version"}},
			{"uv", toolchain.UVBinPath(absDir, p), []string{"--version"}},
			{"Python", toolchain.PythonBinPath(absDir, p), []string{"--version"}},
		}

		for _, tool := range tools {
			if _, err := os.Stat(tool.binPath); os.IsNotExist(err) {
				fmt.Printf("  %-12s ❌ not installed\n", tool.name)
				continue
			}
			ver := getVersion(tool.binPath, tool.verArgs)
			fmt.Printf("  %-12s ✅ %s\n", tool.name, ver)
		}

		return nil
	},
}

func getVersion(binPath string, args []string) string {
	cmd := exec.Command(binPath, args...)
	out, err := cmd.Output()
	if err != nil {
		return "(unknown)"
	}
	// Return first line only
	for i, b := range out {
		if b == '\n' || b == '\r' {
			return string(out[:i])
		}
	}
	return string(out)
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
