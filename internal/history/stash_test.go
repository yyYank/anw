package history

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yyYank/anw/internal/stash"
)

// setupStashRepo は stash を2つ持つテスト用リポジトリを作ってカレントにする
func setupStashRepo(t *testing.T) string {
	t.Helper()
	repo := setupRepo(t)
	for _, name := range []string{"s1", "s2"} {
		if err := os.WriteFile(filepath.Join(repo, name+".txt"), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
		if out, err := gitRun(t, repo, "stash", "push", "-q", "-u", "-m", name); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
	}
	return repo
}

// 消した stash の記録を保存すると、undo で作り直し、履歴が消えること
func TestUndo_消したstashを作り直す(t *testing.T) {
	repo := setupStashRepo(t)
	path := filepath.Join(t.TempDir(), "history.json")
	all, err := stash.List()
	if err != nil {
		t.Fatal(err)
	}
	saved, err := stash.Drop(all)
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveStashes(path, saved); err != nil {
		t.Fatal(err)
	}

	r, err := Undo(path)
	if err != nil {
		t.Fatal(err)
	}

	if r.Kind != KindDropStash || len(r.Stashes) != 2 {
		t.Errorf("got %+v", r)
	}
	out, _ := gitRun(t, repo, "stash", "list", "--format=%gs")
	if out != "On main: s2\nOn main: s1" {
		t.Errorf("stash が作り直されていない:\n%s", out)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("履歴が残っている: %v", err)
	}
}

// 1件でも作り直せなければ、何も作り直さずにエラーにし、履歴を残すこと
func TestUndo_作り直せないstashがあれば何もしない(t *testing.T) {
	repo := setupStashRepo(t)
	path := filepath.Join(t.TempDir(), "history.json")
	all, err := stash.List()
	if err != nil {
		t.Fatal(err)
	}
	saved, err := stash.Drop(all)
	if err != nil {
		t.Fatal(err)
	}
	saved[0].Commit = strings.Repeat("0", len(saved[0].Commit))
	if err := SaveStashes(path, saved); err != nil {
		t.Fatal(err)
	}

	if _, err := Undo(path); err == nil {
		t.Error("エラーにならなかった")
	}

	if out, _ := gitRun(t, repo, "stash", "list"); out != "" {
		t.Errorf("作り直された:\n%s", out)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("履歴が消えた: %v", err)
	}
}
