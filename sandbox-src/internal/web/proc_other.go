//go:build !windows

package web

import "os/exec"

// setSysProcAttr is a no-op on non-Windows platforms.
func setSysProcAttr(cmd *exec.Cmd) {}
