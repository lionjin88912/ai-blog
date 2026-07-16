// pwsh-shim is a tiny executable that masquerades as pwsh.exe.
//
// Copilot CLI (and some tools) require "pwsh" (PowerShell 7). Windows ships
// only powershell.exe (5.1) by default. This shim answers the version probe,
// then forwards everything verbatim to a real PowerShell — preferring an
// actually-installed pwsh.exe, else the built-in powershell.exe.
//
// It delegates to PowerShell, NOT bash: pwsh-style args (-Command / -File /
// -NoProfile) and .ps1 scripts are PowerShell syntax, so running them through
// bash breaks. Delegating to powershell.exe keeps everything native — which
// matters now that agy runs outside the MSYS/git-bash environment.
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

	// Version probe → claim PowerShell 7 so Copilot CLI is satisfied.
	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v") {
		fmt.Println("PowerShell 7.4.0")
		os.Exit(0)
	}

	target := findPowerShell()

	// Forward args verbatim: powershell.exe accepts the same -Command / -File /
	// -NoProfile / -NoLogo flags as pwsh.
	cmd := exec.Command(target, args...)
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

// findPowerShell returns a real PowerShell to delegate to: an installed
// pwsh.exe if present (but never THIS shim), else the built-in powershell.exe.
func findPowerShell() string {
	self, _ := os.Executable()
	if p, err := exec.LookPath("pwsh.exe"); err == nil {
		if abs, _ := filepath.Abs(p); !strings.EqualFold(abs, self) {
			return p // a genuine PowerShell 7 install
		}
	}
	return "powershell.exe" // built into every Windows
}
