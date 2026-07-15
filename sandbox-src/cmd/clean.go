package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var forceClean bool

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove the sandbox directory",
	Long:  `Delete the ./sandbox/ directory and all downloaded tools.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		absDir, err := filepath.Abs(sandboxDir)
		if err != nil {
			return err
		}

		if _, err := os.Stat(absDir); os.IsNotExist(err) {
			fmt.Println("Sandbox directory does not exist. Nothing to clean.")
			return nil
		}

		if !forceClean {
			fmt.Printf("This will delete %s and all contents. Continue? [y/N]: ", absDir)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "y") {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := os.RemoveAll(absDir); err != nil {
			return fmt.Errorf("remove sandbox: %w", err)
		}

		fmt.Println("✅ Sandbox removed.")
		return nil
	},
}

func init() {
	cleanCmd.Flags().BoolVarP(&forceClean, "force", "f", false, "Skip confirmation")
	rootCmd.AddCommand(cleanCmd)
}
