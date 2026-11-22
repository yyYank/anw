package cmd

import (
	"errors"
	"reflect"
	"testing"

	"github.com/yyYank/anw/internal/port"
)

// anw kill --port 3000 で 3000番を使っているプロセスを kill すること
func TestKill_指定ポートのプロセスをkillする(t *testing.T) {
	cli := setupCLI(t)
	cli.listeners = []port.Listener{
		{Port: 3000, PID: 873, Name: "node"},
		{Port: 8080, PID: 900, Name: "java"},
	}

	if err := cli.run("kill", "--port", "3000"); err != nil {
		t.Fatal(err)
	}

	if want := []int{873}; !reflect.DeepEqual(cli.killed, want) {
		t.Errorf("killed %v, want %v", cli.killed, want)
	}
}

// 使っているプロセスがなければエラーにし、何も kill しないこと
func TestKill_該当なしはエラー(t *testing.T) {
	cli := setupCLI(t)
	cli.listeners = []port.Listener{{Port: 3000, PID: 873, Name: "node"}}

	if err := cli.run("kill", "--port", "9999"); err == nil {
		t.Error("エラーにならなかった")
	}

	if len(cli.killed) != 0 {
		t.Errorf("kill された: %v", cli.killed)
	}
}

// ポート番号が数字でなければエラーにし、何も kill しないこと
func TestKill_数字以外はエラー(t *testing.T) {
	cli := setupCLI(t)
	cli.listeners = []port.Listener{{Port: 3000, PID: 873, Name: "node"}}

	if err := cli.run("kill", "--port", "abc"); err == nil {
		t.Error("エラーにならなかった")
	}

	if len(cli.killed) != 0 {
		t.Errorf("kill された: %v", cli.killed)
	}
}

// 引数なしでは一覧から選んだプロセスを kill すること
func TestKill_一覧から選んでkillする(t *testing.T) {
	cli := setupCLI(t)
	cli.listeners = []port.Listener{
		{Port: 3000, PID: 873, Name: "node"},
		{Port: 8080, PID: 900, Name: "java"},
	}
	cli.pick = 1

	if err := cli.run("kill", "--port"); err != nil {
		t.Fatal(err)
	}

	if want := []int{900}; !reflect.DeepEqual(cli.killed, want) {
		t.Errorf("killed %v, want %v", cli.killed, want)
	}
}

// 一覧で選ばずにやめたら何も kill しないこと
func TestKill_一覧でやめたらkillしない(t *testing.T) {
	cli := setupCLI(t)
	cli.listeners = []port.Listener{{Port: 3000, PID: 873, Name: "node"}}
	cli.pick = -1

	// やめたことが呼び出し元に分かるよう、errCanceled を返す
	if err := cli.run("kill", "--port"); !errors.Is(err, errCanceled) {
		t.Errorf("err = %v, want errCanceled", err)
	}

	if len(cli.killed) != 0 {
		t.Errorf("kill された: %v", cli.killed)
	}
}

// LISTEN 中のポートがなければ一覧を出さずにエラーにすること
func TestKill_LISTENがなければエラー(t *testing.T) {
	cli := setupCLI(t)

	if err := cli.run("kill", "--port"); err == nil {
		t.Error("エラーにならなかった")
	}

	if cli.shown {
		t.Error("一覧が表示された")
	}
}

// kill で対象のフラグがなければエラーにし、何も kill しないこと
func TestKill_対象がなければエラー(t *testing.T) {
	cli := setupCLI(t)
	cli.listeners = []port.Listener{{Port: 3000, PID: 873, Name: "node"}}

	if err := cli.run("kill", "3000"); err == nil {
		t.Error("エラーにならなかった")
	}

	if len(cli.killed) != 0 {
		t.Errorf("kill された: %v", cli.killed)
	}
}

// 旧サブコマンドの env / port は無くなっていること
func TestKill_旧サブコマンドは使えない(t *testing.T) {
	for _, name := range []string{"env", "port"} {
		t.Run(name, func(t *testing.T) {
			cli := setupCLI(t)
			if err := cli.run(name); err == nil {
				t.Errorf("anw %s がエラーにならなかった", name)
			}
		})
	}
}
