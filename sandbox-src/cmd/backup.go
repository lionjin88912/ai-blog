package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ai-sandbox/cli/internal/backup"
	"github.com/ai-sandbox/cli/internal/config"
	"github.com/ai-sandbox/cli/internal/migrate"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Export/import personas + WordPress settings to hand work over",
}

var backupExportCmd = &cobra.Command{
	Use:   "export [output.zip]",
	Short: "Export this machine's personas + WordPress settings to a backup zip",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir, err := config.DataDir()
		if err != nil {
			return err
		}
		data, name, err := backup.Export(dataDir, time.Now())
		if err != nil {
			return err
		}
		out := name
		if len(args) > 0 {
			out = args[0]
		} else if OriginalWD() != "" {
			out = filepath.Join(OriginalWD(), name)
		}
		if err := os.WriteFile(out, data, 0o644); err != nil {
			return err
		}
		fmt.Printf("✅ 備份完成:%s\n", out)
		fmt.Println("⚠️  此檔含 WordPress 帳密,請用公司安全管道傳遞,交接完成後刪除。")
		return nil
	},
}

var backupImportCmd = &cobra.Command{
	Use:   "import <backup.zip>",
	Short: "Import personas + WordPress settings from a backup zip",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir, err := config.DataDir()
		if err != nil {
			return err
		}
		path := args[0]
		if !filepath.IsAbs(path) && OriginalWD() != "" {
			path = filepath.Join(OriginalWD(), path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("讀不到備份檔: %w", err)
		}
		rep, err := backup.Import(data, dataDir)
		if err != nil {
			return err
		}
		fmt.Print(migrate.FormatReport(rep))
		if (rep.Imported + rep.Merged) > 0 {
			fmt.Println("完成。重開 Antigravity 後即可用匯入的人格繼續工作。")
		}
		return nil
	},
}

func init() {
	backupCmd.AddCommand(backupExportCmd)
	backupCmd.AddCommand(backupImportCmd)
	rootCmd.AddCommand(backupCmd)
}
