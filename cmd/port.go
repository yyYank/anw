package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/yyYank/anw/internal/port"
)

// テストで差し替えられるよう変数にしておく
var (
	listListeners  = port.List
	killProcess    = func(pid int) error { return syscall.Kill(pid, syscall.SIGTERM) }
	selectListener = runSelector
)

var portCmd = &cobra.Command{
	Use:   "port [ポート番号]",
	Short: "ポートを使っているプロセスを確認して止める",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n := 0
		if len(args) == 1 {
			var err error
			if n, err = strconv.Atoi(args[0]); err != nil {
				return fmt.Errorf("ポート番号が数字ではありません: %q", args[0])
			}
		}
		ls, err := listListeners()
		if err != nil {
			return err
		}

		var targets []port.Listener
		if len(args) == 1 {
			targets = port.Find(ls, n)
			if len(targets) == 0 {
				return fmt.Errorf("%d番を使っているプロセスはありません", n)
			}
		} else {
			if len(ls) == 0 {
				return errors.New("LISTEN 中のポートはありません")
			}
			l, ok, err := selectListener(ls)
			if err != nil || !ok {
				return err
			}
			targets = []port.Listener{l}
		}

		for _, l := range targets {
			if err := killProcess(l.PID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "kill %d (%s) :%d\n", l.PID, l.Name, l.Port)
		}
		return nil
	},
}

// runSelector は一覧の画面を表示し、選ばれた Listener を返す。
func runSelector(ls []port.Listener) (port.Listener, bool, error) {
	m, err := tea.NewProgram(port.NewSelector(ls)).Run()
	if err != nil {
		return port.Listener{}, false, err
	}
	l, ok := m.(port.Selector).Selected()
	return l, ok, nil
}

func init() {
	rootCmd.AddCommand(portCmd)
}
