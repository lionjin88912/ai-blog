package diag

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaskedConfigNeverLeaksSecrets(t *testing.T) {
	dir := t.TempDir()
	// ACTUAL production key names from wp-config.example.json — the exact
	// keys the old substring regex missed (WP_APP_PWD, WP_USER).
	cfg := map[string]any{
		"auth_method":      "application_password",
		"WP_URL":           "https://example-blog.com",
		"WP_USER":          "user@example.com",
		"WP_APP_PWD":       "FAKE PWD1 PWD2 PWD3 PWD4 PWD5",
		"WP_CLIENT_ID":     "140486",
		"WP_CLIENT_SECRET": "cs-sekrit",
		"WP_ACCESS_TOKEN":  "ya29.fake",
		"WP_BLOG_ID":       "12345",
		"seo_plugin":       "rankmath",
		"nested":           map[string]any{"WP_APP_PWD": "deep-secret"},
		"arr":              []any{map[string]any{"WP_APP_PWD": "arr-secret"}},
		"empty_pwd":        "",
	}
	b, _ := json.Marshal(cfg)
	path := filepath.Join(dir, "wp-config.json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}

	out, _ := json.Marshal(maskedConfig(path))
	s := string(out)
	for _, leak := range []string{"FAKE PWD1", "user-account", "cs-sekrit", "ya29.fake", "deep-secret", "arr-secret"} {
		if strings.Contains(s, leak) {
			t.Errorf("secret %q leaked into diagnostic output", leak)
		}
	}
	for _, keep := range []string{"example-blog.com", "application_password", "rankmath", "12345"} {
		if !strings.Contains(s, keep) {
			t.Errorf("allowlisted field %q missing from output", keep)
		}
	}
}

func TestScrubTextRedactsLogSecrets(t *testing.T) {
	in := []byte(strings.Join([]string{
		"I0708 normal log line",
		"Authorization: Bearer ya29.aVeryLongOpaqueTokenValueExceedingForty12345",
		`  "app_password": "FAKE PWD1 PWD2"`,
		"ordinary text no secret",
	}, "\n"))
	out := string(scrubText(in))
	for _, leak := range []string{"aVeryLongOpaqueToken", "FAKE PWD1"} {
		if strings.Contains(out, leak) {
			t.Errorf("scrubText leaked %q", leak)
		}
	}
	if !strings.Contains(out, "normal log line") {
		t.Error("scrubText dropped a benign line")
	}
}

func TestBuildZipSmoke(t *testing.T) {
	ws := t.TempDir()
	skills := filepath.Join(ws, ".agents", "skills", "persona-writer")
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, "SKILL.md"), []byte("# x"), 0o644); err != nil {
		t.Fatal(err)
	}
	data, name, err := BuildZip(ws, "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 || !strings.HasPrefix(name, "ai-sandbox-diag-") {
		t.Errorf("unexpected bundle: %d bytes, name %q", len(data), name)
	}
}
