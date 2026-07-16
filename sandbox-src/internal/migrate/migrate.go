// Package migrate imports a business user's personas and WordPress connection
// settings from an OLD install into the current per-user data dir.
//
// Design goals:
//   - Non-destructive: the old folder is only ever read, never modified.
//   - Clean-only: wp-config.json is re-serialized through a known-key allowlist
//     and auth_method is inferred if missing, so a buggy/legacy format from the
//     old version can't carry its problems forward.
//   - No overwrite: a persona that already exists in the target is skipped.
package migrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// wpConfigKeys are the only keys copied from an old wp-config.json, in a
// stable order. Anything else (legacy/junk) is dropped.
var wpConfigKeys = []string{
	"auth_method",
	"WP_URL",
	"WP_USER",
	"WP_APP_PWD",
	"WP_CLIENT_ID",
	"WP_CLIENT_SECRET",
	"WP_ACCESS_TOKEN",
	"WP_BLOG_ID",
	"seo_plugin",
}

// Persona is a persona discovered in an old install.
type Persona struct {
	Slug       string
	Dir        string
	HasConfig  bool
	HasPersona bool
}

// Outcome is the per-persona result of a migration run.
type Outcome struct {
	Slug   string   `json:"slug"`
	Status string   `json:"status"` // imported | skipped-exists | error
	Notes  []string `json:"notes,omitempty"`
	Err    string   `json:"error,omitempty"`
}

// Report summarizes a migration run.
type Report struct {
	OldRoot   string    `json:"old_root"`
	Target    string    `json:"target"`
	DryRun    bool      `json:"dry_run"`
	Outcomes  []Outcome `json:"outcomes"`
	Imported  int       `json:"imported"`
	Skipped   int       `json:"skipped"`
	Errored   int       `json:"errored"`
}

// Scan finds personas under an old install root. It tolerates both the old
// Gemini-CLI layout (.gemini/skills/...) and the newer (.agents/skills/...),
// and a root that already points at a personas/ dir.
func Scan(oldRoot string) []Persona {
	var found []Persona
	seen := map[string]bool{}
	_ = filepath.Walk(oldRoot, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // unreadable subtree — skip, don't abort the whole scan
		}
		if !info.IsDir() {
			return nil
		}
		// A personas dir whose parent is persona-writer.
		if info.Name() == "personas" && filepath.Base(filepath.Dir(p)) == "persona-writer" {
			entries, _ := os.ReadDir(p)
			for _, e := range entries {
				if !e.IsDir() || e.Name() == "_template" {
					continue
				}
				dir := filepath.Join(p, e.Name())
				if seen[dir] {
					continue
				}
				seen[dir] = true
				_, cfgErr := os.Stat(filepath.Join(dir, "wp-config.json"))
				_, mdErr := os.Stat(filepath.Join(dir, "persona.md"))
				found = append(found, Persona{
					Slug:       e.Name(),
					Dir:        dir,
					HasConfig:  cfgErr == nil,
					HasPersona: mdErr == nil,
				})
			}
		}
		return nil
	})
	sort.Slice(found, func(i, j int) bool { return found[i].Slug < found[j].Slug })
	return found
}

// Migrate imports personas from oldRoot into targetPersonasDir. When dryRun is
// true it scans and reports without writing anything.
func Migrate(oldRoot, targetPersonasDir string, dryRun bool) (Report, error) {
	rep := Report{OldRoot: oldRoot, Target: targetPersonasDir, DryRun: dryRun}
	if _, err := os.Stat(oldRoot); err != nil {
		return rep, fmt.Errorf("找不到舊資料夾: %s", oldRoot)
	}

	for _, p := range Scan(oldRoot) {
		oc := Outcome{Slug: p.Slug}
		dst := filepath.Join(targetPersonasDir, p.Slug)

		if _, err := os.Stat(dst); err == nil {
			oc.Status = "skipped-exists"
			rep.Skipped++
			rep.Outcomes = append(rep.Outcomes, oc)
			continue
		}

		if err := importPersona(p, dst, dryRun, &oc); err != nil {
			oc.Status = "error"
			oc.Err = err.Error()
			rep.Errored++
		} else {
			oc.Status = "imported"
			rep.Imported++
		}
		rep.Outcomes = append(rep.Outcomes, oc)
	}
	return rep, nil
}

func importPersona(p Persona, dst string, dryRun bool, oc *Outcome) error {
	if !dryRun {
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return err
		}
	}

	// persona.md — copied verbatim (markdown, low risk).
	if p.HasPersona {
		if !dryRun {
			if err := copyFile(filepath.Join(p.Dir, "persona.md"), filepath.Join(dst, "persona.md")); err != nil {
				return fmt.Errorf("複製 persona.md: %w", err)
			}
		}
		oc.Notes = append(oc.Notes, "persona.md")
	}

	// wp-config.json — cleaned through the allowlist + auth_method inferred.
	if p.HasConfig {
		cleaned, notes, err := cleanWPConfig(filepath.Join(p.Dir, "wp-config.json"))
		if err != nil {
			oc.Notes = append(oc.Notes, "wp-config.json 略過("+err.Error()+")")
		} else {
			oc.Notes = append(oc.Notes, notes...)
			if !dryRun {
				if err := os.WriteFile(filepath.Join(dst, "wp-config.json"), cleaned, 0o644); err != nil {
					return fmt.Errorf("寫入 wp-config.json: %w", err)
				}
			}
		}
	}

	// published.json — kept only if it is valid JSON.
	src := filepath.Join(p.Dir, "published.json")
	if b, err := os.ReadFile(src); err == nil {
		if json.Valid(b) {
			if !dryRun {
				if err := os.WriteFile(filepath.Join(dst, "published.json"), b, 0o644); err != nil {
					return fmt.Errorf("寫入 published.json: %w", err)
				}
			}
			oc.Notes = append(oc.Notes, "published.json")
		} else {
			oc.Notes = append(oc.Notes, "published.json 損毀,已略過")
		}
	}
	// draft.json is intentionally NOT migrated — it is transient in-progress
	// state and the most likely to carry a broken old format.
	return nil
}

// cleanWPConfig reads an old wp-config.json and returns a re-serialized copy
// containing only known keys, with auth_method inferred when absent.
func cleanWPConfig(path string) ([]byte, []string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, nil, fmt.Errorf("JSON 格式錯誤")
	}

	var notes []string
	// Infer auth_method if missing or empty.
	if s, _ := raw["auth_method"].(string); s == "" {
		if _, ok := raw["WP_APP_PWD"]; ok {
			raw["auth_method"] = "application_password"
			notes = append(notes, "補上 auth_method=application_password")
		} else if hasAny(raw, "WP_ACCESS_TOKEN", "WP_CLIENT_ID", "WP_CLIENT_SECRET") {
			raw["auth_method"] = "oauth2"
			notes = append(notes, "補上 auth_method=oauth2")
		}
	}

	clean := map[string]any{}
	dropped := 0
	for k := range raw {
		if !allowed(k) {
			dropped++
		}
	}
	for _, k := range wpConfigKeys {
		if v, ok := raw[k]; ok {
			clean[k] = v
		}
	}
	if dropped > 0 {
		notes = append(notes, fmt.Sprintf("清掉 %d 個舊格式多餘欄位", dropped))
	}

	out, err := json.MarshalIndent(clean, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	notes = append([]string{"wp-config.json(已清理)"}, notes...)
	return out, notes, nil
}

// FindCandidates scans likely locations for old installs and returns roots that
// contain at least one persona, so the first-run UI can offer them directly.
func FindCandidates(searchDirs []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, base := range searchDirs {
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			root := filepath.Join(base, e.Name())
			if len(Scan(root)) > 0 && !seen[root] {
				seen[root] = true
				out = append(out, root)
			}
		}
	}
	return out
}

// DefaultSearchDirs returns the common per-user locations to look for old
// installs (Downloads, Desktop, Documents under the home dir).
func DefaultSearchDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		filepath.Join(home, "Downloads"),
		filepath.Join(home, "Desktop"),
		filepath.Join(home, "Documents"),
	}
}

func allowed(key string) bool {
	for _, k := range wpConfigKeys {
		if k == key {
			return true
		}
	}
	return false
}

func hasAny(m map[string]any, keys ...string) bool {
	for _, k := range keys {
		if _, ok := m[k]; ok {
			return true
		}
	}
	return false
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

// TargetPersonasDir returns the personas directory inside a data dir.
func TargetPersonasDir(dataDir string) string {
	return filepath.Join(dataDir, ".agents", "skills", "persona-writer", "personas")
}

// FormatReport renders a human-readable summary for the CLI/agent.
func FormatReport(rep Report) string {
	var sb strings.Builder
	verb := "匯入"
	if rep.DryRun {
		verb = "預掃(未實際寫入)"
	}
	fmt.Fprintf(&sb, "%s來源: %s\n", verb, rep.OldRoot)
	for _, o := range rep.Outcomes {
		switch o.Status {
		case "imported":
			fmt.Fprintf(&sb, "  ✅ %s — %s\n", o.Slug, strings.Join(o.Notes, "、"))
		case "skipped-exists":
			fmt.Fprintf(&sb, "  ⏭️  %s — 已存在,跳過\n", o.Slug)
		case "error":
			fmt.Fprintf(&sb, "  ❌ %s — %s\n", o.Slug, o.Err)
		}
	}
	fmt.Fprintf(&sb, "小計:匯入 %d、跳過 %d、失敗 %d\n", rep.Imported, rep.Skipped, rep.Errored)
	return sb.String()
}
