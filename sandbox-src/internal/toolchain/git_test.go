package toolchain

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGitDownloadURL(t *testing.T) {
	p := Platform{OS: "windows", Arch: "amd64"}
	got := gitDownloadURL(p)
	if got == "" {
		t.Fatal("gitDownloadURL(windows/amd64) is empty")
	}
	if !strings.Contains(got, "git-for-windows") {
		t.Errorf("gitDownloadURL should reference git-for-windows, got %s", got)
	}
	if !strings.Contains(got, "PortableGit") {
		t.Errorf("gitDownloadURL should reference PortableGit, got %s", got)
	}
	if !strings.Contains(got, "64-bit") {
		t.Errorf("gitDownloadURL should reference 64-bit, got %s", got)
	}
}

func TestGitDownloadURL_NonWindows(t *testing.T) {
	tests := []struct {
		os string
	}{
		{"linux"},
		{"darwin"},
	}
	for _, tt := range tests {
		p := Platform{OS: tt.os, Arch: "amd64"}
		got := gitDownloadURL(p)
		if got != "" {
			t.Errorf("gitDownloadURL(%s/amd64) should be empty, got %s", tt.os, got)
		}
	}
}

func TestGitBinPath(t *testing.T) {
	p := Platform{OS: "windows", Arch: "amd64"}
	got := GitBinPath("./sandbox", p)
	want := filepath.Join("sandbox", "git", "cmd", "git.exe")
	if !strings.HasSuffix(got, want) {
		t.Errorf("GitBinPath = %s, want suffix %s", got, want)
	}
}

func TestGitBashPath(t *testing.T) {
	p := Platform{OS: "windows", Arch: "amd64"}
	got := GitBashPath("./sandbox", p)
	want := filepath.Join("sandbox", "git", "bin", "bash.exe")
	if !strings.HasSuffix(got, want) {
		t.Errorf("GitBashPath = %s, want suffix %s", got, want)
	}
}
