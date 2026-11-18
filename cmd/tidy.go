package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yyYank/anw/internal/tidy"
)

var tidyBy string

var tidyCmd = &cobra.Command{
	Use:   "tidy",
	Short: "ファイルを更新日時ごとのディレクトリに振り分ける",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if tidyBy != "date" {
			return fmt.Errorf("--by に指定できるのは date のみです: %q", tidyBy)
		}
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		moves, err := tidy.Run(dir)
		for _, m := range moves {
			from, _ := filepath.Rel(dir, m.From)
			to, _ := filepath.Rel(dir, m.To)
			fmt.Fprintf(cmd.OutOrStdout(), "%s -> %s\n", from, to)
		}
		return err
	},
}

func init() {
	tidyCmd.Flags().StringVar(&tidyBy, "by", "date", "振り分けの基準 (date)")
	rootCmd.AddCommand(tidyCmd)
}
