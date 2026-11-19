package cmd

import (
	"reflect"
	"testing"

	"github.com/yyYank/anw/internal/branch"
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

// delete で対象のフラグがなければエラーにし、何も削除しないこと
func TestDelete_対象がなければエラー(t *testing.T) {
	cli := setupCLI(t)
	cli.branches = deletable
	cli.branchPick = []int{0}

	if err := cli.run("delete", "--force"); err == nil {
		t.Error("エラーにならなかった")
	}

	if len(cli.deleted) != 0 {
		t.Errorf("削除された: %v", cli.deleted)
	}
}
