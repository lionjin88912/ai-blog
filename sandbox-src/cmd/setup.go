package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ai-sandbox/cli/internal/seed"
	"github.com/ai-sandbox/cli/internal/toolchain"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Download and install all sandbox tools",
	Long: `Download Node 22, Antigravity CLI, GitHub Copilot CLI,
uv, and Python 3.12 to the ./sandbox/ directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		absDir, err := filepath.Abs(sandboxDir)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(absDir, 0755); err != nil {
			return fmt.Errorf("create sandbox dir: %w", err)
		}

		// First-run seeding: factory skills/docs into the workspace.
		seed.EnsureReport(filepath.Dir(absDir), func(f string, a ...any) { fmt.Printf(f+"\n", a...) })

		p := toolchain.DetectPlatform()
		fmt.Printf("Platform: %s/%s\n", p.OS, p.Arch)
		fmt.Printf("Sandbox:  %s\n\n", absDir)

		steps := []struct {
			name string
			fn   func() error
		}{
			{"Node 22", func() error { return toolchain.DownloadNode(absDir, p) }},
			{"Antigravity CLI", func() error { return toolchain.InstallAntigravityCLI(absDir, p) }},
			{"GitHub Copilot CLI", func() error { return toolchain.InstallCopilot(absDir, p) }},
			{"uv", func() error { return toolchain.DownloadUV(absDir, p) }},
			{"Python 3.12", func() error { return toolchain.InstallPython(absDir, p) }},
			{"Portable Git", func() error { return toolchain.DownloadGit(absDir, p) }},
			{"Shims", func() error { return toolchain.WriteShims(absDir, p) }},
		}

		for _, step := range steps {
			fmt.Printf("[%s]\n", step.name)
			if err := step.fn(); err != nil {
				return fmt.Errorf("%s: %w", step.name, err)
			}
			fmt.Println()
		}

		fmt.Println("✅ Sandbox ready!")
		fmt.Printf("Run 'ai-sandbox shell' to open a shell with all tools in PATH.\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
