package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ai-sandbox/cli/internal/toolchain"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open a shell with sandbox tools in PATH",
	Long:  `Open an interactive shell with ./sandbox/bin prepended to PATH so all sandbox tools are available.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		absDir, err := filepath.Abs(sandboxDir)
		if err != nil {
			return err
		}

		binDir := toolchain.SandboxBinDir(absDir)
		if _, err := os.Stat(binDir); os.IsNotExist(err) {
			return fmt.Errorf("sandbox not set up. Run 'ai-blog setup' first")
		}

		newPath := binDir + string(os.PathListSeparator) + os.Getenv("PATH")

		shellBin, shellArgs := ResolveShell(shellFlag, absDir)

		shellExec := exec.Command(shellBin, shellArgs...)
		shellExec.Stdin = os.Stdin
		shellExec.Stdout = os.Stdout
		shellExec.Stderr = os.Stderr
		shellExec.Env = append(os.Environ(), "PATH="+newPath)



		// Tell child tools (e.g. Copilot CLI) to use bash instead of pwsh
		if runtime.GOOS == "windows" && shellBin != "powershell.exe" {
			shellExec.Env = append(shellExec.Env, "SHELL="+shellBin)
		}

		fmt.Println("Entering sandbox shell. Type 'exit' to leave.")
		return shellExec.Run()
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
}
