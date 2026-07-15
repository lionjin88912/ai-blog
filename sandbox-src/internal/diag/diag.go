// Package diag builds a one-click diagnostic bundle (zip) that business
// users can download from the web terminal and hand to support/AI for
// debugging.
//
// Hard privacy rule: credentials NEVER enter the bundle. Credential-bearing
// JSON (wp-config.json) is masked with an ALLOWLIST — only keys known to be
// non-sensitive survive in cleartext; every other value (including any future
// key we haven't seen) is masked by default. Free-form text that we cannot
// structure (cli.log, settings.json) is run through a secret scrubber before
// it enters the zip.
package diag

import (
	"archive/zip"
	"bufio"
	"bytes"
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

// configSafeKeys are the ONLY wp-config keys emitted in cleartext. Anything
// not listed here is masked, so a newly-added credential key (e.g. WP_APP_PWD,
// which the old substring regex missed) is safe by default. Deliberately
// excludes WP_USER — account + any leaked secret half = a working credential.
var configSafeKeys = map[string]bool{
	"auth_method": true,
	"wp_url":      true,
	"site_url":    true,
	"url":         true,
	"wp_blog_id":  true,
	"blog_id":     true,
	"seo_plugin":  true,
	"slug":        true,
}

// logSecretRe scrubs whole lines from unstructured text (logs, settings) that
// look like they carry a secret: bearer tokens, long opaque hex/base64 blobs,
// or key=value / "key": "value" pairs whose key name smells sensitive.
var logSecretRe = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(authorization|bearer|pass|pwd|token|secret|credential|api[_-]?key)\S*["']?\s*[:=]\s*\S+`),
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._\-]+`),
	regexp.MustCompile(`\b[A-Za-z0-9_\-]{40,}\b`), // long opaque blobs (tokens/keys)
}

const logTailBytes = 300 * 1024

// BuildZip collects diagnostics for the workspace and returns the zip bytes
// plus a suggested filename.
func BuildZip(workspaceDir, exeVersion string) ([]byte, string, error) {
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	var writeErrs []string

	addBytes := func(name string, b []byte) {
		w, err := zw.Create(name)
		if err == nil {
			_, err = w.Write(b)
		}
		if err != nil {
			writeErrs = append(writeErrs, fmt.Sprintf("%s: %v", name, err))
		}
	}
	addJSON := func(name string, v any) {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			b = []byte(fmt.Sprintf(`{"error":%q}`, err.Error()))
		}
		addBytes(name, b)
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
		"note":                  "credentials are masked by design; wp-config uses an allowlist, free-form text is secret-scrubbed",
	})

	addJSON("skills_inventory.json", skillsInventory(workspaceDir))
	addJSON("personas.json", personas(workspaceDir))
	addJSON("sandbox_env.json", sandboxEnv(workspaceDir))

	// Workspace docs the agent actually loads. Scrubbed in case a credential
	// was pasted into an SOP/doc during local evolution.
	for _, f := range []string{"GEMINI.md", "TOOLS.md", "HOWTO.md"} {
		if b, err := os.ReadFile(filepath.Join(workspaceDir, f)); err == nil {
			addBytes("workspace/"+f, scrubText(b))
		}
	}

	// Antigravity CLI state. Not structurally parseable for secrets, so
	// everything here is secret-scrubbed before it enters the bundle.
	home, _ := os.UserHomeDir()
	agyDir := filepath.Join(home, ".gemini", "antigravity-cli")
	if b, err := os.ReadFile(filepath.Join(agyDir, "settings.json")); err == nil {
		addBytes("agy/settings.json", scrubText(b))
	}
	if b := tailFile(filepath.Join(agyDir, "cli.log"), logTailBytes); b != nil {
		addBytes("agy/cli_log_tail.txt", scrubText(b))
	}
	if names := dirNames(filepath.Join(home, ".gemini", "config", "skills")); names != nil {
		addJSON("agy/global_skills.json", names)
	}

	if len(writeErrs) > 0 {
		addBytes("_diag_errors.txt", []byte(strings.Join(writeErrs, "\n")+"\n"))
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
		b, readErr := os.ReadFile(p)
		if readErr != nil {
			// Present but unreadable — say so rather than mislabeling.
			e.SHA = "(unreadable)"
			e.SeedStatus = "unreadable"
			files = append(files, e)
			if _, ok := seedHashes[rel]; ok {
				seen[rel] = true
			}
			return nil
		}
		e.SHA = seed.HashBytes(b)
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
		p["wp_config"] = maskedConfig(filepath.Join(dir, "wp-config.json"))
		p["published"] = rawJSON(filepath.Join(dir, "published.json"))
		if m, ok := rawJSON(filepath.Join(dir, "draft.json")).(map[string]any); ok {
			p["draft_stage"] = m["stage"]
		}
		out = append(out, p)
	}
	return out
}

// maskedConfig parses a credential-bearing JSON file (wp-config.json) and
// keeps ONLY allowlisted keys in cleartext; every other non-empty string is
// masked. Allowlist-by-default means an unrecognized key can never leak.
func maskedConfig(path string) any {
	v := rawJSON(path)
	if v == nil {
		return nil
	}
	return maskAllowlist("", v)
}

func maskAllowlist(key string, v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = maskAllowlist(k, val)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, val := range t {
			out[i] = maskAllowlist(key, val)
		}
		return out
	case string:
		if t != "" && !configSafeKeys[strings.ToLower(key)] {
			return "****(已設定)"
		}
		return t
	default:
		return v
	}
}

// scrubText masks secret-looking content from unstructured text before it
// enters the bundle. Applied to logs and settings we cannot structurally mask.
func scrubText(b []byte) []byte {
	var out bytes.Buffer
	sc := bufio.NewScanner(bytes.NewReader(b))
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		for _, re := range logSecretRe {
			line = re.ReplaceAllString(line, "****(redacted)")
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	return out.Bytes()
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
