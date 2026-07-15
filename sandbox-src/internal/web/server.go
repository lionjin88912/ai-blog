package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"path/filepath"

	"github.com/ai-sandbox/cli/internal/diag"
	"github.com/ai-sandbox/cli/internal/seed"
)

//go:embed static
var staticFiles embed.FS

// Serve starts the web terminal server on the given address.
// sandboxDir is used to locate winpty on Windows; its parent directory is the
// workspace root (where .agents/skills lives).
// shellPath and shellArgs define the shell to spawn for each WebSocket connection.
// env is the environment to pass to the shell process.
// geminiFlags is the extra flags to pass to the AI CLI (e.g. "--dangerously-skip-permissions").
// exeVersion is shown in the version badge and diagnostic bundle.
func Serve(addr, sandboxDir, shellPath string, shellArgs, env []string, geminiFlags, exeVersion string) error {
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("embed static: %w", err)
	}
	workspaceDir := filepath.Dir(sandboxDir)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(staticFS)))
	mux.HandleFunc("/ws", TerminalHandler(sandboxDir, shellPath, shellArgs, env))
	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"geminiFlags": geminiFlags})
	})
	mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"exe":        exeVersion,
			"seed":       seed.Version(),
			"seed_local": seed.LocalVersion(workspaceDir),
		})
	})
	mux.HandleFunc("/api/diagnostic", func(w http.ResponseWriter, r *http.Request) {
		data, name, err := diag.BuildZip(workspaceDir, exeVersion)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
		_, _ = w.Write(data)
		log.Printf("Diagnostic bundle exported: %s (%d bytes)", name, len(data))
	})

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	actualAddr := listener.Addr().String()
	fmt.Printf("🌐 AI Sandbox Terminal: http://%s\n", actualAddr)
	fmt.Println("   Press Ctrl+C to stop.")

	log.Printf("Serving on %s", actualAddr)
	return http.Serve(listener, mux)
}
