package worktree

import (
	"reflect"
	"testing"

	"github.com/yyYank/anw/internal/branch"
)

const porcelain = `worktree /repo
HEAD 1111
branch refs/heads/main

worktree /repo-wt/current
HEAD 2222
branch refs/heads/feature/current

worktree /repo-wt/merged
HEAD 3333
branch refs/heads/feature/merged

worktree /repo-wt/gone
HEAD 4444
branch refs/heads/feature/gone

worktree /repo-wt/missing
HEAD 5555
branch refs/heads/feature/wip
prunable gitdir file points to non-existent location

worktree /repo-wt/active
HEAD 6666
branch refs/heads/feature/active

worktree /repo-wt/detached
HEAD 7777
detached

`

// 消してよい worktree と理由を選び、メインと今いる worktree は含めないこと
func TestCollect_消してよいworktreeを選ぶ(t *testing.T) {
	deletable := []branch.Branch{
		{Name: "feature/merged", Merged: true},
		{Name: "feature/gone", Gone: true},
		{Name: "feature/current", Merged: true},
	}

	got := Collect(porcelain, "/repo-wt/current", deletable)

	want := []Worktree{
		{Path: "/repo-wt/gone", Branch: "feature/gone", Gone: true},
		{Path: "/repo-wt/merged", Branch: "feature/merged", Merged: true},
		{Path: "/repo-wt/missing", Branch: "feature/wip", Missing: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
}

// 理由を表示用の文字列にすること
func TestReason_理由を並べる(t *testing.T) {
	w := Worktree{Missing: true, Merged: true}
	if got := w.Reason(); got != "ディレクトリなし, マージ済み" {
		t.Errorf("got %q", got)
	}
}
