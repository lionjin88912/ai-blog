// Package backup exports and imports a portable bundle of a user's personas +
// WordPress connection settings + publish history, so a new hire on a blank
// install can take over a departing employee's AI Blog work.
//
// The bundle is a zip laid out like a mini install
// (.agents/skills/persona-writer/personas/<slug>/...), so import can reuse the
// migrate engine verbatim — same per-file merge, no-overwrite, wp-config
// cleaning and auth_method inference. draft.json is intentionally excluded
// (transient). The bundle contains real credentials by design; the UI warns
// the user to move it over a trusted channel and delete it afterwards.
package backup

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ai-sandbox/cli/internal/migrate"
)

const personasPrefix = ".agents/skills/persona-writer/personas"

// files copied per persona (draft.json deliberately omitted).
var personaFiles = []string{"persona.md", "wp-config.json", "published.json"}

// Manifest describes a backup bundle.
type Manifest struct {
	CreatedAt string   `json:"created_at"`
	Hostname  string   `json:"hostname"`
	Personas  []string `json:"personas"`
	Tool      string   `json:"tool"`
}

// Export zips the workspace's personas into a portable backup. Returns the zip
// bytes and a suggested filename. now is passed in so the caller controls the
// timestamp.
func Export(dataDir string, now time.Time) ([]byte, string, error) {
	personasDir := migrate.TargetPersonasDir(dataDir)
	entries, err := os.ReadDir(personasDir)
	if err != nil {
		return nil, "", fmt.Errorf("找不到人格資料夾(這台還沒有任何人格?): %w", err)
	}

	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	var slugs []string

	for _, e := range entries {
		if !e.IsDir() || e.Name() == "_template" {
			continue
		}
		slug := e.Name()
		wrote := false
		for _, f := range personaFiles {
			src := filepath.Join(personasDir, slug, f)
			data, err := os.ReadFile(src)
			if err != nil {
				continue // file absent for this persona — fine
			}
			w, err := zw.Create(personasPrefix + "/" + slug + "/" + f)
			if err != nil {
				return nil, "", err
			}
			if _, err := w.Write(data); err != nil {
				return nil, "", err
			}
			wrote = true
		}
		if wrote {
			slugs = append(slugs, slug)
		}
	}

	if len(slugs) == 0 {
		return nil, "", fmt.Errorf("沒有可匯出的人格")
	}

	host, _ := os.Hostname()
	man := Manifest{
		CreatedAt: now.Format(time.RFC3339),
		Hostname:  host,
		Personas:  slugs,
		Tool:      "ai-blog",
	}
	mb, _ := json.MarshalIndent(man, "", "  ")
	if w, err := zw.Create("backup-manifest.json"); err == nil {
		_, _ = w.Write(mb)
	}

	if err := zw.Close(); err != nil {
		return nil, "", err
	}
	name := fmt.Sprintf("ai-blog-backup-%s-%s.zip", sanitize(host), now.Format("20060102-150405"))
	return buf.Bytes(), name, nil
}

// Import extracts a backup zip into a temp dir and runs it through the migrate
// engine, so all the merge/clean/no-overwrite guarantees apply. The workspace's
// existing personas are never overwritten.
func Import(zipData []byte, dataDir string) (migrate.Report, error) {
	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return migrate.Report{}, fmt.Errorf("備份檔不是有效的 zip: %w", err)
	}

	tmp, err := os.MkdirTemp("", "ai-blog-import-*")
	if err != nil {
		return migrate.Report{}, err
	}
	defer os.RemoveAll(tmp)

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		// zip-slip guard: reject entries that escape tmp.
		dst := filepath.Join(tmp, filepath.FromSlash(f.Name))
		if !strings.HasPrefix(dst, filepath.Clean(tmp)+string(os.PathSeparator)) {
			return migrate.Report{}, fmt.Errorf("備份檔含不安全路徑: %s", f.Name)
		}
		if err := extractOne(f, dst); err != nil {
			return migrate.Report{}, err
		}
	}

	return migrate.Migrate(tmp, migrate.TargetPersonasDir(dataDir), false)
}

// UploadFile is one file from a folder-based import (browser webkitdirectory):
// RelPath is the path the browser reported (…/personas/<slug>/<file>).
type UploadFile struct {
	RelPath string
	Data    []byte
}

// ImportFiles imports personas from a set of uploaded files (a folder the user
// picked in the browser). It keeps only the persona files, lays them out in a
// temp dir, and runs them through the migrate engine — same guarantees as the
// zip path. Files outside a persona-writer/personas/<slug>/ path are ignored.
func ImportFiles(files []UploadFile, dataDir string) (migrate.Report, error) {
	tmp, err := os.MkdirTemp("", "ai-blog-import-*")
	if err != nil {
		return migrate.Report{}, err
	}
	defer os.RemoveAll(tmp)

	kept := 0
	for _, f := range files {
		rel := keepPersonaPath(f.RelPath)
		if rel == "" {
			continue // not a persona file we care about
		}
		dst := filepath.Join(tmp, filepath.FromSlash(rel))
		if !strings.HasPrefix(dst, filepath.Clean(tmp)+string(os.PathSeparator)) {
			continue // zip-slip / traversal guard
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return migrate.Report{}, err
		}
		if err := os.WriteFile(dst, f.Data, 0o644); err != nil {
			return migrate.Report{}, err
		}
		kept++
	}
	if kept == 0 {
		return migrate.Report{}, fmt.Errorf("這個資料夾裡找不到人格資料(persona-writer/personas)")
	}
	return migrate.Migrate(tmp, migrate.TargetPersonasDir(dataDir), false)
}

// keepPersonaPath maps an arbitrary uploaded path to a normalized
// persona-writer/personas/<slug>/<file> path, or "" if it isn't a persona file
// we migrate (persona.md / wp-config.json / published.json).
func keepPersonaPath(rel string) string {
	rel = filepath.ToSlash(rel)
	i := strings.Index(rel, "persona-writer/personas/")
	if i == -1 {
		return ""
	}
	tail := rel[i+len("persona-writer/personas/"):] // want exactly "<slug>/<file>"
	parts := strings.Split(tail, "/")
	if len(parts) != 2 || parts[0] == "" || parts[0] == "_template" {
		return "" // not directly under a persona dir (ignore nested/odd paths)
	}
	slug, base := parts[0], parts[1]
	for _, f := range personaFiles {
		if base == f {
			return "persona-writer/personas/" + slug + "/" + base
		}
	}
	return ""
}

func extractOne(f *zip.File, dst string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	buf := make([]byte, 32*1024)
	for {
		n, rerr := rc.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if rerr != nil {
			break
		}
	}
	return nil
}

func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' {
			return r
		}
		return '-'
	}, s)
}
