// Command seedsync populates internal/seed/content from the repo root before
// a build. It is cross-platform (replaces the old rsync/cp/printf recipe that
// broke Windows `make build`) and is the single source of truth for what
// counts as per-machine user state.
//
// Persona files are filtered through `git ls-files`: only tracked persona
// content ships in the seed, so a developer's untracked local test personas
// (and any in-flight drafts) never get embedded into distributed binaries.
//
// Usage: go run ./tools/seedsync <repo-root> <seed-content-dir>
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ai-sandbox/cli/internal/seed"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: seedsync <repo-root> <seed-content-dir>")
		os.Exit(2)
	}
	repoRoot, seedDir := os.Args[1], os.Args[2]
	if err := run(repoRoot, seedDir); err != nil {
		fmt.Fprintf(os.Stderr, "seedsync: %v\n", err)
		os.Exit(1)
	}
}

func run(repoRoot, seedDir string) error {
	// Clean previous seed payload (keep README.md).
	agentsOut := filepath.Join(seedDir, ".agents")
	if err := os.RemoveAll(agentsOut); err != nil {
		return err
	}
	for _, f := range []string{"GEMINI.md", "HOWTO.md", "BOOTSTRAP.md", "SEED_VERSION"} {
		_ = os.Remove(filepath.Join(seedDir, f))
	}

	tracked := gitTracked(repoRoot)

	skillsSrc := filepath.Join(repoRoot, ".agents", "skills")
	skillsDst := filepath.Join(seedDir, ".agents", "skills")
	err := filepath.Walk(skillsSrc, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(skillsSrc, p)
		relSlash := ".agents/skills/" + filepath.ToSlash(rel)

		if seed.IsUserState(relSlash) || strings.Contains(relSlash, "__pycache__") || filepath.Base(p) == ".DS_Store" {
			return nil
		}
		// Persona content: ship only tracked files, so untracked local test
		// personas and drafts on the build machine never get embedded.
		if seed.IsPersonaData(relSlash) && tracked != nil && !tracked[relSlash] {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		dst := filepath.Join(skillsDst, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0o644)
	})
	if err != nil {
		return err
	}

	for _, f := range []string{"GEMINI.md", "HOWTO.md", "BOOTSTRAP.md"} {
		data, err := os.ReadFile(filepath.Join(repoRoot, f))
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		if err := os.WriteFile(filepath.Join(seedDir, f), data, 0o644); err != nil {
			return err
		}
	}

	ver := fmt.Sprintf("%s-%s", time.Now().Format("20060102-150405"), gitShort(repoRoot))
	if err := os.WriteFile(filepath.Join(seedDir, "SEED_VERSION"), []byte(ver+"\n"), 0o644); err != nil {
		return err
	}
	fmt.Printf("Seed synced: %s\n", ver)
	return nil
}

// gitTracked returns the set of repo-relative (slash) paths git tracks, or nil
// if git is unavailable (in which case the caller ships everything).
func gitTracked(repoRoot string) map[string]bool {
	cmd := exec.Command("git", "-C", repoRoot, "ls-files")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	set := map[string]bool{}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		set[strings.TrimSpace(sc.Text())] = true
	}
	return set
}

func gitShort(repoRoot string) string {
	out, err := exec.Command("git", "-C", repoRoot, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "nogit"
	}
	return strings.TrimSpace(string(out))
}
