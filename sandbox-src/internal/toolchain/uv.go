package toolchain

import (
	"fmt"
	"os"
	"path/filepath"
)

const uvVersion = "0.7.2"

func uvDownloadURL(p Platform) string {
	var target string
	switch {
	case p.OS == "windows" && p.Arch == "amd64":
		target = "uv-x86_64-pc-windows-msvc.zip"
	case p.OS == "linux" && p.Arch == "amd64":
		target = "uv-x86_64-unknown-linux-gnu.tar.gz"
	case p.OS == "linux" && p.Arch == "arm64":
		target = "uv-aarch64-unknown-linux-gnu.tar.gz"
	case p.OS == "darwin" && p.Arch == "amd64":
		target = "uv-x86_64-apple-darwin.tar.gz"
	case p.OS == "darwin" && p.Arch == "arm64":
		target = "uv-aarch64-apple-darwin.tar.gz"
	}
	return fmt.Sprintf("https://github.com/astral-sh/uv/releases/download/%s/%s", uvVersion, target)
}

func uvExtractDirName(p Platform) string {
	switch {
	case p.OS == "windows" && p.Arch == "amd64":
		return "uv-x86_64-pc-windows-msvc"
	case p.OS == "linux" && p.Arch == "amd64":
		return "uv-x86_64-unknown-linux-gnu"
	case p.OS == "linux" && p.Arch == "arm64":
		return "uv-aarch64-unknown-linux-gnu"
	case p.OS == "darwin" && p.Arch == "amd64":
		return "uv-x86_64-apple-darwin"
	case p.OS == "darwin" && p.Arch == "arm64":
		return "uv-aarch64-apple-darwin"
	}
	return ""
}

// UVBinPath returns the path to the uv executable.
func UVBinPath(sandboxDir string, p Platform) string {
	return filepath.Join(sandboxDir, "uv", "uv"+p.ExeExt())
}

// PythonBinPath returns the path to the python3.12 executable.
// uv installs Python into a versioned subdirectory (e.g. cpython-3.12.10-windows-x86_64-none/).
func PythonBinPath(sandboxDir string, p Platform) string {
	pythonDir := filepath.Join(sandboxDir, "python")

	// Look for cpython-* subdirectory created by uv
	entries, err := os.ReadDir(pythonDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() && len(e.Name()) >= 7 && e.Name()[:7] == "cpython" {
				if p.OS == "windows" {
					return filepath.Join(pythonDir, e.Name(), "python.exe")
				}
				return filepath.Join(pythonDir, e.Name(), "bin", "python3.12")
			}
		}
	}

	// Fallback: direct layout
	if p.OS == "windows" {
		return filepath.Join(pythonDir, "python.exe")
	}
	return filepath.Join(pythonDir, "bin", "python3.12")
}

// DownloadUV downloads and extracts uv to sandboxDir/uv/.
func DownloadUV(sandboxDir string, p Platform) error {
	uvDir := filepath.Join(sandboxDir, "uv")
	if _, err := os.Stat(UVBinPath(sandboxDir, p)); err == nil {
		fmt.Println("  uv already installed, skipping.")
		return nil
	}

	url := uvDownloadURL(p)
	archiveName := filepath.Base(url)
	archivePath := filepath.Join(sandboxDir, archiveName)

	fmt.Printf("  Downloading uv %s...\n", uvVersion)
	if err := DownloadFile(url, archivePath); err != nil {
		return fmt.Errorf("download uv: %w", err)
	}
	defer os.Remove(archivePath)

	// Extract directly to target directory
	os.RemoveAll(uvDir)
	if err := os.MkdirAll(uvDir, 0755); err != nil {
		return err
	}

	fmt.Println("  Extracting uv...")
	if err := ExtractArchive(archivePath, uvDir); err != nil {
		return fmt.Errorf("extract uv: %w", err)
	}

	// If archive contained a subdirectory, move its contents up
	subDir := filepath.Join(uvDir, uvExtractDirName(p))
	if info, err := os.Stat(subDir); err == nil && info.IsDir() {
		entries, _ := os.ReadDir(subDir)
		for _, e := range entries {
			os.Rename(filepath.Join(subDir, e.Name()), filepath.Join(uvDir, e.Name()))
		}
		os.RemoveAll(subDir)
	}

	fmt.Println("  ✅ uv installed.")
	return nil
}

// InstallPython installs Python 3.12 using uv to sandboxDir/python/.
func InstallPython(sandboxDir string, p Platform) error {
	if _, err := os.Stat(PythonBinPath(sandboxDir, p)); err == nil {
		fmt.Println("  Python 3.12 already installed, skipping.")
		return nil
	}

	uvPath := UVBinPath(sandboxDir, p)
	if _, err := os.Stat(uvPath); os.IsNotExist(err) {
		return fmt.Errorf("uv not found at %s — run uv setup first", uvPath)
	}

	pythonDir := filepath.Join(sandboxDir, "python")
	if err := os.MkdirAll(pythonDir, 0755); err != nil {
		return err
	}

	fmt.Println("  Installing Python 3.12 via uv...")
	cmd := execCommand(uvPath, "python", "install", "3.12",
		"--python-preference", "only-managed",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"UV_PYTHON_INSTALL_DIR="+pythonDir,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("uv python install: %w", err)
	}

	fmt.Println("  ✅ Python 3.12 installed.")
	return nil
}
