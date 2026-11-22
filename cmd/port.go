package cmd

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"syscall"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yyYank/anw/internal/port"
)

// テストで差し替えられるよう変数にしておく
var (
	listListeners  = port.List
	killProcess    = func(pid int) error { return syscall.Kill(pid, syscall.SIGTERM) }
	selectListener = runSelector
)

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
		if err != nil {
			return err
		}
		if !ok {
			return errCanceled
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

// viewPort は LISTEN 中のポートを表で出力する。ポート番号があればそのポートだけにする。
func viewPort(w io.Writer, args []string) error {
	n, hasPort, err := parsePort(args)
	if err != nil {
		return err
	}
	ls, err := listListeners()
	if err != nil {
		return err
	}
	if hasPort {
		ls = port.Find(ls, n)
		if len(ls) == 0 {
			return fmt.Errorf("%d番を使っているプロセスはありません", n)
		}
	}
	if len(ls) == 0 {
		return errors.New("LISTEN 中のポートはありません")
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PORT\tPID\tNAME\tCOMMAND")
	for _, l := range ls {
		fmt.Fprintf(tw, "%d\t%d\t%s\t%s\n", l.Port, l.PID, l.Name, l.Command)
	}
	return tw.Flush()
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
