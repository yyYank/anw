package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/yyYank/anw/internal/branch"
	"github.com/yyYank/anw/internal/history"
)

var deletable = []branch.Branch{
	{Name: "feature/a", Merged: true},
	{Name: "feature/b", Gone: true},
	{Name: "feature/c", Merged: true},
}

// delete --branch で、一覧から選んだブランチを強制せずに削除すること
func TestDelete_選んだブランチを削除する(t *testing.T) {
	cli := setupCLI(t)
	cli.branches = deletable
	cli.branchPick = []int{0, 2}

	if err := cli.run("delete", "--branch"); err != nil {
		t.Fatal(err)
	}

	if want := []string{"feature/a", "feature/c"}; !reflect.DeepEqual(cli.deleted, want) {
		t.Errorf("deleted %v, want %v", cli.deleted, want)
	}
}

// --force を付けると強制削除すること
func TestDelete_forceで強制削除する(t *testing.T) {
	cli := setupCLI(t)
	cli.branches = deletable
	cli.branchPick = []int{1}

	if err := cli.run("delete", "--force", "--branch"); err != nil {
		t.Fatal(err)
	}

	if want := []string{"feature/b!"}; !reflect.DeepEqual(cli.deleted, want) {
		t.Errorf("deleted %v, want %v", cli.deleted, want)
	}
}

// 削除できないブランチがあっても残りは削除し、最後にエラーにすること
func TestDelete_失敗しても残りを削除する(t *testing.T) {
	cli := setupCLI(t)
	cli.branches = deletable
	cli.branchPick = []int{0, 1, 2}
	cli.deleteFail = map[string]bool{"feature/b": true}

	if err := cli.run("delete", "--branch"); err == nil {
		t.Error("エラーにならなかった")
	}

	if want := []string{"feature/a", "feature/c"}; !reflect.DeepEqual(cli.deleted, want) {
		t.Errorf("deleted %v, want %v", cli.deleted, want)
	}
}

// 一覧で選ばずにやめたら何も削除しないこと
func TestDelete_やめたら削除しない(t *testing.T) {
	cli := setupCLI(t)
	cli.branches = deletable

	if err := cli.run("delete", "--branch"); err != nil {
		t.Fatal(err)
	}

	if len(cli.deleted) != 0 {
		t.Errorf("削除された: %v", cli.deleted)
	}
}

// 削除できるブランチがなければ、一覧を出さずにエラーにすること
func TestDelete_候補がなければエラー(t *testing.T) {
	cli := setupCLI(t)

	if err := cli.run("delete", "--branch"); err == nil {
		t.Error("エラーにならなかった")
	}

	if cli.branchShown {
		t.Error("一覧が表示された")
	}
}

// delete で対象のフラグがなければ、対象が branch だけなので選ばせずにブランチの選択に進むこと
func TestDelete_対象がなければbranchに進む(t *testing.T) {
	cli := setupCLI(t)
	cli.branches = deletable
	cli.branchPick = []int{0}

	if err := cli.run("delete", "--force"); err != nil {
		t.Fatal(err)
	}

	if cli.targetItems != nil {
		t.Error("対象を選ぶ画面が表示された")
	}
	if !cli.branchShown {
		t.Error("ブランチを選ぶ画面に進んでいない")
	}
	if want := []string{"feature/a!"}; !reflect.DeepEqual(cli.deleted, want) {
		t.Errorf("deleted %v, want %v", cli.deleted, want)
	}
}

// 削除できたブランチだけを、undo で作り直せるよう履歴に書くこと
func TestDelete_削除したブランチを履歴に書く(t *testing.T) {
	cli := setupCLI(t)
	cli.branches = deletable
	cli.branchPick = []int{0, 1}
	cli.deleteFail = map[string]bool{deletable[1].Name: true}

	_ = cli.run("delete", "--branch")

	r, err := history.Load(filepath.Join(cli.historyDir, "history.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := []branch.Saved{{Repo: "/repo", Name: deletable[0].Name, Commit: "c-" + deletable[0].Name}}
	if r.Kind != history.KindDeleteBranch || !reflect.DeepEqual(r.Branches, want) {
		t.Errorf("got %+v, want %+v", r, want)
	}
}

// 1件も削除しなければ、履歴を書かないこと
func TestDelete_何も削除しなければ履歴を書かない(t *testing.T) {
	cli := setupCLI(t)
	cli.branches = deletable

	if err := cli.run("delete", "--branch"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(cli.historyDir, "history.json")); !os.IsNotExist(err) {
		t.Errorf("履歴が書かれた: %v", err)
	}
}
