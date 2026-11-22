package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yyYank/anw/internal/history"
	"github.com/yyYank/anw/internal/tidy"
)

var tidyBy string

var tidyCmd = &cobra.Command{
	Use:   "tidy",
	Short: "ファイルを更新日時ごとのディレクトリに振り分ける",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		by, err := tidy.ParseBy(tidyBy)
		if err != nil {
			return err
		}
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		moves, err := tidy.Run(dir, by)
		// 途中で失敗しても、移動済みの分は元に戻せるように保存する
		if len(moves) > 0 {
			path, herr := historyPath()
			if herr != nil {
				return errors.Join(err, herr)
			}
			if herr := history.Save(path, moves); herr != nil {
				return errors.Join(err, herr)
			}
		}
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
