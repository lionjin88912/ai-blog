package seed

import "testing"

func TestIsProtected(t *testing.T) {
	cases := []struct {
		rel  string
		want bool
	}{
		{".agents/skills/persona-writer/personas/mrs-lin-slow-travel/persona.md", true},
		{".agents/skills/persona-writer/personas/yoyo/persona.md", true},
		{".agents/skills/persona-writer/personas/_template/persona.md", false},
		{".agents/skills/persona-writer/personas/_template/wp-config.example.json", false},
		{".agents/skills/persona-writer/SKILL.md", false},
		{".agents/skills/persona-writer/scripts/wp_poster.py", false},
		{"GEMINI.md", false},
	}
	for _, c := range cases {
		if got := isProtected(c.rel); got != c.want {
			t.Errorf("isProtected(%q) = %v, want %v", c.rel, got, c.want)
		}
	}
}

func TestIsUserState(t *testing.T) {
	cases := []struct {
		rel  string
		want bool
	}{
		{".agents/skills/persona-writer/personas/yoyo/wp-config.json", true},
		{".agents/skills/persona-writer/personas/yoyo/published.json", true},
		{".agents/skills/persona-writer/personas/yoyo/articles/x.html", true},
		{".agents/skills/persona-writer/personas/_template/wp-config.example.json", false},
		{".agents/skills/persona-writer/SKILL.md", false},
	}
	for _, c := range cases {
		if got := isUserState(c.rel); got != c.want {
			t.Errorf("isUserState(%q) = %v, want %v", c.rel, got, c.want)
		}
	}
}

func TestVersionDevWhenUnsynced(t *testing.T) {
	// In a fresh checkout the seed only has README.md, so Version() falls
	// back to "dev" and EnsureAt must be a no-op.
	if empty() {
		if v := Version(); v != "dev" {
			t.Errorf("Version() = %q, want dev", v)
		}
		st, err := EnsureAt(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		if len(st.Created) != 0 || len(st.Updated) != 0 {
			t.Errorf("EnsureAt on empty seed wrote files: %+v", st)
		}
	}
}

func TestUserStateAndPersonaPolicy(t *testing.T) {
	if !IsPersonaData(".agents/skills/persona-writer/personas/mrs-lin/persona.md") {
		t.Error("persona file should be protected")
	}
	if IsPersonaData(".agents/skills/persona-writer/personas/_template/persona.md") {
		t.Error("_template must not be protected")
	}
	if IsPersonaData(".agents/skills/persona-writer/SKILL.md") {
		t.Error("SKILL.md must not be protected")
	}
	if !IsUserState(".agents/skills/persona-writer/personas/x/draft.json") {
		t.Error("draft.json must be user state")
	}
}
