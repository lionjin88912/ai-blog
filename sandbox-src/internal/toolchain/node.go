package toolchain

import (
	"fmt"
	"os"
	"path/filepath"
)

const nodeVersion = "22.22.2"

func nodeDownloadURL(p Platform) string {
	arch := p.Arch
	if arch == "amd64" {
		arch = "x64"
	}

	var ext string
	var osName string
	switch p.OS {
	case "windows":
		osName = "win"
		ext = "zip"
	case "linux":
		osName = "linux"
		ext = "tar.gz"
	case "darwin":
		osName = "darwin"
		ext = "tar.gz"
	}

	return fmt.Sprintf("https://nodejs.org/dist/v%s/node-v%s-%s-%s.%s",
		nodeVersion, nodeVersion, osName, arch, ext)
}

// nodeDirName returns the extracted directory name inside the archive.
func nodeDirName(p Platform) string {
	arch := p.Arch
	if arch == "amd64" {
		arch = "x64"
	}
	osName := p.OS
	if p.OS == "windows" {
		osName = "win"
	}
	return fmt.Sprintf("node-v%s-%s-%s", nodeVersion, osName, arch)
}

// NodeBinPath returns the path to the node executable.
func NodeBinPath(sandboxDir string, p Platform) string {
	if p.OS == "windows" {
		return filepath.Join(sandboxDir, "node", "node.exe")
	}
	return filepath.Join(sandboxDir, "node", "bin", "node")
}

// NpmBinPath returns the path to the npm executable.
func NpmBinPath(sandboxDir string, p Platform) string {
	if p.OS == "windows" {
		return filepath.Join(sandboxDir, "node", "npm.cmd")
	}
	return filepath.Join(sandboxDir, "node", "bin", "npm")
}

// DownloadNode downloads and extracts Node.js to sandboxDir/node/.
func DownloadNode(sandboxDir string, p Platform) error {
	nodeDir := filepath.Join(sandboxDir, "node")
	if _, err := os.Stat(NodeBinPath(sandboxDir, p)); err == nil {
		fmt.Println("  Node already installed, skipping.")
		return nil
	}

	url := nodeDownloadURL(p)
	archiveName := filepath.Base(url)
	archivePath := filepath.Join(sandboxDir, archiveName)

	fmt.Printf("  Downloading Node %s...\n", nodeVersion)
	if err := DownloadFile(url, archivePath); err != nil {
		return fmt.Errorf("download node: %w", err)
	}
	defer os.Remove(archivePath)

	// Extract directly to target directory
	os.RemoveAll(nodeDir)
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		return err
	}

	fmt.Println("  Extracting Node...")
	if err := ExtractArchive(archivePath, nodeDir); err != nil {
		return fmt.Errorf("extract node: %w", err)
	}

	// If archive contained a subdirectory, move its contents up
	subDir := filepath.Join(nodeDir, nodeDirName(p))
	if info, err := os.Stat(subDir); err == nil && info.IsDir() {
		entries, _ := os.ReadDir(subDir)
		for _, e := range entries {
			os.Rename(filepath.Join(subDir, e.Name()), filepath.Join(nodeDir, e.Name()))
		}
		os.RemoveAll(subDir)
	}

	fmt.Println("  ✅ Node installed.")
	return nil
}
