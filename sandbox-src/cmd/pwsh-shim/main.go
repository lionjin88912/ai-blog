// pwsh-shim is a tiny executable that masquerades as pwsh.exe.
// Copilot CLI requires PowerShell 6+ (pwsh); this shim translates
// pwsh-style invocations into bash calls so everything works inside
// the MINGW64 sandbox shell.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func main() {
	args := os.Args[1:]

	// pwsh.exe --version → fake a PowerShell version so Copilot CLI is satisfied.
	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v") {
		fmt.Println("PowerShell 7.4.0")
		os.Exit(0)
	}

	// Locate bash.exe next to this binary (sandbox/bin/) or fallback to
	// the portable git bash bundled in the sandbox.
	bashPath := findBash()
	if bashPath == "" {
		fmt.Fprintln(os.Stderr, "pwsh-shim: cannot find bash.exe")
		os.Exit(1)
	}

	// Translate pwsh arguments to bash arguments.
	// pwsh -Command "..." | pwsh -c "..."  →  bash -c "..."
	// pwsh -File script.ps1               →  bash script.ps1
	// pwsh <anything else>                →  bash <anything else>
	bashArgs := translateArgs(args)

	cmd := exec.Command(bashPath, bashArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		os.Exit(1)
	}
}

func translateArgs(args []string) []string {
	if len(args) == 0 {
		return []string{"-i"}
	}

	// Collect the command after -Command / -c flag
	for i, a := range args {
		lower := strings.ToLower(a)
		if lower == "-command" || lower == "-c" {
			// Everything after -Command is the command string
			rest := args[i+1:]
			if len(rest) == 0 {
				return []string{"-i"}
			}
			return []string{"-c", strings.Join(rest, " ")}
		}
		if lower == "-file" {
			return args[i+1:]
		}
		if lower == "-noprofile" || lower == "-nologo" || lower == "-noninteractive" {
			// Skip pwsh-specific flags, continue
			continue
		}
	}

	// Fallback: pass everything as a -c command
	return []string{"-c", strings.Join(args, " ")}
}

func findBash() string {
	// 1. Look for bash.exe relative to this binary: ../git/bin/bash.exe
	self, err := os.Executable()
	if err == nil {
		sandboxDir := filepath.Dir(filepath.Dir(self)) // sandbox/bin/ → sandbox/
		candidate := filepath.Join(sandboxDir, "git", "bin", "bash.exe")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// 2. Look on PATH
	if p, err := exec.LookPath("bash.exe"); err == nil {
		return p
	}

	return ""
}
