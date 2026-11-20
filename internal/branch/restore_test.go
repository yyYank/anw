package branch

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// run はテスト用リポジトリで git を実行して出力を返す
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

// setupRepo は origin に push 済みの feature/a を持つテスト用リポジトリを作り、カレントにする
func setupRepo(t *testing.T) (repo, commit string) {
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
	run(t, repo, "checkout", "-q", "-b", "feature/a")
	run(t, repo, "commit", "-q", "--allow-empty", "-m", "a")
	run(t, repo, "push", "-q", "-u", "origin", "feature/a")
	run(t, repo, "checkout", "-q", "main")
	t.Chdir(repo)
	return repo, run(t, repo, "rev-parse", "feature/a")
}

// 削除前のブランチの、リポジトリ・名前・コミット・上流ブランチを記録すること
func TestSnapshot_ブランチを記録する(t *testing.T) {
	repo, commit := setupRepo(t)

	got, err := Snapshot("feature/a")
	if err != nil {
		t.Fatal(err)
	}

	want := Saved{Repo: repo, Name: "feature/a", Commit: commit, Upstream: "origin/feature/a"}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// 削除したブランチを、同じ名前・同じコミット・同じ上流ブランチで作り直すこと
func TestRestore_同じコミットと上流で作り直す(t *testing.T) {
	repo, commit := setupRepo(t)
	saved, err := Snapshot("feature/a")
	if err != nil {
		t.Fatal(err)
	}
	run(t, repo, "branch", "-q", "-D", "feature/a")

	if err := CheckRestore(saved); err != nil {
		t.Fatal(err)
	}
	if err := Restore(saved); err != nil {
		t.Fatal(err)
	}

	if got := run(t, repo, "rev-parse", "feature/a"); got != commit {
		t.Errorf("commit %s, want %s", got, commit)
	}
	if got := run(t, repo, "rev-parse", "--abbrev-ref", "feature/a@{upstream}"); got != "origin/feature/a" {
		t.Errorf("upstream %s", got)
	}
}

// 上流ブランチがもうなければ、上流は設定せずにブランチだけ作り直すこと
func TestRestore_上流がなければブランチだけ戻す(t *testing.T) {
	repo, commit := setupRepo(t)
	saved, err := Snapshot("feature/a")
	if err != nil {
		t.Fatal(err)
	}
	run(t, repo, "branch", "-q", "-D", "feature/a")
	run(t, repo, "update-ref", "-d", "refs/remotes/origin/feature/a")

	if err := Restore(saved); err != nil {
		t.Fatal(err)
	}

	if got := run(t, repo, "rev-parse", "feature/a"); got != commit {
		t.Errorf("commit %s, want %s", got, commit)
	}
}

// 同じ名前のブランチがある、またはコミットがもうなければ、復元できないとすること
func TestCheckRestore_復元できない状況(t *testing.T) {
	repo, _ := setupRepo(t)
	saved, err := Snapshot("feature/a")
	if err != nil {
		t.Fatal(err)
	}

	// まだ消していないので、同じ名前のブランチがある
	if err := CheckRestore(saved); err == nil {
		t.Error("同じ名前のブランチがあるのにエラーにならなかった")
	}

	run(t, repo, "branch", "-q", "-D", "feature/a")
	gone := saved
	gone.Commit = strings.Repeat("0", len(saved.Commit))
	if err := CheckRestore(gone); err == nil {
		t.Error("コミットがないのにエラーにならなかった")
	}
}
