package cmd

import (
	"fmt"

	"github.com/ai-sandbox/cli/internal/config"
	"github.com/ai-sandbox/cli/internal/wizard"
	"github.com/spf13/cobra"
)

var forceInit bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Configure API keys and workspace path",
	Long:  `Run the first-time setup wizard to configure Gemini API key, GitHub token, and workspace directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if config.Exists() && !forceInit {
			fmt.Println("Already configured. Use --force to re-run setup.")
			return nil
		}

		cfg, err := wizard.Run()
		if err != nil {
			return err
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println()
		fmt.Println("✅ Configuration saved!")
		fmt.Println("Run 'ai-blog setup' to download tools.")
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Re-run setup even if already configured")
	rootCmd.AddCommand(initCmd)
}
