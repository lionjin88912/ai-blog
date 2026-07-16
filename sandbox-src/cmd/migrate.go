package cmd

import (
	"fmt"

	"github.com/ai-sandbox/cli/internal/config"
	"github.com/ai-sandbox/cli/internal/migrate"
	"github.com/spf13/cobra"
)

var (
	migrateFrom   string
	migrateDryRun bool
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Import personas + WordPress settings from an old install",
	Long: `Copy your personas and WordPress connection settings from an older
version's folder into the current data directory.

Safe by design: the old folder is only read (never modified), wp-config.json is
cleaned through a known-key allowlist (legacy junk dropped, auth_method inferred
if missing), and personas that already exist are skipped.

With no --from, it scans Downloads/Desktop/Documents for old installs and lists
what it finds.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir, err := config.DataDir()
		if err != nil {
			return err
		}
		target := migrate.TargetPersonasDir(dataDir)

		if migrateFrom == "" {
			cands := migrate.FindCandidates(migrate.DefaultSearchDirs())
			if len(cands) == 0 {
				fmt.Println("沒有在 Downloads/Desktop/Documents 找到含人格的舊資料夾。")
				fmt.Println("若知道舊資料夾位置,請用:ai-blog migrate --from \"<路徑>\"")
				return nil
			}
			fmt.Println("找到可能的舊安裝(含人格):")
			for _, c := range cands {
				fmt.Printf("  • %s(%d 個人格)\n", c, len(migrate.Scan(c)))
			}
			fmt.Println("\n要匯入其中一個,請用:ai-blog migrate --from \"<上面的路徑>\"")
			fmt.Println("先預覽不寫入,可加 --dry-run")
			return nil
		}

		rep, err := migrate.Migrate(migrateFrom, target, migrateDryRun)
		if err != nil {
			return err
		}
		fmt.Print(migrate.FormatReport(rep))
		if rep.Imported > 0 && !migrateDryRun {
			fmt.Println("完成。重開 Antigravity 後即可看到匯入的人格。舊資料夾未更動,確認無誤後可自行刪除。")
		}
		return nil
	},
}

func init() {
	migrateCmd.Flags().StringVar(&migrateFrom, "from", "", "舊安裝資料夾路徑")
	migrateCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "只預覽掃描結果,不實際寫入")
	rootCmd.AddCommand(migrateCmd)
}
