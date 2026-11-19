package cmd

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yyYank/anw/internal/branch"
)

// テストで差し替えられるよう変数にしておく
var (
	listBranches   = branch.List
	selectBranches = runBranchSelector
	deleteBranch   = branch.Delete
)

// listDeletableBranches は削除できるブランチを返す。なければエラーにする。
func listDeletableBranches() ([]branch.Branch, error) {
	bs, err := listBranches()
	if err != nil {
		return nil, err
	}
	if len(bs) == 0 {
		return nil, errors.New("削除できるブランチはありません")
	}
	return bs, nil
}

// viewBranches は削除できるブランチと、その理由を表で出力する。
func viewBranches(w io.Writer) error {
	bs, err := listDeletableBranches()
	if err != nil {
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "BRANCH\tREASON")
	for _, b := range bs {
		fmt.Fprintf(tw, "%s\t%s\n", b.Name, b.Reason())
	}
	return tw.Flush()
}

// deleteBranches は一覧から選んだブランチを削除する。force なら強制削除する。
// 削除できないブランチがあっても残りは削除し、最後にエラーを返す。
func deleteBranches(out, errOut io.Writer, force bool) error {
	bs, err := listDeletableBranches()
	if err != nil {
		return err
	}
	picked, err := selectBranches(bs)
	if err != nil {
		return err
	}

	failed := 0
	for _, b := range picked {
		if err := deleteBranch(b.Name, force); err != nil {
			fmt.Fprintf(errOut, "削除できませんでした: %s: %v\n", b.Name, err)
			failed++
			continue
		}
		fmt.Fprintf(out, "deleted %s\n", b.Name)
	}
	if failed > 0 {
		if !force {
			return fmt.Errorf("%d件のブランチを削除できませんでした (--force で強制削除できます)", failed)
		}
		return fmt.Errorf("%d件のブランチを削除できませんでした", failed)
	}
	return nil
}

// runBranchSelector は一覧の画面を表示し、選ばれたブランチを返す。
func runBranchSelector(bs []branch.Branch) ([]branch.Branch, error) {
	m, err := tea.NewProgram(branch.NewSelector(bs)).Run()
	if err != nil {
		return nil, err
	}
	return m.(branch.Selector).Selected(), nil
}
