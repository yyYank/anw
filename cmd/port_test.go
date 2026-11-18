package cmd

import (
	"reflect"
	"testing"

	"github.com/yyYank/anw/internal/port"
)

// anw port 3000 で 3000番を使っているプロセスを kill すること
func TestPort_指定ポートのプロセスをkillする(t *testing.T) {
	cli := setupCLI(t)
	cli.listeners = []port.Listener{
		{Port: 3000, PID: 873, Name: "node"},
		{Port: 8080, PID: 900, Name: "java"},
	}

	if err := cli.run("port", "3000"); err != nil {
		t.Fatal(err)
	}

	if want := []int{873}; !reflect.DeepEqual(cli.killed, want) {
		t.Errorf("killed %v, want %v", cli.killed, want)
	}
}

// 使っているプロセスがなければエラーにし、何も kill しないこと
func TestPort_該当なしはエラー(t *testing.T) {
	cli := setupCLI(t)
	cli.listeners = []port.Listener{{Port: 3000, PID: 873, Name: "node"}}

	if err := cli.run("port", "9999"); err == nil {
		t.Error("エラーにならなかった")
	}

	if len(cli.killed) != 0 {
		t.Errorf("kill された: %v", cli.killed)
	}
}

// ポート番号が数字でなければエラーにし、何も kill しないこと
func TestPort_数字以外はエラー(t *testing.T) {
	cli := setupCLI(t)
	cli.listeners = []port.Listener{{Port: 3000, PID: 873, Name: "node"}}

	if err := cli.run("port", "abc"); err == nil {
		t.Error("エラーにならなかった")
	}

	if len(cli.killed) != 0 {
		t.Errorf("kill された: %v", cli.killed)
	}
}

// 引数なしでは一覧から選んだプロセスを kill すること
func TestPort_一覧から選んでkillする(t *testing.T) {
	cli := setupCLI(t)
	cli.listeners = []port.Listener{
		{Port: 3000, PID: 873, Name: "node"},
		{Port: 8080, PID: 900, Name: "java"},
	}
	cli.pick = 1

	if err := cli.run("port"); err != nil {
		t.Fatal(err)
	}

	if want := []int{900}; !reflect.DeepEqual(cli.killed, want) {
		t.Errorf("killed %v, want %v", cli.killed, want)
	}
}

// 一覧で選ばずにやめたら何も kill しないこと
func TestPort_一覧でやめたらkillしない(t *testing.T) {
	cli := setupCLI(t)
	cli.listeners = []port.Listener{{Port: 3000, PID: 873, Name: "node"}}
	cli.pick = -1

	if err := cli.run("port"); err != nil {
		t.Fatal(err)
	}

	if len(cli.killed) != 0 {
		t.Errorf("kill された: %v", cli.killed)
	}
}

// LISTEN 中のポートがなければ一覧を出さずにエラーにすること
func TestPort_LISTENがなければエラー(t *testing.T) {
	cli := setupCLI(t)

	if err := cli.run("port"); err == nil {
		t.Error("エラーにならなかった")
	}

	if cli.shown {
		t.Error("一覧が表示された")
	}
}
