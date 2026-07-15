package toolchain

import (
	"testing"
)

func TestUVDownloadURL(t *testing.T) {
	tests := []struct {
		os, arch string
		contains string
	}{
		{"windows", "amd64", "uv-x86_64-pc-windows-msvc.zip"},
		{"linux", "amd64", "uv-x86_64-unknown-linux-gnu.tar.gz"},
		{"linux", "arm64", "uv-aarch64-unknown-linux-gnu.tar.gz"},
		{"darwin", "amd64", "uv-x86_64-apple-darwin.tar.gz"},
		{"darwin", "arm64", "uv-aarch64-apple-darwin.tar.gz"},
	}
	for _, tt := range tests {
		p := Platform{OS: tt.os, Arch: tt.arch}
		got := uvDownloadURL(p)
		if got == "" {
			t.Errorf("uvDownloadURL(%s/%s) is empty", tt.os, tt.arch)
		}
	}
}
