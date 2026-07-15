package toolchain

import (
	"runtime"
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	p := DetectPlatform()

	// Must return valid OS
	switch p.OS {
	case "windows", "linux", "darwin":
	default:
		t.Errorf("unexpected OS: %s", p.OS)
	}

	// Must return valid Arch
	switch p.Arch {
	case "amd64", "arm64":
	default:
		t.Errorf("unexpected Arch: %s", p.Arch)
	}

	// Must match runtime
	if p.OS != runtime.GOOS {
		t.Errorf("OS mismatch: got %s, want %s", p.OS, runtime.GOOS)
	}
}

func TestPlatformArchiveExt(t *testing.T) {
	tests := []struct {
		os   string
		want string
	}{
		{"windows", ".zip"},
		{"linux", ".tar.gz"},
		{"darwin", ".tar.gz"},
	}
	for _, tt := range tests {
		p := Platform{OS: tt.os, Arch: "amd64"}
		if got := p.ArchiveExt(); got != tt.want {
			t.Errorf("ArchiveExt(%s) = %s, want %s", tt.os, got, tt.want)
		}
	}
}

func TestPlatformExeExt(t *testing.T) {
	tests := []struct {
		os   string
		want string
	}{
		{"windows", ".exe"},
		{"linux", ""},
		{"darwin", ""},
	}
	for _, tt := range tests {
		p := Platform{OS: tt.os, Arch: "amd64"}
		if got := p.ExeExt(); got != tt.want {
			t.Errorf("ExeExt(%s) = %q, want %q", tt.os, got, tt.want)
		}
	}
}
