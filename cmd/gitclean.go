package cmd

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yyYank/anw/internal/history"
	"github.com/yyYank/anw/internal/multiselect"
	"github.com/yyYank/anw/internal/stash"
	"github.com/yyYank/anw/internal/worktree"
)

// テストで差し替えられるよう変数にしておく
var (
	listWorktrees  = worktree.List
	removeWorktree = worktree.Remove
	listStashes    = stash.List
	dropStashes    = stash.Drop
	selectItems    = runMultiSelect
)

// listStaleWorktrees は消してよい worktree を返す。なければエラーにする。
func listStaleWorktrees() ([]worktree.Worktree, error) {
	ws, err := listWorktrees()
	if err != nil {
		return nil, err
	}
	if len(ws) == 0 {
		return nil, errors.New("消してよい worktree はありません")
	}
	return ws, nil
}

// viewWorktrees は消してよい worktree と、その理由を表で出力する。
func viewWorktrees(w io.Writer) error {
	ws, err := listStaleWorktrees()
	if err != nil {
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PATH\tBRANCH\tREASON")
	for _, wt := range ws {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", wt.Path, wt.Branch, wt.Reason())
	}
	return tw.Flush()
}

// deleteWorktrees は一覧から選んだ worktree を削除する。force なら変更が残っていても削除する。
// 削除できないものがあっても残りは削除し、最後にエラーを返す。
func deleteWorktrees(out, errOut io.Writer, force bool) error {
	ws, err := listStaleWorktrees()
	if err != nil {
		return err
	}
	items := make([]multiselect.Item, len(ws))
	for i, wt := range ws {
		items[i] = multiselect.Item{Name: wt.Path, Reason: wt.Reason()}
	}
	picked, err := selectItems("削除する worktree を選んでください (↑↓/jk: 移動, Space: 選択, Enter: 削除, q: やめる)", items)
	if err != nil {
		return err
	}

	failed := 0
	for _, i := range picked {
		if err := removeWorktree(ws[i].Path, force); err != nil {
			fmt.Fprintf(errOut, "削除できませんでした: %s: %v\n", ws[i].Path, err)
			failed++
			continue
		}
		fmt.Fprintf(out, "removed %s\n", ws[i].Path)
	}
	if failed > 0 {
		if !force {
			return fmt.Errorf("%d件の worktree を削除できませんでした (--force で強制削除できます)", failed)
		}
		return fmt.Errorf("%d件の worktree を削除できませんでした", failed)
	}
	return nil
}

// listAllStashes は stash をすべて返す。なければエラーにする。
func listAllStashes() ([]stash.Stash, error) {
	ss, err := listStashes()
	if err != nil {
		return nil, err
	}
	if len(ss) == 0 {
		return nil, errors.New("stash はありません")
	}
	return ss, nil
}

// viewStashes は stash を、元のブランチと何日前かの表で出力する。
func viewStashes(w io.Writer) error {
	ss, err := listAllStashes()
	if err != nil {
		return err
	}
	now := time.Now()
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "STASH\tMESSAGE\tBRANCH")
	for _, s := range ss {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", s.Ref, s.Message, s.Reason(now))
	}
	return tw.Flush()
}

// deleteStashes は一覧から選んだ stash を消し、undo で作り直せるよう履歴に書く。
func deleteStashes(out io.Writer) error {
	ss, err := listAllStashes()
	if err != nil {
		return err
	}
	now := time.Now()
	items := make([]multiselect.Item, len(ss))
	for i, s := range ss {
		items[i] = multiselect.Item{Name: s.Ref + " " + s.Message, Reason: s.Reason(now)}
	}
	picked, err := selectItems("削除する stash を選んでください (↑↓/jk: 移動, Space: 選択, Enter: 削除, q: やめる)", items)
	if err != nil || len(picked) == 0 {
		return err
	}

	targets := make([]stash.Stash, len(picked))
	for i, p := range picked {
		targets[i] = ss[p]
	}
	saved, dropErr := dropStashes(targets)
	for _, s := range saved {
		fmt.Fprintf(out, "dropped %s\n", s.Message)
	}
	// 途中で失敗しても、消した分は作り直せるように保存する
	if len(saved) > 0 {
		path, err := historyPath()
		if err == nil {
			err = history.SaveStashes(path, saved)
		}
		if err != nil {
			return errors.Join(dropErr, fmt.Errorf("消した stash を履歴に書けませんでした: %w", err))
		}
	}
	return dropErr
}

// runMultiSelect は複数選択の画面を表示し、選ばれた要素の添字を返す。
func runMultiSelect(title string, items []multiselect.Item) ([]int, error) {
	m, err := tea.NewProgram(multiselect.New(title, items)).Run()
	if err != nil {
		return nil, err
	}
	return m.(multiselect.Model).Selected(), nil
}
