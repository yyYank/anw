package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// setupRepo は origin/HEAD が main を指すテスト用リポジトリに、3つの worktree を作ってカレントにする。
// merged はマージ済みのブランチ、active は未マージのブランチ、missing はディレクトリを移動して消したもの
func setupRepo(t *testing.T) (repo, base string) {
	t.Helper()
	// 利用者の git 設定を読まない
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	t.Setenv("GIT_AUTHOR_NAME", "t")
	t.Setenv("GIT_AUTHOR_EMAIL", "t@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "t")
	t.Setenv("GIT_COMMITTER_EMAIL", "t@example.com")

	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	remote := filepath.Join(base, "remote.git")
	repo = filepath.Join(base, "repo")
	run(t, base, "init", "-q", "--bare", remote)
	run(t, base, "init", "-q", "-b", "main", repo)
	run(t, repo, "commit", "-q", "--allow-empty", "-m", "init")
	run(t, repo, "remote", "add", "origin", remote)
	run(t, repo, "push", "-q", "origin", "main")
	run(t, repo, "remote", "set-head", "origin", "main")

	run(t, repo, "worktree", "add", "-q", "-b", "feature/merged", filepath.Join(base, "merged"))
	run(t, repo, "worktree", "add", "-q", "-b", "feature/active", filepath.Join(base, "active"))
	run(t, filepath.Join(base, "active"), "commit", "-q", "--allow-empty", "-m", "wip")
	run(t, repo, "worktree", "add", "-q", "-b", "feature/missing", filepath.Join(base, "missing"))
	run(t, filepath.Join(base, "missing"), "commit", "-q", "--allow-empty", "-m", "wip")
	if err := os.Rename(filepath.Join(base, "missing"), filepath.Join(base, "moved-away")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)
	return repo, base
}

// カレントのリポジトリで、マージ済みのものとディレクトリが消えたものを一覧すること
func TestList_消してよいworktreeを一覧する(t *testing.T) {
	_, base := setupRepo(t)

	got, err := List()
	if err != nil {
		t.Fatal(err)
	}

	var paths []string
	for _, w := range got {
		paths = append(paths, w.Path+"="+w.Reason())
	}
	want := filepath.Join(base, "merged") + "=マージ済み," + filepath.Join(base, "missing") + "=ディレクトリなし"
	if strings.Join(paths, ",") != want {
		t.Errorf("got %v, want %s", paths, want)
	}
}

// 変更のない worktree は削除でき、ディレクトリも消えること
func TestRemove_変更がなければ消す(t *testing.T) {
	_, base := setupRepo(t)
	path := filepath.Join(base, "merged")

	if err := Remove(path, false); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("ディレクトリが残っている: %v", err)
	}
}

// 変更が残っている worktree は、force でなければ消さないこと
func TestRemove_変更が残っていればforceでなければ消さない(t *testing.T) {
	_, base := setupRepo(t)
	path := filepath.Join(base, "merged")
	if err := os.WriteFile(filepath.Join(path, "draft.txt"), []byte("wip"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Remove(path, false); err == nil {
		t.Error("エラーにならなかった")
	}
	if _, err := os.Stat(filepath.Join(path, "draft.txt")); err != nil {
		t.Errorf("変更が消えた: %v", err)
	}

	if err := Remove(path, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("force でも消えていない: %v", err)
	}
}

// ディレクトリが消えた worktree は、git の管理情報から消すこと
func TestRemove_ディレクトリが消えたものは管理情報を消す(t *testing.T) {
	repo, base := setupRepo(t)
	path := filepath.Join(base, "missing")

	if err := Remove(path, false); err != nil {
		t.Fatal(err)
	}

	if out := run(t, repo, "worktree", "list", "--porcelain"); strings.Contains(out, path+"\n") {
		t.Errorf("管理情報が残っている:\n%s", out)
	}
}
