package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yyYank/anw/internal/history"
	"github.com/yyYank/anw/internal/port"
)

// setupSnippet はテスト用の toml に snippet "s" を足し、ステップの dir に使うディレクトリ a, b を作る
func setupSnippet(t *testing.T, steps ...string) (cli *testCLI, a, b string) {
	t.Helper()
	cli = setupCLI(t)
	a, b = filepath.Join(cli.workDir, "a"), filepath.Join(cli.workDir, "b")
	for _, d := range []string{a, b} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	f, err := os.OpenFile(cli.configPath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	for _, s := range steps {
		s = strings.NewReplacer("{a}", a, "{b}", b).Replace(s)
		if _, err := fmt.Fprintf(f, "\n[[snippet.s]]\n%s\n", s); err != nil {
			t.Fatal(err)
		}
	}
	return cli, a, b
}

// createFileIn は dir にテスト対象のファイルを作る
func createFileIn(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(name), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, oct27, oct27); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertMoved(t *testing.T, dir, name string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(dir, "2026-10-27", name)); err != nil {
		t.Errorf("%s が振り分けられていない: %v", filepath.Join(dir, name), err)
	}
}

func assertStays(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("%s が元の場所にない: %v", path, err)
	}
}

// 各ステップの dir で、anw のコマンドが順に実行されること
func TestRun_ステップを順に実行する(t *testing.T) {
	cli, a, b := setupSnippet(t,
		`dir = "{a}"
args = ["tidy"]`,
		`dir = "{b}"
args = ["tidy", "--by", "date"]`)
	createFileIn(t, a, "memo.md")
	createFileIn(t, b, "report.pdf")

	if err := cli.run("run", "--snippet", "s"); err != nil {
		t.Fatal(err)
	}

	assertMoved(t, a, "memo.md")
	assertMoved(t, b, "report.pdf")
	if wd, _ := os.Getwd(); wd != cli.workDir {
		t.Errorf("カレントディレクトリが戻っていない: %s", wd)
	}
}

// snippet で行った移動を、anw undo でまとめて戻せること
func TestRun_undoでまとめて戻す(t *testing.T) {
	cli, a, b := setupSnippet(t,
		`dir = "{a}"
args = ["tidy"]`,
		`dir = "{b}"
args = ["tidy"]`)
	memo := createFileIn(t, a, "memo.md")
	report := createFileIn(t, b, "report.pdf")

	if err := cli.run("run", "--snippet", "s"); err != nil {
		t.Fatal(err)
	}
	if err := cli.run("undo"); err != nil {
		t.Fatal(err)
	}

	assertStays(t, memo)
	assertStays(t, report)
}

// どれかのステップが正しくなければ、どのステップも実行しないこと
func TestRun_検証で誤りがあれば何もしない(t *testing.T) {
	cases := map[string]string{
		"dir がない":     `dir = "{a}/none"` + "\n" + `args = ["tidy"]`,
		"コマンドがない":     `dir = "{b}"` + "\n" + `args = ["nope"]`,
		"フラグが正しくない":   `dir = "{b}"` + "\n" + `args = ["tidy", "--nope"]`,
		"対象のフラグを複数指定": `dir = "{b}"` + "\n" + `args = ["view", "--env", "--port"]`,
		"run を呼んでいる":  `dir = "{b}"` + "\n" + `args = ["run", "--snippet", "s"]`,
	}
	for name, bad := range cases {
		t.Run(name, func(t *testing.T) {
			cli, a, _ := setupSnippet(t, `dir = "{a}"`+"\n"+`args = ["tidy"]`, bad)
			memo := createFileIn(t, a, "memo.md")

			if err := cli.run("run", "--snippet", "s"); err == nil {
				t.Error("エラーにならなかった")
			}

			assertStays(t, memo)
		})
	}
}

// 途中のステップが失敗したらそこで止め、それまでの移動は undo で戻せること
func TestRun_失敗したら止める(t *testing.T) {
	cli, a, b := setupSnippet(t,
		`dir = "{a}"
args = ["tidy"]`,
		`dir = "{a}"
args = ["view", "--port", "9999"]`,
		`dir = "{b}"
args = ["tidy"]`)
	memo := createFileIn(t, a, "memo.md")
	report := createFileIn(t, b, "report.pdf")

	if err := cli.run("run", "--snippet", "s"); err == nil {
		t.Error("エラーにならなかった")
	}
	assertMoved(t, a, "memo.md")
	assertStays(t, report)

	if err := cli.run("undo"); err != nil {
		t.Fatal(err)
	}
	assertStays(t, memo)
}

// 一覧の画面でやめたら、そこで snippet 全体を止めること
func TestRun_やめたら止める(t *testing.T) {
	cli, _, b := setupSnippet(t,
		`dir = "{a}"
args = ["kill", "--port"]`,
		`dir = "{b}"
args = ["tidy"]`)
	cli.listeners = []port.Listener{{Port: 3000, PID: 873, Name: "node"}}
	cli.pick = -1
	report := createFileIn(t, b, "report.pdf")

	if err := cli.run("run", "--snippet", "s"); !errors.Is(err, errCanceled) {
		t.Errorf("err = %v, want errCanceled", err)
	}

	assertStays(t, report)
}

// --snippet がない、または snippet が見つからなければエラーにすること
func TestRun_snippetの指定がなければエラー(t *testing.T) {
	cli, _, _ := setupSnippet(t, `dir = "{a}"`+"\n"+`args = ["tidy"]`)

	if err := cli.run("run"); err == nil {
		t.Error("--snippet なしがエラーにならなかった")
	}
	if err := cli.run("run", "--snippet", "none"); err == nil {
		t.Error("存在しない snippet がエラーにならなかった")
	}
}

// snippet で行った種類の違う操作を、snippet の1件の履歴として保存すること
func TestRun_種類の違う操作をまとめて履歴に書く(t *testing.T) {
	cli, a, _ := setupSnippet(t,
		`dir = "{a}"
args = ["tidy"]`,
		`dir = "{a}"
args = ["delete", "--branch"]`)
	cli.branches = deletable
	cli.branchPick = []int{0}
	createFileIn(t, a, "memo.md")

	if err := cli.run("run", "--snippet", "s"); err != nil {
		t.Fatal(err)
	}

	r, err := history.Load(filepath.Join(cli.historyDir, "history.json"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Kind != history.KindSnippet || len(r.Steps) != 2 ||
		r.Steps[0].Kind != history.KindTidy || r.Steps[1].Kind != history.KindDeleteBranch {
		t.Errorf("snippet の履歴になっていない: %+v", r)
	}
}
