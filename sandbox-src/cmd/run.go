package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ai-sandbox/cli/internal/toolchain"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [command] [args...]",
	Short: "Run a command with sandbox tools in PATH",
	Long:  `Execute a command with ./sandbox/bin prepended to PATH.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		absDir, err := filepath.Abs(sandboxDir)
		if err != nil {
			return err
		}

		binDir := toolchain.SandboxBinDir(absDir)
		if _, err := os.Stat(binDir); os.IsNotExist(err) {
			return fmt.Errorf("sandbox not set up. Run 'ai-sandbox setup' first")
		}

		newPath := binDir + string(os.PathListSeparator) + os.Getenv("PATH")

		execCmd := exec.Command(args[0], args[1:]...)
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execCmd.Env = append(os.Environ(), "PATH="+newPath)

		return execCmd.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
