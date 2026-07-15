package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ai-sandbox/cli/internal/toolchain"
	"github.com/spf13/cobra"
)

var (
	version    = "0.1.0"
	sandboxDir string
	shellFlag  string
)

var rootCmd = &cobra.Command{
	Use:   "ai-sandbox",
	Short: "AI CLI sandbox manager",
	Long: `A fully containerized, local AI sandbox environment containing
Node 22, Python 3.12, Antigravity CLI, GitHub Copilot CLI, and uv.

Get started:
  ai-sandbox init     Configure API keys and workspace
  ai-sandbox setup    Download tools to ./sandbox/
  ai-sandbox shell    Open a shell with sandbox tools in PATH`,
	Version: version,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&sandboxDir, "dir", "d", "./sandbox", "Sandbox directory path")
	rootCmd.PersistentFlags().StringVar(&shellFlag, "shell", "", "Shell to use in terminal (macOS/Linux only, e.g. zsh, /bin/fish)")
}

// ResolveShell determines the shell binary and args for the current platform.
// On Windows it always uses portable Git Bash (flag is ignored).
// On macOS/Linux: --shell flag → $SHELL → /bin/bash.
func ResolveShell(flag, absDir string) (bin string, args []string) {
	if runtime.GOOS == "windows" {
		bashPath := toolchain.GitBashPath(absDir, toolchain.DetectPlatform())
		if _, err := os.Stat(bashPath); err == nil {
			return bashPath, []string{"--login", "-i"}
		}
		return "powershell.exe", []string{"-NoLogo", "-NoProfile", "-NoExit"}
	}

	// macOS/Linux: flag → $SHELL → /bin/bash
	shell := flag
	if shell == "" {
		shell = os.Getenv("SHELL")
	}
	if shell == "" {
		shell = "/bin/bash"
	}

	// Support shorthand: "zsh" → resolve via PATH
	if !filepath.IsAbs(shell) {
		if p, err := exec.LookPath(shell); err == nil {
			shell = p
		}
	}

	return shell, []string{"--login", "-i"}
}
