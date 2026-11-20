package aisession

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// テスト用ディレクトリで git を実行する
func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// 同じリポジトリの worktree は、別ディレクトリでも同じリポジトリになること
func TestRepoOf_worktreeは同じリポジトリ(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(base, "app")
	wt := filepath.Join(base, "app-wt")
	if err := os.Mkdir(main, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, main, "init", "-q")
	git(t, main, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "--allow-empty", "-m", "init")
	git(t, main, "worktree", "add", "-q", wt)
	sub := filepath.Join(main, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	r1, ok1 := RepoOf(main)
	r2, ok2 := RepoOf(wt)
	r3, ok3 := RepoOf(sub)

	if !ok1 || !ok2 || !ok3 {
		t.Fatalf("存在するのに ok が false: %v %v %v", ok1, ok2, ok3)
	}
	if r1 != main || r2 != main || r3 != main {
		t.Errorf("got %q %q %q, want すべて %q", r1, r2, r3, main)
	}
}

// git のリポジトリでなければ、そのディレクトリ自身をリポジトリとみなすこと
func TestRepoOf_リポジトリでなければ自身(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	got, ok := RepoOf(dir)

	if !ok || got != dir {
		t.Errorf("got %q, ok=%v", got, ok)
	}
}

// ディレクトリがもう存在しなければ、ok が false になること
func TestRepoOf_存在しなければfalse(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "deleted-worktree")

	got, ok := RepoOf(missing)

	if ok || got != missing {
		t.Errorf("got %q, ok=%v", got, ok)
	}
}
