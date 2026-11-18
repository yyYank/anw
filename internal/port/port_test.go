package port

import (
	"reflect"
	"testing"
)

// lsof -F pcn の出力から、ポート・PID・プロセス名を取り出せること
func TestParse_lsofの出力を読み取る(t *testing.T) {
	out := "p370\ncrapportd\nf3\nn*:49162\n" +
		"p873\ncnode\nf12\nn127.0.0.1:3000\n"

	got := Parse(out)

	want := []Listener{
		{Port: 3000, PID: 873, Name: "node"},
		{Port: 49162, PID: 370, Name: "rapportd"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// IPv4 と IPv6 で同じポートが重複して出ても1件にまとめること
func TestParse_重複を除く(t *testing.T) {
	out := "p873\ncnode\nf12\nn127.0.0.1:3000\nf13\nn[::1]:3000\n"

	got := Parse(out)

	want := []Listener{{Port: 3000, PID: 873, Name: "node"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// ポート番号で使っているプロセスを探せること
func TestFind_ポート番号で探す(t *testing.T) {
	ls := []Listener{
		{Port: 3000, PID: 873, Name: "node"},
		{Port: 8080, PID: 900, Name: "java"},
	}

	got := Find(ls, 8080)

	want := []Listener{{Port: 8080, PID: 900, Name: "java"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// 該当がなければ空になること
func TestFind_該当なし(t *testing.T) {
	if got := Find([]Listener{{Port: 3000, PID: 873}}, 9999); len(got) != 0 {
		t.Errorf("got %+v", got)
	}
}
