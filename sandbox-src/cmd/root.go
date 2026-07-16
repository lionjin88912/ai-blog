package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ai-sandbox/cli/internal/config"
	"github.com/ai-sandbox/cli/internal/toolchain"
	"github.com/spf13/cobra"
)

var (
	version    = "0.1.0"
	sandboxDir string
	shellFlag  string
)

var rootCmd = &cobra.Command{
	Use:   "ai-blog",
	Short: "AI Blog 部落格 — 行銷內容自動化",
	Long: `AI Blog 部落格 — a local, portable environment for the marketing
content factory: Node 22, Python 3.12, Antigravity CLI, GitHub Copilot CLI, uv.

Get started:
  ai-blog init     Configure API keys and workspace
  ai-blog setup    Download tools to ./sandbox/
  ai-blog shell    Open a shell with sandbox tools in PATH`,
	Version: version,
}

// origWD is the working directory at process start, captured before
// enterDataDir chdir's away. Commands that take a user-supplied relative path
// (e.g. `migrate --from ./old`) must resolve it against this, not the data dir.
var origWD string

// OriginalWD returns the directory the user launched the command from.
func OriginalWD() string { return origWD }

// enterDataDir moves into the per-user data directory (config.DataDir) unless
// the user explicitly picked a --dir. This is what lets the exe live in
// Downloads and act as a pure launcher: the .bat entry point runs `web`, which
// lands here just like the double-click path in main.go, so nothing scatters
// next to the exe.
func enterDataDir(cmd *cobra.Command, args []string) error {
	if wd, err := os.Getwd(); err == nil {
		origWD = wd
	}
	if rootCmd.PersistentFlags().Changed("dir") {
		return nil // developer explicitly chose a sandbox location
	}
	dataDir, err := config.DataDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("create data directory %s: %w", dataDir, err)
	}
	if err := os.Chdir(dataDir); err != nil {
		return fmt.Errorf("enter data directory %s: %w", dataDir, err)
	}
	log.Printf("📁 資料夾: %s", dataDir)
	return nil
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentPreRunE = enterDataDir
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
