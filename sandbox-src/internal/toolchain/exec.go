package toolchain

import "os/exec"

// execCommand wraps exec.Command for consistency.
var execCommand = exec.Command
