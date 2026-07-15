// Package diag builds a one-click diagnostic bundle (zip) that business
// users can download from the web terminal and hand to support/AI for
// debugging.
//
// Hard privacy rule: credentials NEVER enter the bundle. wp-config.json is
// re-serialized with every secret-looking value masked; token/credential
// files are never read at all.
package diag

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/ai-sandbox/cli/internal/seed"
)

var secretKeyRe = regexp.MustCompile(`(?i)(pass|token|secret|credential|api[_-]?key|auth)`)

const logTailBytes = 300 * 1024

// BuildZip collects diagnostics for the workspace and returns the zip bytes
// plus a suggested filename.
func BuildZip(workspaceDir, exeVersion string) ([]byte, string, error) {
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)

	addJSON := func(name string, v any) {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			b = []byte(fmt.Sprintf(`{"error":%q}`, err.Error()))
		}
		if w, err := zw.Create(name); err == nil {
			_, _ = w.Write(b)
		}
	}
	addBytes := func(name string, b []byte) {
		if w, err := zw.Create(name); err == nil {
			_, _ = w.Write(b)
		}
	}

	host, _ := os.Hostname()
	addJSON("manifest.json", map[string]any{
		"generated_at":          time.Now().Format(time.RFC3339),
		"hostname":              host,
		"os":                    runtime.GOOS,
		"arch":                  runtime.GOARCH,
		"exe_version":           exeVersion,
		"seed_version_embedded": seed.Version(),
		"seed_version_local":    seed.LocalVersion(workspaceDir),
		"workspace":             workspaceDir,
		"note":                  "credentials are masked by design; wp-config passwords never leave the machine",
	})

	addJSON("skills_inventory.json", skillsInventory(workspaceDir))
	addJSON("personas.json", personas(workspaceDir))
	addJSON("sandbox_env.json", sandboxEnv(workspaceDir))

	// Workspace docs the agent actually loads.
	for _, f := range []string{"GEMINI.md", "TOOLS.md", "HOWTO.md"} {
		if b, err := os.ReadFile(filepath.Join(workspaceDir, f)); err == nil {
			addBytes("workspace/"+f, b)
		}
	}

	// Antigravity CLI state (no tokens live in these files).
	home, _ := os.UserHomeDir()
	agyDir := filepath.Join(home, ".gemini", "antigravity-cli")
	if b, err := os.ReadFile(filepath.Join(agyDir, "settings.json")); err == nil {
		addBytes("agy/settings.json", b)
	}
	if b := tailFile(filepath.Join(agyDir, "cli.log"), logTailBytes); b != nil {
		addBytes("agy/cli_log_tail.txt", b)
	}
	if names := dirNames(filepath.Join(home, ".gemini", "config", "skills")); names != nil {
		addJSON("agy/global_skills.json", names)
	}

	if err := zw.Close(); err != nil {
		return nil, "", err
	}
	name := fmt.Sprintf("ai-sandbox-diag-%s-%s.zip", sanitizeName(host), time.Now().Format("20060102-150405"))
	return buf.Bytes(), name, nil
}

type fileEntry struct {
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	MTime      string `json:"mtime"`
	SHA        string `json:"sha256_12"`
	SeedStatus string `json:"seed_status"` // same | modified | extra
}

func skillsInventory(workspaceDir string) map[string]any {
	seedHashes := seed.ContentHashes()
	seen := map[string]bool{}
	var files []fileEntry

	root := filepath.Join(workspaceDir, ".agents", "skills")
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel := ".agents/skills/" + filepath.ToSlash(strings.TrimPrefix(p, root+string(os.PathSeparator)))
		if strings.Contains(rel, "__pycache__") {
			return nil
		}
		e := fileEntry{Path: rel, Size: info.Size(), MTime: info.ModTime().Format(time.RFC3339)}
		// Never hash-or-flag credentials; just note presence.
		base := filepath.Base(rel)
		if base == "wp-config.json" {
			e.SHA = "(masked)"
			e.SeedStatus = "user-state"
			files = append(files, e)
			return nil
		}
		if b, err := os.ReadFile(p); err == nil {
			sum := sha256.Sum256(b)
			e.SHA = fmt.Sprintf("%x", sum)[:12]
		}
		if want, ok := seedHashes[rel]; ok {
			seen[rel] = true
			if want == e.SHA {
				e.SeedStatus = "same"
			} else {
				e.SeedStatus = "modified"
			}
		} else {
			e.SeedStatus = "extra"
		}
		files = append(files, e)
		return nil
	})

	var missing []string
	for rel := range seedHashes {
		if strings.HasPrefix(rel, ".agents/skills/") && !seen[rel] {
			missing = append(missing, rel)
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	sort.Strings(missing)
	return map[string]any{"files": files, "missing_vs_seed": missing}
}

func personas(workspaceDir string) []map[string]any {
	root := filepath.Join(workspaceDir, ".agents", "skills", "persona-writer", "personas")
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var out []map[string]any
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		p := map[string]any{"slug": e.Name()}
		_, err := os.Stat(filepath.Join(dir, "persona.md"))
		p["has_persona_md"] = err == nil
		p["wp_config"] = maskedJSON(filepath.Join(dir, "wp-config.json"))
		p["published"] = rawJSON(filepath.Join(dir, "published.json"))
		if d := rawJSON(filepath.Join(dir, "draft.json")); d != nil {
			if m, ok := d.(map[string]any); ok {
				p["draft_stage"] = m["stage"]
			}
		}
		out = append(out, p)
	}
	return out
}

// maskedJSON parses a JSON file and masks every secret-looking string value.
func maskedJSON(path string) any {
	v := rawJSON(path)
	if v == nil {
		return nil
	}
	return maskValue("", v)
}

func maskValue(key string, v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = maskValue(k, val)
		}
		return out
	case []any:
		for i, val := range t {
			t[i] = maskValue(key, val)
		}
		return t
	case string:
		if secretKeyRe.MatchString(key) && t != "" {
			return "****(已設定)"
		}
		return t
	default:
		return v
	}
}

func rawJSON(path string) any {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var v any
	if json.Unmarshal(b, &v) != nil {
		return map[string]any{"parse_error": true, "size": len(b)}
	}
	return v
}

func sandboxEnv(workspaceDir string) map[string]any {
	sb := filepath.Join(workspaceDir, "sandbox")
	return map[string]any{
		"bin":         dirNames(filepath.Join(sb, "bin")),
		"antigravity": dirNames(filepath.Join(sb, "antigravity")),
		"node":        dirNames(filepath.Join(sb, "node")),
		"venv_exists": exists(filepath.Join(workspaceDir, ".venv")),
	}
}

func dirNames(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func tailFile(path string, n int64) []byte {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil
	}
	if st.Size() > n {
		_, _ = f.Seek(-n, io.SeekEnd)
	}
	b, _ := io.ReadAll(f)
	return b
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func sanitizeName(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' {
			return r
		}
		return '-'
	}, s)
}
