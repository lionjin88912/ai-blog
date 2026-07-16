package migrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildOldInstall creates a fake old install under root with the given layout
// prefix (e.g. ".gemini/skills" or ".agents/skills").
func buildOldInstall(t *testing.T, root, layout string, personas map[string]map[string]string) {
	t.Helper()
	base := filepath.Join(root, filepath.FromSlash(layout), "persona-writer", "personas")
	for slug, files := range personas {
		dir := filepath.Join(base, slug)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for name, content := range files {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestScanBothLayouts(t *testing.T) {
	for _, layout := range []string{".gemini/skills", ".agents/skills"} {
		root := t.TempDir()
		buildOldInstall(t, root, layout, map[string]map[string]string{
			"mrs-lin":   {"persona.md": "# lin", "wp-config.json": "{}"},
			"_template": {"persona.md": "# tmpl"}, // must be ignored
		})
		got := Scan(root)
		if len(got) != 1 || got[0].Slug != "mrs-lin" {
			t.Fatalf("layout %s: expected [mrs-lin], got %+v", layout, got)
		}
	}
}

func TestMigrateCleansWPConfigAndInfersAuth(t *testing.T) {
	old := t.TempDir()
	buildOldInstall(t, old, ".gemini/skills", map[string]map[string]string{
		"yoyo": {
			"persona.md": "# yoyo",
			// legacy config: no auth_method, plus a junk key that must be dropped
			"wp-config.json": `{"WP_URL":"https://y.com","WP_USER":"a@b.com","WP_APP_PWD":"x x x","legacy_junk":"drop-me","seo_plugin":"rankmath"}`,
			"published.json": `[{"id":1}]`,
			"draft.json":     `{"stage":"h1_done"}`,
		},
	})
	target := filepath.Join(t.TempDir(), "personas")

	rep, err := Migrate(old, target, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Imported != 1 {
		t.Fatalf("expected 1 imported, got %+v", rep)
	}

	// wp-config cleaned: auth_method inferred, junk dropped, secret kept.
	b, err := os.ReadFile(filepath.Join(target, "yoyo", "wp-config.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(b, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg["auth_method"] != "application_password" {
		t.Errorf("auth_method not inferred: %v", cfg["auth_method"])
	}
	if _, ok := cfg["legacy_junk"]; ok {
		t.Error("junk key was not dropped")
	}
	if cfg["WP_APP_PWD"] != "x x x" {
		t.Error("credential not preserved")
	}

	// published.json migrated; draft.json intentionally NOT.
	if _, err := os.Stat(filepath.Join(target, "yoyo", "published.json")); err != nil {
		t.Error("published.json should be migrated")
	}
	if _, err := os.Stat(filepath.Join(target, "yoyo", "draft.json")); !os.IsNotExist(err) {
		t.Error("draft.json must NOT be migrated (transient state)")
	}
}

func TestMigrateNeverOverwritesExistingFile(t *testing.T) {
	old := t.TempDir()
	buildOldInstall(t, old, ".agents/skills", map[string]map[string]string{
		"lin": {"persona.md": "# OLD", "wp-config.json": "{}"},
	})
	target := filepath.Join(t.TempDir(), "personas")
	// persona.md already present (e.g. seeded), wp-config absent.
	existing := filepath.Join(target, "lin")
	if err := os.MkdirAll(existing, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(existing, "persona.md"), []byte("# KEEP ME"), 0o644); err != nil {
		t.Fatal(err)
	}

	rep, err := Migrate(old, target, false)
	if err != nil {
		t.Fatal(err)
	}
	// Merge: persona.md kept, wp-config filled in.
	if rep.Merged != 1 || rep.Imported != 0 || rep.Skipped != 0 {
		t.Fatalf("expected merge, got %+v", rep)
	}
	if b, _ := os.ReadFile(filepath.Join(existing, "persona.md")); string(b) != "# KEEP ME" {
		t.Error("existing persona.md was overwritten")
	}
	if _, err := os.Stat(filepath.Join(existing, "wp-config.json")); err != nil {
		t.Error("missing wp-config.json should have been merged in")
	}
}

func TestMigrateMergesSeededPersonaWpConfig(t *testing.T) {
	// Reproduces the mrs-lin case: seed pre-creates persona.md (no wp-config);
	// migrate must still import the old wp-config.
	old := t.TempDir()
	buildOldInstall(t, old, ".gemini/skills", map[string]map[string]string{
		"mrs-lin": {"persona.md": "# old lin", "wp-config.json": `{"WP_URL":"https://lin.com","WP_APP_PWD":"p"}`},
	})
	target := filepath.Join(t.TempDir(), "personas")
	seeded := filepath.Join(target, "mrs-lin")
	os.MkdirAll(seeded, 0o755)
	os.WriteFile(filepath.Join(seeded, "persona.md"), []byte("# seeded lin"), 0o644)

	rep, err := Migrate(old, target, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Merged != 1 {
		t.Fatalf("expected merged=1, got %+v", rep)
	}
	if _, err := os.Stat(filepath.Join(seeded, "wp-config.json")); err != nil {
		t.Error("old mrs-lin wp-config must be merged into the seeded persona")
	}
	if b, _ := os.ReadFile(filepath.Join(seeded, "persona.md")); string(b) != "# seeded lin" {
		t.Error("seeded persona.md must be preserved")
	}
}

func TestMigrateSkipsWhenFullyPresent(t *testing.T) {
	old := t.TempDir()
	buildOldInstall(t, old, ".agents/skills", map[string]map[string]string{
		"lin": {"persona.md": "# a", "wp-config.json": "{}"},
	})
	target := filepath.Join(t.TempDir(), "personas")
	dst := filepath.Join(target, "lin")
	os.MkdirAll(dst, 0o755)
	os.WriteFile(filepath.Join(dst, "persona.md"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dst, "wp-config.json"), []byte("{}"), 0o644)

	rep, err := Migrate(old, target, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Skipped != 1 || rep.Merged != 0 || rep.Imported != 0 {
		t.Fatalf("expected skip when nothing missing, got %+v", rep)
	}
}

func TestScanDedupsSlugPrefersAgents(t *testing.T) {
	root := t.TempDir()
	// same slug in both layouts
	buildOldInstall(t, root, ".gemini/skills", map[string]map[string]string{
		"lin": {"persona.md": "# gemini"},
	})
	buildOldInstall(t, root, ".agents/skills", map[string]map[string]string{
		"lin": {"persona.md": "# agents"},
	})
	got := Scan(root)
	if len(got) != 1 {
		t.Fatalf("expected 1 deduped persona, got %d: %+v", len(got), got)
	}
	if !strings.Contains(filepath.ToSlash(got[0].Dir), "/.agents/") {
		t.Errorf("expected .agents layout to win, got %s", got[0].Dir)
	}
}

func TestMigrateSkipsBrokenConfigButKeepsPersona(t *testing.T) {
	old := t.TempDir()
	buildOldInstall(t, old, ".gemini/skills", map[string]map[string]string{
		"broken": {"persona.md": "# b", "wp-config.json": "{not json"},
	})
	target := filepath.Join(t.TempDir(), "personas")

	rep, err := Migrate(old, target, false)
	if err != nil {
		t.Fatal(err)
	}
	// persona still imported; config skipped (not written).
	if rep.Imported != 1 {
		t.Fatalf("expected persona imported, got %+v", rep)
	}
	if _, err := os.Stat(filepath.Join(target, "broken", "persona.md")); err != nil {
		t.Error("persona.md should still be copied")
	}
	if _, err := os.Stat(filepath.Join(target, "broken", "wp-config.json")); !os.IsNotExist(err) {
		t.Error("broken wp-config.json must not be written")
	}
}

func TestDryRunWritesNothing(t *testing.T) {
	old := t.TempDir()
	buildOldInstall(t, old, ".agents/skills", map[string]map[string]string{
		"lin": {"persona.md": "# lin", "wp-config.json": "{}"},
	})
	target := filepath.Join(t.TempDir(), "personas")

	rep, err := Migrate(old, target, true)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Imported != 1 {
		t.Fatalf("dry-run should still report 1 would-import, got %+v", rep)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Error("dry-run must not create the target dir")
	}
}

func TestFindCandidates(t *testing.T) {
	search := t.TempDir()
	// one folder with a persona, one without
	buildOldInstall(t, filepath.Join(search, "old-pkg"), ".gemini/skills", map[string]map[string]string{
		"lin": {"persona.md": "# lin"},
	})
	if err := os.MkdirAll(filepath.Join(search, "unrelated"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := FindCandidates([]string{search})
	if len(got) != 1 || filepath.Base(got[0].Path) != "old-pkg" || got[0].Count != 1 {
		t.Fatalf("expected [old-pkg count=1], got %+v", got)
	}
}
