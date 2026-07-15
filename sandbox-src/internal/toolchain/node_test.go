package toolchain

import (
	"testing"
)

func TestNodeDownloadURL(t *testing.T) {
	tests := []struct {
		os, arch string
		wantURL  string
	}{
		{"windows", "amd64", "https://nodejs.org/dist/v22.22.2/node-v22.22.2-win-x64.zip"},
		{"linux", "amd64", "https://nodejs.org/dist/v22.22.2/node-v22.22.2-linux-x64.tar.gz"},
		{"linux", "arm64", "https://nodejs.org/dist/v22.22.2/node-v22.22.2-linux-arm64.tar.gz"},
		{"darwin", "amd64", "https://nodejs.org/dist/v22.22.2/node-v22.22.2-darwin-x64.tar.gz"},
		{"darwin", "arm64", "https://nodejs.org/dist/v22.22.2/node-v22.22.2-darwin-arm64.tar.gz"},
	}
	for _, tt := range tests {
		p := Platform{OS: tt.os, Arch: tt.arch}
		got := nodeDownloadURL(p)
		if got != tt.wantURL {
			t.Errorf("nodeDownloadURL(%s/%s) = %s, want %s", tt.os, tt.arch, got, tt.wantURL)
		}
	}
}

func TestNodeBinPath(t *testing.T) {
	tests := []struct {
		os   string
		want string // should end with
	}{
		{"windows", "node.exe"},
		{"linux", "node"},
		{"darwin", "node"},
	}
	for _, tt := range tests {
		p := Platform{OS: tt.os, Arch: "amd64"}
		got := NodeBinPath("./sandbox", p)
		if got == "" {
			t.Errorf("NodeBinPath empty for %s", tt.os)
		}
	}
}
