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
	Short: "直前の操作（tidy・delete --branch）を元に戻す",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := historyPath()
		if err != nil {
			return err
		}
		r, err := history.Undo(path)
		// snippet の記録は、各ステップで戻した内容も表示する
		for _, s := range append([]history.Record{r}, r.Steps...) {
			for _, m := range s.Moves {
				fmt.Fprintf(cmd.OutOrStdout(), "%s -> %s\n", m.To, m.From)
			}
			for _, b := range s.Branches {
				fmt.Fprintf(cmd.OutOrStdout(), "restored %s (%s)\n", b.Name, b.Commit)
			}
		}
		if errors.Is(err, os.ErrNotExist) && len(r.Moves) == 0 {
			return errors.New("元に戻せる操作がありません")
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(undoCmd)
}
