package history

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/yyYank/anw/internal/tidy"
)

// 保存した移動の記録を、tidy の履歴として読み戻せること
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

	if got.Kind != KindTidy || !reflect.DeepEqual(got.Moves, moves) {
		t.Errorf("got %+v, want tidy %+v", got, moves)
	}
}

// 保存したファイルには操作の種類が書かれていること
func TestSave_操作の種類を書く(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")

	if err := Save(path, []tidy.Move{{From: "/a/x", To: "/a/d/x"}}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"kind":"tidy"`) {
		t.Errorf("種類が書かれていない: %s", data)
	}
}

// 移動の配列だけの古い形式の履歴も、tidy の履歴として読めること
func TestLoad_古い形式はtidyとして読む(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	old := `[{"From":"/a/memo.md","To":"/a/2026-10-27/memo.md"}]`
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	want := []tidy.Move{{From: "/a/memo.md", To: "/a/2026-10-27/memo.md"}}
	if got.Kind != KindTidy || !reflect.DeepEqual(got.Moves, want) {
		t.Errorf("got %+v", got)
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

// 入れ子の日付ディレクトリから戻したとき、空になったディレクトリを上の階層まで消すこと
func TestUndo_空になった入れ子のディレクトリを消す(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".anw", "history.json")
	from := filepath.Join(dir, "memo.md")
	to := filepath.Join(dir, "2026-10", "2026-10-27", "memo.md")
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
	if _, err := os.Stat(filepath.Join(dir, "2026-10")); !os.IsNotExist(err) {
		t.Errorf("空の月ディレクトリが残っている: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("元の場所のディレクトリまで消えた: %v", err)
	}
}

// 移動元のディレクトリが無くなっていても、作り直して元の場所に戻すこと
func TestUndo_移動元のディレクトリを作り直す(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".anw", "history.json")
	from := filepath.Join(dir, "2026-10-27", "memo.md")
	to := filepath.Join(dir, "2026-10", "2026-10-27", "memo.md")
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
	if _, err := os.Stat(filepath.Join(dir, "2026-10")); !os.IsNotExist(err) {
		t.Errorf("空の月ディレクトリが残っている: %v", err)
	}
}
