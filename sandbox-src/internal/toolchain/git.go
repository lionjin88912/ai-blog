package toolchain

import (
	"fmt"
	"os"
	"path/filepath"
)

const gitVersion = "2.49.0"

func gitDownloadURL(p Platform) string {
	if p.OS != "windows" {
		return ""
	}
	// Official PortableGit self-extracting archive from git-for-windows
	return fmt.Sprintf(
		"https://github.com/git-for-windows/git/releases/download/v%s.windows.1/PortableGit-%s-64-bit.7z.exe",
		gitVersion, gitVersion,
	)
}

// GitBinPath returns the path to the git executable.
func GitBinPath(sandboxDir string, p Platform) string {
	return filepath.Join(sandboxDir, "git", "cmd", "git.exe")
}

// GitBashPath returns the path to the MINGW64 bash executable.
func GitBashPath(sandboxDir string, p Platform) string {
	return filepath.Join(sandboxDir, "git", "bin", "bash.exe")
}

// DownloadGit downloads and extracts Portable Git to sandboxDir/git/.
// Only runs on Windows; other platforms skip (git is assumed system-installed).
func DownloadGit(sandboxDir string, p Platform) error {
	if p.OS != "windows" {
		fmt.Println("  Skipping Portable Git (not Windows).")
		return nil
	}

	if _, err := os.Stat(GitBinPath(sandboxDir, p)); err == nil {
		fmt.Println("  Portable Git already installed, skipping.")
		return nil
	}

	url := gitDownloadURL(p)
	archiveName := filepath.Base(url)
	archivePath := filepath.Join(sandboxDir, archiveName)

	fmt.Printf("  Downloading Portable Git %s...\n", gitVersion)
	if err := DownloadFile(url, archivePath); err != nil {
		return fmt.Errorf("download git: %w", err)
	}
	defer os.Remove(archivePath)

	gitDir := filepath.Join(sandboxDir, "git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		return err
	}

	// PortableGit .7z.exe supports silent extraction with -y -o<dir>
	fmt.Println("  Extracting Portable Git (this may take a moment)...")
	cmd := execCommand(archivePath, "-y", "-o"+gitDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("extract portable git: %w", err)
	}

	fmt.Println("  ✅ Portable Git installed.")
	return nil
}
