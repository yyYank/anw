package cmd

import (
	"errors"
	"fmt"
	"io"
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
		return killPort(cmd.OutOrStdout(), args)
	},
}

// parsePort は引数のポート番号を返す。引数がなければ ok は false。
func parsePort(args []string) (n int, ok bool, err error) {
	if len(args) == 0 {
		return 0, false, nil
	}
	if n, err = strconv.Atoi(args[0]); err != nil {
		return 0, false, fmt.Errorf("ポート番号が数字ではありません: %q", args[0])
	}
	return n, true, nil
}

// killPort は指定ポートを使っているプロセスを kill する。ポート番号がなければ一覧から選ばせる。
func killPort(w io.Writer, args []string) error {
	n, hasPort, err := parsePort(args)
	if err != nil {
		return err
	}
	ls, err := listListeners()
	if err != nil {
		return err
	}

	var targets []port.Listener
	if hasPort {
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
		fmt.Fprintf(w, "kill %d (%s) :%d\n", l.PID, l.Name, l.Port)
	}
	return nil
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
