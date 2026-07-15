package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ai-sandbox/cli/cmd"
	"github.com/ai-sandbox/cli/internal/config"
	"github.com/ai-sandbox/cli/internal/toolchain"
	"github.com/ai-sandbox/cli/internal/web"
)

// openBrowser tries to open the default browser with retry logic
func openBrowser(url string) {
	maxRetries := 10
	initialWait := 500 * time.Millisecond

	time.Sleep(initialWait)

	client := &http.Client{Timeout: 2 * time.Second}
	for i := 0; i < maxRetries; i++ {
		log.Printf("Checking if server is ready (attempt %d/%d)...", i+1, maxRetries)
		_, err := client.Get(url)
		if err == nil {
			log.Printf("Server is ready, opening browser: %s", url)
			launchBrowser(url)
			return
		}
		waitTime := initialWait * time.Duration(i+1)
		time.Sleep(waitTime)
	}

	log.Printf("Server may not be fully ready, attempting to open browser anyway: %s", url)
	launchBrowser(url)
}

func launchBrowser(url string) {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		c = exec.Command("cmd", "/c", "start", "chrome", "--new-window", url)
		if err := c.Start(); err != nil {
			c = exec.Command("cmd", "/c", "start", url)
			_ = c.Start()
		}
	case "darwin":
		c = exec.Command("open", "-n", "-a", "Google Chrome", "--args", "--new-window", url)
		if err := c.Start(); err != nil {
			c = exec.Command("open", "-n", url)
			_ = c.Start()
		}
	default:
		c = exec.Command("google-chrome", "--new-window", url)
		if err := c.Start(); err != nil {
			c = exec.Command("xdg-open", url)
			_ = c.Start()
		}
	}
}

func main() {
	// If args provided (CLI usage), use cobra commands
	if len(os.Args) > 1 {
		cmd.Execute()
		return
	}

	// --- Double-click mode: start web server directly ---

	// Change CWD to the exe's directory so "./sandbox" resolves correctly.
	// If sandbox is not next to the exe, try the parent directory (e.g. exe in dist/).
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		_ = os.Chdir(exeDir)

		sandboxCheck := filepath.Join(exeDir, "sandbox", "bin")
		if _, err := os.Stat(sandboxCheck); os.IsNotExist(err) {
			parentDir := filepath.Dir(exeDir)
			parentCheck := filepath.Join(parentDir, "sandbox", "bin")
			if _, err := os.Stat(parentCheck); err == nil {
				_ = os.Chdir(parentDir)
			}
		}
	}

	// Auto-init config if not configured
	workspacePath, _ := filepath.Abs("./workspace")
	if err := config.QuickInit(workspacePath); err != nil {
		log.Fatalf("Auto-init failed: %v", err)
	}

	sandboxDir := "./sandbox"
	absDir, err := filepath.Abs(sandboxDir)
	if err != nil {
		log.Fatalf("Failed to resolve sandbox directory: %v", err)
	}

	// Auto-setup if sandbox doesn't exist
	binDir := toolchain.SandboxBinDir(absDir)
	p := toolchain.DetectPlatform()
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		log.Println("Sandbox not found, running setup...")
		if err := os.MkdirAll(absDir, 0755); err != nil {
			log.Fatalf("Create sandbox dir: %v", err)
		}
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
				log.Fatalf("%s: %v", step.name, err)
			}
			fmt.Println()
		}
		log.Println("✅ Sandbox ready!")
	} else {
		// Sandbox exists — regenerate shims to fix stale absolute paths
		if err := toolchain.WriteShims(absDir, p); err != nil {
			log.Printf("⚠️  Shim regeneration failed: %v", err)
		}
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

	shellBin, shellArgs := cmd.ResolveShell("", absDir)

	if runtime.GOOS == "windows" && shellBin != "powershell.exe" {
		env = append(env, "SHELL="+shellBin)
		env = append(env, "MSYSTEM=MINGW64")
	}

	env = append(env, "TERM=xterm-256color")
	env = append(env, "LANG=en_US.UTF-8")
	env = append(env, "LC_ALL=en_US.UTF-8")

	port := "8088"
	addr := "127.0.0.1:" + port
	url := fmt.Sprintf("http://%s", addr)

	// Open browser with retry in background
	go openBrowser(url)

	// Start web server (blocks until Ctrl+C)
	log.Printf("🌐 AI Sandbox Terminal: %s", url)
	log.Println("   Press Ctrl+C to stop.")
	log.Fatal(web.Serve(addr, absDir, shellBin, shellArgs, env, ""))
}
