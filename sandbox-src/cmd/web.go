package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ai-sandbox/cli/internal/seed"
	"github.com/ai-sandbox/cli/internal/toolchain"
	"github.com/ai-sandbox/cli/internal/web"
	"github.com/spf13/cobra"
)

var (
	webPort             string
	skipPermissionsFlag bool
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Open a browser-based terminal with sandbox tools",
	Long:  `Start a local web server with an xterm.js terminal connected to the sandbox shell.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		absDir, err := filepath.Abs(sandboxDir)
		if err != nil {
			return err
		}

		// First-run seeding: lay down factory skills/docs into the workspace
		// (parent of sandbox/). Existing files are never touched.
		seed.EnsureReport(filepath.Dir(absDir), func(f string, a ...any) { fmt.Printf(f+"\n", a...) })

		binDir := toolchain.SandboxBinDir(absDir)
		if _, err := os.Stat(binDir); os.IsNotExist(err) {
			return fmt.Errorf("sandbox not set up. Run 'ai-blog setup' first")
		}

		// Refresh shims on every launch so a new exe's fixes (e.g. the pwsh
		// shim) apply without a reinstall — same as the double-click path.
		if err := toolchain.WriteShims(absDir, toolchain.DetectPlatform()); err != nil {
			fmt.Printf("⚠️  Shim regeneration failed: %v\n", err)
		}

		// Build PATH: sandbox/bin + git/mingw64/bin + git/usr/bin + system PATH
		gitDir := filepath.Join(absDir, "git")
		newPath := binDir
		if runtime.GOOS == "windows" {
			newPath += string(os.PathListSeparator) + filepath.Join(gitDir, "mingw64", "bin")
			newPath += string(os.PathListSeparator) + filepath.Join(gitDir, "usr", "bin")
		}
		newPath += string(os.PathListSeparator) + os.Getenv("PATH")
		env := append(os.Environ(), "PATH="+newPath)

		shellBin, shellArgs := ResolveShell(shellFlag, absDir)

		if runtime.GOOS == "windows" && shellBin != "powershell.exe" {
			env = append(env, "SHELL="+shellBin)
			env = append(env, "MSYSTEM=MINGW64")
		}

		// Set TERM for color support and UTF-8 locale for Chinese
		env = append(env, "TERM=xterm-256color")
		env = append(env, "LANG=en_US.UTF-8")
		env = append(env, "LC_ALL=en_US.UTF-8")

		addr := "127.0.0.1:" + webPort

		geminiFlags := ""
		if skipPermissionsFlag {
			geminiFlags = "--dangerously-skip-permissions"
		}

		// Auto-open browser
		go openBrowser("http://" + addr)

		return web.Serve(addr, absDir, shellBin, shellArgs, env, geminiFlags, version)
	},
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func init() {
	webCmd.Flags().StringVarP(&webPort, "port", "p", "8088", "Port to serve on")
	webCmd.Flags().BoolVar(&skipPermissionsFlag, "dangerously-skip-permissions", false, "Pass --dangerously-skip-permissions to Gemini CLI")
	rootCmd.AddCommand(webCmd)
}
