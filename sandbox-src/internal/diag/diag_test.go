package diag

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaskedJSONNeverLeaksSecrets(t *testing.T) {
	dir := t.TempDir()
	cfg := map[string]any{
		"site_url":             "https://example.wordpress.com",
		"username":             "lin",
		"app_password":         "abcd efgh ijkl mnop",
		"password":             "hunter2",
		"api_key":              "AIzaSyFAKE",
		"oauth_token":          "ya29.fake",
		"client_secret":        "sekrit",
		"nested":               map[string]any{"wp_password": "deep-secret"},
		"empty_password":       "",
		"harmless_description": "just text",
	}
	b, _ := json.Marshal(cfg)
	path := filepath.Join(dir, "wp-config.json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}

	out, _ := json.Marshal(maskedJSON(path))
	s := string(out)
	for _, leak := range []string{"hunter2", "abcd efgh", "AIzaSyFAKE", "ya29.fake", "sekrit", "deep-secret"} {
		if strings.Contains(s, leak) {
			t.Errorf("secret %q leaked into diagnostic output", leak)
		}
	}
	for _, keep := range []string{"example.wordpress.com", "lin", "just text"} {
		if !strings.Contains(s, keep) {
			t.Errorf("non-secret %q missing from output", keep)
		}
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
