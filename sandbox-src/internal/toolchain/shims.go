package toolchain

import (
	"fmt"
	"os"
	"path/filepath"
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



// writePwshShim creates pwsh.exe in the bin directory.
// If cmd/pwsh-shim was pre-built as pwsh.exe next to the main binary, copy it.
// Otherwise fall back to script-based shims.
func writePwshShim(binDir, bashPath string) error {
	relBash, err := filepath.Rel(binDir, bashPath)
	if err != nil {
		relBash = bashPath
	}

	// Try to find a pre-built pwsh.exe next to the running binary
	self, err := os.Executable()
	if err == nil {
		prebuilt := filepath.Join(filepath.Dir(self), "pwsh.exe")
		if _, err := os.Stat(prebuilt); err == nil {
			dst := filepath.Join(binDir, "pwsh.exe")
			if err := copyFile(prebuilt, dst); err != nil {
				return fmt.Errorf("copy pwsh.exe: %w", err)
			}
			return nil
		}
	}

	// Try to build pwsh.exe on the fly if Go is available
	if goPath, lookErr := execCommand("go", "env", "GOEXE").CombinedOutput(); lookErr == nil {
		_ = goPath // go is available
		dst := filepath.Join(binDir, "pwsh.exe")
		build := execCommand("go", "build", "-ldflags", "-s -w", "-o", dst, "./cmd/pwsh-shim")
		if err := build.Run(); err == nil {
			return nil
		}
	}

	// Fallback: .cmd + shell script shims with relative paths
	batchRel := filepath.ToSlash(relBash)
	cmdContent := fmt.Sprintf("@echo off\r\n\"%%~dp0%s\" %%*\r\n", batchRel)
	if err := os.WriteFile(filepath.Join(binDir, "pwsh.cmd"), []byte(cmdContent), 0755); err != nil {
		return err
	}
	shContent := fmt.Sprintf("#!/bin/sh\nexec \"$(dirname \"$0\")/%s\" \"$@\"\n", filepath.ToSlash(relBash))
	return os.WriteFile(filepath.Join(binDir, "pwsh"), []byte(shContent), 0755)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0755)
}
