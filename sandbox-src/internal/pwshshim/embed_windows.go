//go:build windows

// Package pwshshim carries the prebuilt pwsh.exe shim so the single distributed
// exe is self-contained: WriteShims materializes it into sandbox/bin without
// needing a pwsh.exe next to the launcher or a Go toolchain on the machine.
package pwshshim

import _ "embed"

//go:embed pwsh_windows_amd64.exe
var Binary []byte
