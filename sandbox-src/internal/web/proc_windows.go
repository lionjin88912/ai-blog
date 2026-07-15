//go:build windows

package web

import (
	"os/exec"
	"syscall"
)

// setSysProcAttr configures Windows-specific process attributes.
// CREATE_NEW_PROCESS_GROUP + HideWindow lets winpty allocate a hidden console.
func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    true,
	}
}
