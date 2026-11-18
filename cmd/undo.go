package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yyYank/anw/internal/history"
)

var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "直前のファイル操作を元に戻す",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := historyPath()
		if err != nil {
			return err
		}
		moves, err := history.Undo(path)
		for _, m := range moves {
			fmt.Fprintf(cmd.OutOrStdout(), "%s -> %s\n", m.To, m.From)
		}
		if errors.Is(err, os.ErrNotExist) && len(moves) == 0 {
			return errors.New("元に戻せる操作がありません")
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(undoCmd)
}
