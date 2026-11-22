package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/yyYank/anw/internal/branch"
	"github.com/yyYank/anw/internal/port"
)

// view --env で閲覧だけの一覧を表示し、何も開かないこと
func TestView_envは閲覧だけ(t *testing.T) {
	cli := setupCLI(t)
	cli.envFiles = []string{".env"}

	if err := cli.run("view", "--env"); err != nil {
		t.Fatal(err)
	}

	if cli.envPicker == nil || *cli.envPicker {
		t.Error("閲覧だけの一覧が表示されていない")
	}
	if len(cli.edited) != 0 {
		t.Errorf("開かれた: %v", cli.edited)
	}
}

// env ファイルがなければ、一覧を出さずにエラーにすること
func TestView_envが見つからなければエラー(t *testing.T) {
	cli := setupCLI(t)

	if err := cli.run("view", "--env"); err == nil {
		t.Error("エラーにならなかった")
	}

	if cli.envPicker != nil {
		t.Error("一覧が表示された")
	}
}

// view --port で LISTEN 中のポートを表で出力し、何も kill しないこと
func TestView_portは一覧を出力する(t *testing.T) {
	cli := setupCLI(t)
	cli.listeners = []port.Listener{
		{Port: 3000, PID: 873, Name: "node", Command: "node server.js"},
		{Port: 8080, PID: 900, Name: "java", Command: "java -jar app.jar"},
	}

	if err := cli.run("view", "--port"); err != nil {
		t.Fatal(err)
	}

	for _, s := range []string{"3000", "873", "node server.js", "8080", "900", "java -jar app.jar"} {
		if !strings.Contains(cli.out.String(), s) {
			t.Errorf("%q が出力されていない:\n%s", s, cli.out.String())
		}
	}
	if cli.shown || len(cli.killed) != 0 {
		t.Error("選択の画面が出たか kill された")
	}
}

// view --port 3000 で 3000番を使っているプロセスだけを出力すること
func TestView_port番号指定はそのプロセスだけ(t *testing.T) {
	cli := setupCLI(t)
	cli.listeners = []port.Listener{
		{Port: 3000, PID: 873, Name: "node"},
		{Port: 8080, PID: 900, Name: "java"},
	}

	if err := cli.run("view", "--port", "3000"); err != nil {
		t.Fatal(err)
	}

	if out := cli.out.String(); !strings.Contains(out, "873") || strings.Contains(out, "900") {
		t.Errorf("3000番のプロセスだけになっていない:\n%s", out)
	}
	if len(cli.killed) != 0 {
		t.Errorf("kill された: %v", cli.killed)
	}
}

// view --branch で削除できるブランチと理由を出力し、何も削除しないこと
func TestView_branchは一覧を出力する(t *testing.T) {
	cli := setupCLI(t)
	cli.branches = []branch.Branch{
		{Name: "feature/a", Merged: true},
		{Name: "feature/b", Gone: true},
	}

	if err := cli.run("view", "--branch"); err != nil {
		t.Fatal(err)
	}

	for _, s := range []string{"feature/a", "マージ済み", "feature/b", "リモート削除済み"} {
		if !strings.Contains(cli.out.String(), s) {
			t.Errorf("%q が出力されていない:\n%s", s, cli.out.String())
		}
	}
	if cli.branchShown || len(cli.deleted) != 0 {
		t.Error("選択の画面が出たか削除された")
	}
}

// view で対象のフラグがなければ、対象を選ぶ画面を出すこと
func TestView_対象がなければ選ばせる(t *testing.T) {
	cli := setupCLI(t)

	// 選ばずにやめたので errCanceled を返す
	if err := cli.run("view"); !errors.Is(err, errCanceled) {
		t.Errorf("err = %v, want errCanceled", err)
	}

	if cli.targetItems == nil {
		t.Error("対象を選ぶ画面が表示されていない")
	}
}

// view で対象のフラグを複数指定するとエラーにすること
func TestView_対象を複数指定するとエラー(t *testing.T) {
	cli := setupCLI(t)
	cli.envFiles = []string{".env"}

	if err := cli.run("view", "--env", "--port"); err == nil {
		t.Error("エラーにならなかった")
	}

	if cli.envPicker != nil {
		t.Error("一覧が表示された")
	}
}
