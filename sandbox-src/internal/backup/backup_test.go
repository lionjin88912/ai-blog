package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ai-sandbox/cli/internal/migrate"
)

func writePersona(t *testing.T, dataDir, slug string, files map[string]string) {
	t.Helper()
	dir := filepath.Join(migrate.TargetPersonasDir(dataDir), slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestExportThenImportRoundTrip(t *testing.T) {
	src := t.TempDir()
	writePersona(t, src, "mrs-lin", map[string]string{
		"persona.md":     "# lin",
		"wp-config.json": `{"WP_URL":"https://lin.com","WP_APP_PWD":"secret","junk":"x"}`,
		"published.json": `[{"id":1}]`,
		"draft.json":     `{"stage":"h1"}`, // must NOT be exported
	})
	writePersona(t, src, "_template", map[string]string{"persona.md": "# tmpl"}) // must be skipped

	data, name, err := Export(src, time.Unix(1700000000, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 || name == "" {
		t.Fatal("empty export")
	}

	// Import into a fresh (blank) install.
	dst := t.TempDir()
	rep, err := Import(data, dst)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Imported != 1 {
		t.Fatalf("expected 1 imported, got %+v", rep)
	}

	base := filepath.Join(migrate.TargetPersonasDir(dst), "mrs-lin")
	if _, err := os.Stat(filepath.Join(base, "persona.md")); err != nil {
		t.Error("persona.md not imported")
	}
	if _, err := os.Stat(filepath.Join(base, "published.json")); err != nil {
		t.Error("published.json not imported")
	}
	// draft.json excluded from backup AND not migrated.
	if _, err := os.Stat(filepath.Join(base, "draft.json")); !os.IsNotExist(err) {
		t.Error("draft.json should not be present")
	}
	// _template must not have been exported/imported.
	if _, err := os.Stat(filepath.Join(migrate.TargetPersonasDir(dst), "_template")); !os.IsNotExist(err) {
		t.Error("_template should not be in the backup")
	}
	// wp-config cleaned (junk dropped, auth_method inferred, secret kept).
	b, _ := os.ReadFile(filepath.Join(base, "wp-config.json"))
	s := string(b)
	if !contains(s, "application_password") || !contains(s, "secret") || contains(s, "junk") {
		t.Errorf("wp-config not cleaned as expected: %s", s)
	}
}

func TestImportDoesNotOverwriteExisting(t *testing.T) {
	src := t.TempDir()
	writePersona(t, src, "lin", map[string]string{"persona.md": "# from backup", "wp-config.json": "{}"})
	data, _, err := Export(src, time.Unix(1700000000, 0))
	if err != nil {
		t.Fatal(err)
	}

	dst := t.TempDir()
	writePersona(t, dst, "lin", map[string]string{"persona.md": "# KEEP"}) // persona.md exists, wp-config missing

	rep, err := Import(data, dst)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Merged != 1 {
		t.Fatalf("expected merge, got %+v", rep)
	}
	if b, _ := os.ReadFile(filepath.Join(migrate.TargetPersonasDir(dst), "lin", "persona.md")); string(b) != "# KEEP" {
		t.Error("existing persona.md was overwritten")
	}
	if _, err := os.Stat(filepath.Join(migrate.TargetPersonasDir(dst), "lin", "wp-config.json")); err != nil {
		t.Error("missing wp-config should have been merged from backup")
	}
}

func TestImportFilesFromFolder(t *testing.T) {
	// Simulate a browser folder pick: files with arbitrary prefix paths.
	files := []UploadFile{
		{RelPath: "oldpkg/.gemini/skills/persona-writer/personas/lin/persona.md", Data: []byte("# lin")},
		{RelPath: "oldpkg/.gemini/skills/persona-writer/personas/lin/wp-config.json", Data: []byte(`{"WP_URL":"https://l.com","WP_APP_PWD":"p","junk":"x"}`)},
		{RelPath: "oldpkg/.gemini/skills/persona-writer/personas/lin/draft.json", Data: []byte(`{"stage":"h1"}`)},        // ignored
		{RelPath: "oldpkg/.gemini/skills/persona-writer/personas/_template/persona.md", Data: []byte("# t")},           // ignored
		{RelPath: "oldpkg/random/other.txt", Data: []byte("nope")},                                                     // ignored
		{RelPath: "oldpkg/.gemini/skills/persona-writer/personas/lin/sub/persona.md", Data: []byte("# nested")},        // ignored (nested)
	}
	dst := t.TempDir()
	rep, err := ImportFiles(files, dst)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Imported != 1 {
		t.Fatalf("expected 1 imported, got %+v", rep)
	}
	base := filepath.Join(migrate.TargetPersonasDir(dst), "lin")
	if _, err := os.Stat(filepath.Join(base, "wp-config.json")); err != nil {
		t.Error("wp-config not imported")
	}
	if _, err := os.Stat(filepath.Join(base, "draft.json")); !os.IsNotExist(err) {
		t.Error("draft.json must not be imported")
	}
	if _, err := os.Stat(filepath.Join(migrate.TargetPersonasDir(dst), "_template")); !os.IsNotExist(err) {
		t.Error("_template must not be imported")
	}
}

func TestImportFilesEmptyErrors(t *testing.T) {
	files := []UploadFile{{RelPath: "some/folder/readme.txt", Data: []byte("x")}}
	if _, err := ImportFiles(files, t.TempDir()); err == nil {
		t.Error("expected error when no persona files present")
	}
}

func TestExportEmptyErrors(t *testing.T) {
	if _, _, err := Export(t.TempDir(), time.Unix(1700000000, 0)); err == nil {
		t.Error("expected error exporting a workspace with no personas")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
