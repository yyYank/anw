package stash

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

// setupRepo は stash を3つ持つテスト用リポジトリを作ってカレントにする。stash@{0} が s3、stash@{2} が s1
func setupRepo(t *testing.T) string {
	t.Helper()
	// 利用者の git 設定を読まない
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	t.Setenv("GIT_AUTHOR_NAME", "t")
	t.Setenv("GIT_AUTHOR_EMAIL", "t@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "t")
	t.Setenv("GIT_COMMITTER_EMAIL", "t@example.com")

	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	run(t, repo, "init", "-q", "-b", "main")
	run(t, repo, "commit", "-q", "--allow-empty", "-m", "init")
	for _, name := range []string{"s1", "s2", "s3"} {
		if err := os.WriteFile(filepath.Join(repo, name+".txt"), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
		run(t, repo, "stash", "push", "-q", "-u", "-m", name)
	}
	t.Chdir(repo)
	return repo
}

func messages(t *testing.T, ss []Stash) string {
	t.Helper()
	var ms []string
	for _, s := range ss {
		ms = append(ms, s.Message)
	}
	return strings.Join(ms, ",")
}

// カレントのリポジトリの stash をすべて一覧すること
func TestList_すべて一覧する(t *testing.T) {
	setupRepo(t)

	got, err := List()
	if err != nil {
		t.Fatal(err)
	}

	if want := "On main: s3,On main: s2,On main: s1"; messages(t, got) != want {
		t.Errorf("got %s, want %s", messages(t, got), want)
	}
	if got[0].BranchGone {
		t.Error("main があるのにブランチなしになっている")
	}
}

// 複数の stash を、番号がずれても正しく消し、作り直すための記録を返すこと
func TestDrop_選んだものだけ消す(t *testing.T) {
	setupRepo(t)
	all, err := List()
	if err != nil {
		t.Fatal(err)
	}

	saved, err := Drop([]Stash{all[0], all[2]})
	if err != nil {
		t.Fatal(err)
	}

	rest, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if want := "On main: s2"; messages(t, rest) != want {
		t.Errorf("残り %s, want %s", messages(t, rest), want)
	}
	if len(saved) != 2 {
		t.Errorf("記録 %+v", saved)
	}
}

// 消した stash を同じコミットとメッセージで作り直すこと。
// git stash store は一番上に積むので、作り直したものは元の前後関係のまま一番上に並ぶ
func TestRestore_消したstashを作り直す(t *testing.T) {
	setupRepo(t)
	all, err := List()
	if err != nil {
		t.Fatal(err)
	}
	saved, err := Drop([]Stash{all[0], all[2]})
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range saved {
		if err := CheckRestore(s); err != nil {
			t.Fatal(err)
		}
	}
	if err := Restore(saved); err != nil {
		t.Fatal(err)
	}

	got, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if want := "On main: s3,On main: s1,On main: s2"; messages(t, got) != want {
		t.Errorf("got %s, want %s", messages(t, got), want)
	}
	if got[0].Commit != all[0].Commit || got[1].Commit != all[2].Commit {
		t.Error("コミットが元と違う")
	}
}
