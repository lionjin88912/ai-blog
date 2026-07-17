package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/ai-sandbox/cli/internal/backup"
	"github.com/ai-sandbox/cli/internal/diag"
	"github.com/ai-sandbox/cli/internal/migrate"
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
		// Block DNS-rebinding: only loopback Host values may pull the bundle.
		if !isLoopbackHost(r.Host) {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
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

	// /api/migrate/scan — list old installs found under the usual locations,
	// with persona counts. Only offered on a first run (see the web UI).
	mux.HandleFunc("/api/migrate/scan", func(w http.ResponseWriter, r *http.Request) {
		if !isLoopbackHost(r.Host) {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"candidates": migrate.FindCandidates(migrate.DefaultSearchDirs()),
		})
	})

	// /api/migrate/run?from=<path> — import personas from an old install into
	// this workspace's data dir. Loopback-only; the old folder is only read.
	mux.HandleFunc("/api/migrate/run", func(w http.ResponseWriter, r *http.Request) {
		if !isLoopbackHost(r.Host) {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		from := r.URL.Query().Get("from")
		if from == "" {
			http.Error(w, "missing from", http.StatusBadRequest)
			return
		}
		rep, err := migrate.Migrate(from, migrate.TargetPersonasDir(workspaceDir), false)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(rep)
		log.Printf("Migrate from %s: imported=%d merged=%d skipped=%d errored=%d",
			from, rep.Imported, rep.Merged, rep.Skipped, rep.Errored)
	})

	// /api/backup/export — download a portable backup zip (personas + WP
	// settings + publish history) to hand work over to another machine.
	mux.HandleFunc("/api/backup/export", func(w http.ResponseWriter, r *http.Request) {
		if !isLoopbackHost(r.Host) {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		data, name, err := backup.Export(workspaceDir, time.Now())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
		_, _ = w.Write(data)
		log.Printf("Backup exported: %s (%d bytes)", name, len(data))
	})

	// /api/backup/import — upload a backup zip and merge it into this
	// workspace (reuses the migrate engine: clean + no-overwrite).
	mux.HandleFunc("/api/backup/import", func(w http.ResponseWriter, r *http.Request) {
		if !isLoopbackHost(r.Host) {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "bad upload", http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file", http.StatusBadRequest)
			return
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		rep, err := backup.Import(data, workspaceDir)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(rep)
		log.Printf("Backup imported (zip): imported=%d merged=%d skipped=%d errored=%d",
			rep.Imported, rep.Merged, rep.Skipped, rep.Errored)
	})

	// /api/backup/import-folder — import from a folder the user picked in the
	// browser (webkitdirectory): each uploaded file part's filename is its
	// relative path. Only persona files are kept; reuses the migrate engine.
	mux.HandleFunc("/api/backup/import-folder", func(w http.ResponseWriter, r *http.Request) {
		if !isLoopbackHost(r.Host) {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			http.Error(w, "bad upload", http.StatusBadRequest)
			return
		}
		var files []backup.UploadFile
		// The relative path is carried in the form FIELD NAME (map key), not the
		// filename — Go runs FileHeader.Filename through filepath.Base, which
		// would strip the path we need.
		for name, headers := range r.MultipartForm.File {
			for _, fh := range headers {
				f, err := fh.Open()
				if err != nil {
					continue
				}
				data, err := io.ReadAll(f)
				f.Close()
				if err != nil {
					continue
				}
				files = append(files, backup.UploadFile{RelPath: name, Data: data})
			}
		}
		rep, err := backup.ImportFiles(files, workspaceDir)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(rep)
		log.Printf("Backup imported (folder): imported=%d merged=%d skipped=%d errored=%d",
			rep.Imported, rep.Merged, rep.Skipped, rep.Errored)
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

// isLoopbackHost reports whether the request Host targets localhost/127.0.0.1,
// defeating DNS-rebinding attempts to reach the diagnostic endpoint from a
// malicious page (which would carry an attacker-controlled Host).
func isLoopbackHost(host string) bool {
	h := host
	if i := strings.LastIndex(h, ":"); i != -1 {
		h = h[:i]
	}
	h = strings.Trim(h, "[]")
	if h == "localhost" {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}
