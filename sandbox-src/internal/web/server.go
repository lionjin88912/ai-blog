package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
)

//go:embed static
var staticFiles embed.FS

// Serve starts the web terminal server on the given address.
// sandboxDir is used to locate winpty on Windows.
// shellPath and shellArgs define the shell to spawn for each WebSocket connection.
// env is the environment to pass to the shell process.
// geminiFlags is the extra flags to pass to the gemini CLI (e.g. "--dangerously-skip-permissions").
func Serve(addr, sandboxDir, shellPath string, shellArgs, env []string, geminiFlags string) error {
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("embed static: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(staticFS)))
	mux.HandleFunc("/ws", TerminalHandler(sandboxDir, shellPath, shellArgs, env))
	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"geminiFlags": geminiFlags})
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
