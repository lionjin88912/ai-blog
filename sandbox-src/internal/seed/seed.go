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
	"bytes"
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

// IsPersonaData reports whether rel (slash-separated, relative to workspace
// root) is a user's persona (not the _template). Such files are never
// overwritten by refresh and, once a workspace is seeded, never resurrected.
func IsPersonaData(rel string) bool {
	const personas = ".agents/skills/persona-writer/personas/"
	if !strings.HasPrefix(rel, personas) {
		return false
	}
	return !strings.HasPrefix(rel, personas+"_template/")
}

// IsUserState reports runtime files that are per-machine state, not factory
// content: they are never embedded in the seed (seedsync skips them) and never
// written by Materialize. Single source of truth shared with the seedsync tool
// and the diag package.
func IsUserState(rel string) bool {
	base := filepath.Base(rel)
	if base == "wp-config.json" || base == "published.json" || base == "draft.json" {
		return true
	}
	return strings.Contains(rel, "/articles/")
}

func isProtected(rel string) bool { return IsPersonaData(rel) }
func isUserState(rel string) bool { return IsUserState(rel) }

// Materialize lays the embedded seed into targetDir.
//
// refresh=false (first run): only create missing files.
// refresh=true (refresh-skills): also overwrite non-protected files whose
// content differs, backing up the old file first.
func Materialize(targetDir string, refresh bool) (Stats, error) {
	var st Stats
	stamp := Version() + "-" + time.Now().Format("20060102-150405")
	backupRoot := filepath.Join(targetDir, ".agents", "_backup", stamp)

	// Once a workspace has been seeded, a missing protected (persona) file is
	// a deliberate user deletion — never resurrect it. Missing SOP files are
	// still restored so a broken workspace can self-heal.
	alreadySeeded := LocalVersion(targetDir) != ""

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
		target := filepath.Join(targetDir, filepath.FromSlash(rel))

		info, statErr := os.Stat(target)
		switch {
		case os.IsNotExist(statErr):
			if alreadySeeded && isProtected(rel) {
				st.Skipped++ // user deleted this persona file; respect it
				return nil
			}
			data, err := contentFS.ReadFile(p)
			if err != nil {
				return fmt.Errorf("read embedded %s: %w", rel, err)
			}
			if err := writeFile(target, data); err != nil {
				return err
			}
			st.Created = append(st.Created, rel)
		case statErr != nil:
			return fmt.Errorf("stat %s: %w", target, statErr)
		case info.IsDir():
			return fmt.Errorf("%s exists as a directory, expected a file", target)
		case refresh && !isProtected(rel):
			data, err := contentFS.ReadFile(p)
			if err != nil {
				return fmt.Errorf("read embedded %s: %w", rel, err)
			}
			existing, err := os.ReadFile(target)
			if err != nil {
				return fmt.Errorf("read %s: %w", target, err)
			}
			if bytes.Equal(existing, data) {
				st.Skipped++
				return nil
			}
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

	// Always record the seed version we just applied, even when nothing
	// changed, so LocalVersion reflects reality after a no-op refresh.
	marker := filepath.Join(targetDir, ".agents", ".seed-version")
	if err := writeFile(marker, []byte(Version()+"\n")); err != nil {
		return st, err
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

// EnsureReport runs EnsureAt and reports the outcome through printf (which may
// be log.Printf or fmt.Printf). The single seeding entry point shared by every
// command so behavior can't drift between them.
func EnsureReport(targetDir string, printf func(string, ...any)) {
	st, err := EnsureAt(targetDir)
	if err != nil {
		printf("⚠️  Seed materialize failed: %v", err)
		return
	}
	if len(st.Created) > 0 {
		printf("🌱 Factory content seeded: %d files (seed %s)", len(st.Created), Version())
	}
}

// empty reports whether the binary was built without a synced seed. A synced
// seed always carries SEED_VERSION, so Version()=="dev" is exactly that case.
func empty() bool {
	return Version() == "dev"
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
			out[rel] = HashBytes(b)
		}
		return nil
	})
	return out
}

// HashBytes returns the first 12 hex chars of sha256(b). Exported so the diag
// package computes workspace-file hashes identically to the seed baseline —
// the two MUST agree for seed-drift detection to work.
func HashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return fmt.Sprintf("%x", sum)[:12]
}
