package history

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/yyYank/anw/internal/tidy"
)

// 保存した移動の記録をそのまま読み戻せること
func TestSaveLoad_保存した記録を読み戻せる(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".anw", "history.json")
	moves := []tidy.Move{
		{From: "/a/memo.md", To: "/a/2026-10-27/memo.md"},
		{From: "/a/report.pdf", To: "/a/2026-10-29/report.pdf"},
	}

	if err := Save(path, moves); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, moves) {
		t.Errorf("got %+v, want %+v", got, moves)
	}
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(filepath.Base(path)), 0o644); err != nil {
		t.Fatal(err)
	}
}

// 直前の操作で移動したファイルが元の場所に戻り、空になった日付ディレクトリが消えること
func TestUndo_元の場所に戻す(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".anw", "history.json")
	from := filepath.Join(dir, "memo.md")
	to := filepath.Join(dir, "2026-10-27", "memo.md")
	writeFile(t, to)
	if err := Save(path, []tidy.Move{{From: from, To: to}}); err != nil {
		t.Fatal(err)
	}

	if _, err := Undo(path); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(from); err != nil {
		t.Errorf("元の場所に戻っていない: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "2026-10-27")); !os.IsNotExist(err) {
		t.Errorf("空の日付ディレクトリが残っている: %v", err)
	}
}

// 元の場所に同名ファイルがある場合はエラーにし、上書きしないこと
func TestUndo_元の場所に同名ファイルがあれば上書きしない(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".anw", "history.json")
	from := filepath.Join(dir, "memo.md")
	to := filepath.Join(dir, "2026-10-27", "memo.md")
	writeFile(t, to)
	if err := os.WriteFile(from, []byte("新しいファイル"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, []tidy.Move{{From: from, To: to}}); err != nil {
		t.Fatal(err)
	}

	if _, err := Undo(path); err == nil {
		t.Error("エラーにならなかった")
	}

	data, err := os.ReadFile(from)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "新しいファイル" {
		t.Errorf("上書きされた: %q", data)
	}
	if _, err := os.Stat(to); err != nil {
		t.Errorf("移動先のファイルが消えた: %v", err)
	}
}

// 履歴がない場合はエラーになること
func TestUndo_履歴がなければエラー(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".anw", "history.json")

	if _, err := Undo(path); err == nil {
		t.Error("エラーにならなかった")
	}
}

// 元に戻したあとは履歴が消え、2回続けて元に戻せないこと
func TestUndo_2回目はエラー(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".anw", "history.json")
	from := filepath.Join(dir, "memo.md")
	to := filepath.Join(dir, "2026-10-27", "memo.md")
	writeFile(t, to)
	if err := Save(path, []tidy.Move{{From: from, To: to}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Undo(path); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("履歴が残っている: %v", err)
	}
	if _, err := Undo(path); err == nil {
		t.Error("2回目の undo がエラーにならなかった")
	}
}
