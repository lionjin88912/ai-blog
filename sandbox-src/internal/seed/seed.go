// Package seed embeds the factory content (skills, GEMINI.md, docs) into the
// binary and materializes it into a workspace on first run.
//
// Design contract (matches the marketing-factory iteration model):
//   - First run: lay down every missing file. Never touch existing ones.
//   - refresh-skills: overwrite SOP files with the embedded version, but NEVER
//     touch persona user data (anything under personas/ except _template/).
//     Overwritten files are backed up under .agents/_backup/<stamp>/ first.
//   - Runtime/user state (wp-config.json, published.json, articles/) is
//     excluded at build time by `make seed-sync`, and defensively skipped
//     here as well.
package seed

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// content/ is populated by `make seed-sync` from the repo root before build.
// Only README.md and SEED_VERSION live there in a fresh checkout; a binary
// built without seed-sync simply has an empty seed (engine-only, still works).
//
//go:embed all:content
var contentFS embed.FS

const contentRoot = "content"

// metadata files at the content root that must not be materialized.
var metaFiles = map[string]bool{
	"README.md":    true,
	"SEED_VERSION": true,
}

// Stats reports what Materialize did.
type Stats struct {
	Created []string
	Updated []string
	Skipped int
}

// Version returns the embedded seed version, or "dev" when built without
// a synced seed.
func Version() string {
	b, err := contentFS.ReadFile(contentRoot + "/SEED_VERSION")
	if err != nil {
		return "dev"
	}
	return strings.TrimSpace(string(b))
}

// LocalVersion returns the seed version marker previously written into the
// workspace, or "" if none.
func LocalVersion(targetDir string) string {
	b, err := os.ReadFile(filepath.Join(targetDir, ".agents", ".seed-version"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// isProtected reports whether rel (slash-separated, relative to workspace
// root) is persona user data that must never be overwritten.
func isProtected(rel string) bool {
	const personas = ".agents/skills/persona-writer/personas/"
	if !strings.HasPrefix(rel, personas) {
		return false
	}
	return !strings.HasPrefix(rel, personas+"_template/")
}

// isUserState reports runtime files that must never be written even when
// missing (they are per-machine state, not factory content).
func isUserState(rel string) bool {
	base := filepath.Base(rel)
	if base == "wp-config.json" || base == "published.json" {
		return true
	}
	return strings.Contains(rel, "/articles/")
}

// Materialize lays the embedded seed into targetDir.
//
// refresh=false (first run): only create missing files.
// refresh=true (refresh-skills): also overwrite non-protected files whose
// content differs, backing up the old file first.
func Materialize(targetDir string, refresh bool) (Stats, error) {
	var st Stats
	stamp := Version() + "-" + time.Now().Format("20060102-150405")
	backupRoot := filepath.Join(targetDir, ".agents", "_backup", stamp)

	err := fs.WalkDir(contentFS, contentRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := strings.TrimPrefix(p, contentRoot+"/")
		if metaFiles[rel] {
			return nil
		}
		if isUserState(rel) {
			st.Skipped++
			return nil
		}

		data, err := contentFS.ReadFile(p)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", rel, err)
		}
		target := filepath.Join(targetDir, filepath.FromSlash(rel))

		existing, readErr := os.ReadFile(target)
		switch {
		case os.IsNotExist(readErr):
			if err := writeFile(target, data); err != nil {
				return err
			}
			st.Created = append(st.Created, rel)
		case readErr != nil:
			return fmt.Errorf("stat %s: %w", target, readErr)
		case refresh && !isProtected(rel) && string(existing) != string(data):
			backup := filepath.Join(backupRoot, filepath.FromSlash(rel))
			if err := writeFile(backup, existing); err != nil {
				return fmt.Errorf("backup %s: %w", rel, err)
			}
			if err := writeFile(target, data); err != nil {
				return err
			}
			st.Updated = append(st.Updated, rel)
		default:
			st.Skipped++
		}
		return nil
	})
	if err != nil {
		return st, err
	}

	if len(st.Created) > 0 || len(st.Updated) > 0 {
		marker := filepath.Join(targetDir, ".agents", ".seed-version")
		if err := writeFile(marker, []byte(Version()+"\n")); err != nil {
			return st, err
		}
	}
	return st, nil
}

// EnsureAt is the first-run entry point: materialize missing files only,
// and stay quiet when the seed is empty (engine-only build).
func EnsureAt(targetDir string) (Stats, error) {
	if empty() {
		return Stats{}, nil
	}
	return Materialize(targetDir, false)
}

// empty reports whether the binary was built without a synced seed.
func empty() bool {
	entries, err := contentFS.ReadDir(contentRoot)
	if err != nil {
		return true
	}
	for _, e := range entries {
		if !metaFiles[e.Name()] {
			return false
		}
	}
	return true
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ContentHashes returns sha256 (first 12 hex chars) of every embedded seed
// file, keyed by workspace-relative slash path. Used by the diagnostic
// bundle to tell factory files from locally-evolved ones.
func ContentHashes() map[string]string {
	out := map[string]string{}
	_ = fs.WalkDir(contentFS, contentRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(p, contentRoot+"/")
		if metaFiles[rel] || isUserState(rel) {
			return nil
		}
		if b, err := contentFS.ReadFile(p); err == nil {
			out[rel] = hashBytes(b)
		}
		return nil
	})
	return out
}

func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return fmt.Sprintf("%x", sum)[:12]
}
