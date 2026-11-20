package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/yyYank/anw/internal/history"
	"github.com/yyYank/anw/internal/stash"
	"github.com/yyYank/anw/internal/worktree"
)

var staleWorktrees = []worktree.Worktree{
	{Path: "/repo-wt/a", Branch: "feature/a", Merged: true},
	{Path: "/repo-wt/b", Branch: "feature/b", Missing: true},
}

var sampleStashes = []stash.Stash{
	{Ref: "stash@{0}", Commit: "aaaa", Message: "On main: 実験", Branch: "main", Time: time.Now().Add(-48 * time.Hour)},
	{Ref: "stash@{1}", Commit: "bbbb", Message: "WIP on feature/x: 作業中", Branch: "feature/x", BranchGone: true, Time: time.Now().Add(-240 * time.Hour)},
}

// view --worktree で、消してよい worktree をパス・ブランチ・理由の表で出すこと
func TestView_worktreeを表で出す(t *testing.T) {
	cli := setupCLI(t)
	cli.worktrees = staleWorktrees

	if err := cli.run("view", "--worktree"); err != nil {
		t.Fatal(err)
	}

	out := cli.out.String()
	for _, s := range []string{"PATH", "/repo-wt/a", "feature/a", "マージ済み", "/repo-wt/b", "ディレクトリなし"} {
		if !strings.Contains(out, s) {
			t.Errorf("%q がない:\n%s", s, out)
		}
	}
	if cli.multiTitle != "" || len(cli.removed) != 0 {
		t.Error("選択の画面が出たか削除された")
	}
}

// delete --worktree で選んだ worktree を削除し、変更が残って消せないものは --force を案内すること
func TestDelete_worktreeを選んで消す(t *testing.T) {
	cli := setupCLI(t)
	cli.worktrees = staleWorktrees
	cli.multiPick = []int{0, 1}
	cli.removeFail = map[string]bool{"/repo-wt/a": true}

	err := cli.run("delete", "--worktree")

	if err == nil || !strings.Contains(err.Error(), "--force") {
		t.Errorf("err = %v", err)
	}
	if want := []string{"/repo-wt/b"}; !reflect.DeepEqual(cli.removed, want) {
		t.Errorf("removed %v, want %v", cli.removed, want)
	}
}

// delete --force --worktree で強制削除すること
func TestDelete_worktreeを強制削除する(t *testing.T) {
	cli := setupCLI(t)
	cli.worktrees = staleWorktrees
	cli.multiPick = []int{0}

	if err := cli.run("delete", "--force", "--worktree"); err != nil {
		t.Fatal(err)
	}

	if want := []string{"/repo-wt/a!"}; !reflect.DeepEqual(cli.removed, want) {
		t.Errorf("removed %v, want %v", cli.removed, want)
	}
}

// view --stash で、すべての stash を番号・メッセージ・元のブランチと日数の表で出すこと
func TestView_stashを表で出す(t *testing.T) {
	cli := setupCLI(t)
	cli.stashes = sampleStashes

	if err := cli.run("view", "--stash"); err != nil {
		t.Fatal(err)
	}

	out := cli.out.String()
	for _, s := range []string{"STASH", "stash@{0}", "On main: 実験", "main, 2日前", "feature/x (ブランチなし), 10日前"} {
		if !strings.Contains(out, s) {
			t.Errorf("%q がない:\n%s", s, out)
		}
	}
}

// delete --stash で選んだ stash を消し、undo で作り直せるよう履歴に書くこと
func TestDelete_stashを選んで消して履歴に書く(t *testing.T) {
	cli := setupCLI(t)
	cli.stashes = sampleStashes
	cli.multiPick = []int{1}

	if err := cli.run("delete", "--stash"); err != nil {
		t.Fatal(err)
	}

	if want := []string{"WIP on feature/x: 作業中"}; !reflect.DeepEqual(cli.dropped, want) {
		t.Errorf("dropped %v, want %v", cli.dropped, want)
	}
	r, err := history.Load(filepath.Join(cli.historyDir, "history.json"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Kind != history.KindDropStash || len(r.Stashes) != 1 || r.Stashes[0].Commit != "bbbb" {
		t.Errorf("got %+v", r)
	}
}

// 選ばずにやめたら何も消さず、履歴も書かないこと
func TestDelete_stashを選ばずにやめたら何もしない(t *testing.T) {
	cli := setupCLI(t)
	cli.stashes = sampleStashes

	if err := cli.run("delete", "--stash"); err != nil {
		t.Fatal(err)
	}

	if len(cli.dropped) != 0 {
		t.Errorf("消された: %v", cli.dropped)
	}
	if _, err := os.Stat(filepath.Join(cli.historyDir, "history.json")); !os.IsNotExist(err) {
		t.Errorf("履歴が書かれた: %v", err)
	}
}

// 対象がなければ、選択の画面を出さずにエラーにすること
func TestDelete_worktreeとstashがなければエラー(t *testing.T) {
	for _, target := range []string{"--worktree", "--stash"} {
		t.Run(target, func(t *testing.T) {
			cli := setupCLI(t)

			if err := cli.run("delete", target); err == nil {
				t.Error("エラーにならなかった")
			}
			if cli.multiTitle != "" {
				t.Error("選択の画面が出た")
			}
		})
	}
}
