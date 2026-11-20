package history

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yyYank/anw/internal/branch"
)

// gitRun はテスト用リポジトリで git を実行して出力を返す
func gitRun(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// setupRepo は main と同じコミットの feature/a と feature/b を持つテスト用リポジトリを作り、カレントにする
func setupRepo(t *testing.T) string {
	t.Helper()
	// 利用者の git 設定を読まない
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	t.Setenv("GIT_AUTHOR_NAME", "t")
	t.Setenv("GIT_AUTHOR_EMAIL", "t@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "t")
	t.Setenv("GIT_COMMITTER_EMAIL", "t@example.com")
	repo := filepath.Join(t.TempDir(), "repo")
	for _, args := range [][]string{
		{"init", "-q", "-b", "main", repo},
		{"-C", repo, "commit", "-q", "--allow-empty", "-m", "init"},
		{"-C", repo, "branch", "feature/a"},
		{"-C", repo, "branch", "feature/b"},
	} {
		if out, err := gitRun(t, ".", args...); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	t.Chdir(repo)
	return repo
}

// snapshotAndDelete はブランチを記録してから削除する。main と同じコミットなので強制せずに消せる
func snapshotAndDelete(t *testing.T, repo string, names ...string) []branch.Saved {
	t.Helper()
	var saved []branch.Saved
	for _, name := range names {
		s, err := branch.Snapshot(name)
		if err != nil {
			t.Fatal(err)
		}
		saved = append(saved, s)
		if out, err := gitRun(t, repo, "branch", "-q", "--delete", name); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
	}
	return saved
}

// 削除したブランチの記録を保存すると、undo でまとめて作り直し、履歴が消えること
func TestUndo_削除したブランチを作り直す(t *testing.T) {
	repo := setupRepo(t)
	path := filepath.Join(t.TempDir(), "history.json")
	saved := snapshotAndDelete(t, repo, "feature/a", "feature/b")
	if err := SaveBranches(path, saved); err != nil {
		t.Fatal(err)
	}

	r, err := Undo(path)
	if err != nil {
		t.Fatal(err)
	}

	if r.Kind != KindDeleteBranch || len(r.Branches) != 2 {
		t.Errorf("got %+v", r)
	}
	for _, s := range saved {
		if got, err := gitRun(t, repo, "rev-parse", s.Name); err != nil || got != s.Commit {
			t.Errorf("%s が作り直されていない: %q %v", s.Name, got, err)
		}
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("履歴が残っている: %v", err)
	}
}

// 1件でも作り直せなければ、何も作り直さずにエラーにし、履歴を残すこと
func TestUndo_作り直せないものがあれば何もしない(t *testing.T) {
	repo := setupRepo(t)
	path := filepath.Join(t.TempDir(), "history.json")
	saved := snapshotAndDelete(t, repo, "feature/a", "feature/b")
	if err := SaveBranches(path, saved); err != nil {
		t.Fatal(err)
	}
	// 削除後に同じ名前の feature/b が作られた
	if out, err := gitRun(t, repo, "branch", "feature/b"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}

	if _, err := Undo(path); err == nil {
		t.Error("エラーにならなかった")
	}

	if _, err := gitRun(t, repo, "rev-parse", "--verify", "refs/heads/feature/a"); err == nil {
		t.Error("feature/a が作り直された")
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("履歴が消えた: %v", err)
	}
}
