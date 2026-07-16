package toolchain

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ai-sandbox/cli/internal/pwshshim"
)

// WriteShims creates wrapper scripts in sandboxDir/bin/ that point to the actual binaries.
func WriteShims(sandboxDir string, p Platform) error {
	binDir := filepath.Join(sandboxDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	type shim struct {
		name   string
		target string
	}

	shims := []shim{
		{"node", NodeBinPath(sandboxDir, p)},
		{"npm", NpmBinPath(sandboxDir, p)},
		{"copilot", CopilotBinPath(sandboxDir, p)},
		{"uv", UVBinPath(sandboxDir, p)},
		{"python", PythonBinPath(sandboxDir, p)},
		{"python3", PythonBinPath(sandboxDir, p)}, // SOPs call python3; needed in native cmd mode
		{"git", GitBinPath(sandboxDir, p)},
		{"agy", AntigravityBinPath(sandboxDir, p)},
	}

	for _, s := range shims {
		if err := writeShim(binDir, s.name, s.target, p); err != nil {
			return fmt.Errorf("write shim %s: %w", s.name, err)
		}
	}



	// On Windows, create a pwsh shim that delegates to mingw64 bash.
	// Copilot CLI expects pwsh; this makes it use bash instead.
	if p.OS == "windows" {
		bashPath := GitBashPath(sandboxDir, p)
		if err := writePwshShim(binDir, bashPath); err != nil {
			return fmt.Errorf("write pwsh shim: %w", err)
		}
	}

	fmt.Println("  ✅ Shim scripts created in sandbox/bin/")
	return nil
}

func writeShim(binDir, name, target string, p Platform) error {
	// Compute path relative to the shim's directory (sandbox/bin/).
	// This makes the sandbox portable — it works from any location.
	relTarget, err := filepath.Rel(binDir, target)
	if err != nil {
		relTarget = target
	}

	if p.OS == "windows" {
		// %~dp0 is the directory of the .cmd file itself (with trailing \)
		// Convert forward slashes for batch compatibility
		batchRel := filepath.ToSlash(relTarget)
		shimPath := filepath.Join(binDir, name+".cmd")
		content := fmt.Sprintf("@echo off\r\n\"%%~dp0%s\" %%*\r\n", batchRel)
		if err := os.WriteFile(shimPath, []byte(content), 0755); err != nil {
			return err
		}

		// Shell script (no extension) for MINGW64 bash
		// $(dirname "$0") resolves to the shim's directory at runtime
		// Strip .cmd extension — bash needs the extensionless shell script in node_modules/.bin/
		bashTarget := filepath.ToSlash(relTarget)
		if ext := filepath.Ext(bashTarget); ext == ".cmd" {
			bashTarget = bashTarget[:len(bashTarget)-len(ext)]
		}
		shimPathSh := filepath.Join(binDir, name)
		contentSh := fmt.Sprintf("#!/bin/sh\nexec \"$(dirname \"$0\")/%s\" \"$@\"\n", bashTarget)
		return os.WriteFile(shimPathSh, []byte(contentSh), 0755)
	}

	// Unix shell script
	shimPath := filepath.Join(binDir, name)
	content := fmt.Sprintf("#!/bin/sh\nexec \"$(dirname \"$0\")/%s\" \"$@\"\n", filepath.ToSlash(relTarget))
	return os.WriteFile(shimPath, []byte(content), 0755)
}

// SandboxBinDir returns the path to the sandbox bin directory.
func SandboxBinDir(sandboxDir string) string {
	return filepath.Join(sandboxDir, "bin")
}



// writePwshShim writes the embedded pwsh.exe shim into the bin directory.
// The shim (cmd/pwsh-shim) delegates to native powershell.exe. Embedding it
// means the single distributed launcher is self-contained — no pwsh.exe beside
// the exe and no Go toolchain required on the user's machine.
func writePwshShim(binDir, bashPath string) error {
	if len(pwshshim.Binary) == 0 {
		return fmt.Errorf("embedded pwsh shim is empty (built without the pwsh embed)")
	}
	dst := filepath.Join(binDir, "pwsh.exe")
	if err := os.WriteFile(dst, pwshshim.Binary, 0755); err != nil {
		return fmt.Errorf("write pwsh.exe: %w", err)
	}
	// Remove any stale script fallbacks from older versions so PATH resolves
	// to pwsh.exe, not a leftover pwsh.cmd/pwsh that delegates to bash.
	_ = os.Remove(filepath.Join(binDir, "pwsh.cmd"))
	_ = os.Remove(filepath.Join(binDir, "pwsh"))
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0755)
}
