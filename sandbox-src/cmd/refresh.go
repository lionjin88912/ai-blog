package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ai-sandbox/cli/internal/seed"
	"github.com/spf13/cobra"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh-skills",
	Short: "Update factory skills/SOP to the version embedded in this binary",
	Long: `Overwrite factory content (.agents/skills SOP files, GEMINI.md, docs) with
the seed embedded in this executable.

Persona user data is never touched: personas/<slug>/ (except _template),
wp-config.json, published.json and articles/ are always preserved.
Overwritten files are backed up under .agents/_backup/<stamp>/ first.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		// Guard: refuse to scaffold a fresh tree into a random directory.
		// A real workspace has already been seeded (.agents/.seed-version)
		// or at least contains .agents/.
		if seed.LocalVersion(wd) == "" {
			if _, statErr := os.Stat(filepath.Join(wd, ".agents")); os.IsNotExist(statErr) {
				return fmt.Errorf("目前目錄不是已初始化的 workspace(找不到 .agents/)。請在雙擊 exe 產生的資料夾內執行,或先跑一次 setup")
			}
		}

		fmt.Printf("Seed version: %s (local: %s)\n", seed.Version(), orNone(seed.LocalVersion(wd)))
		st, err := seed.Materialize(wd, true)
		if err != nil {
			return err
		}
		fmt.Printf("✅ 更新完成:新增 %d 檔、更新 %d 檔、保留 %d 檔(人格資料一律不動)\n",
			len(st.Created), len(st.Updated), st.Skipped)
		for _, f := range st.Updated {
			fmt.Printf("   ↻ %s\n", f)
		}
		for _, f := range st.Created {
			fmt.Printf("   + %s\n", f)
		}
		if len(st.Updated) > 0 {
			fmt.Println("   舊版已備份到 .agents/_backup/")
		}
		return nil
	},
}

func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

// AppVersion exposes the CLI version for other packages (web server badge).
func AppVersion() string { return version }

func init() {
	rootCmd.AddCommand(refreshCmd)
}
