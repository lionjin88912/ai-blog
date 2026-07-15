package toolchain

import "runtime"

// Platform describes the current OS and architecture.
type Platform struct {
	OS   string // "windows", "linux", "darwin"
	Arch string // "amd64", "arm64"
}

// DetectPlatform returns the current platform.
func DetectPlatform() Platform {
	arch := runtime.GOARCH
	if arch == "386" {
		arch = "amd64" // fallback
	}
	return Platform{
		OS:   runtime.GOOS,
		Arch: arch,
	}
}

// ArchiveExt returns the typical archive extension for this platform.
func (p Platform) ArchiveExt() string {
	if p.OS == "windows" {
		return ".zip"
	}
	return ".tar.gz"
}

// ExeExt returns the executable extension for this platform.
func (p Platform) ExeExt() string {
	if p.OS == "windows" {
		return ".exe"
	}
	return ""
}
